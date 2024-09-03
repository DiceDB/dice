package core

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/cockroachdb/swiss"
	"github.com/dicedb/dice/internal/constants"

	"github.com/charmbracelet/log"
)

type queryCache *swiss.Map[string, *Obj]

func newQueryCache() queryCache {
	return swiss.New[string, *Obj](0)
}

// WatchEvent Event to notify clients about changes in query results due to key updates
type WatchEvent struct {
	Key       string
	Operation string
	Value     *Obj
}

// WatchSubscription Event to watch/unwatch a query
type WatchSubscription struct {
	subscribe bool      // true for subscribe, false for unsubscribe
	query     DSQLQuery // query to watch
	clientFd  int       // client file descriptor
}

// WatchChan Channel to receive updates about keys that are being watched
var WatchChan chan WatchEvent

// WatchSubscriptionChan Channel to receive updates about query subscriptions
var WatchSubscriptionChan chan WatchSubscription

// QueryWatcher watches for changes in keys and notifies clients
type QueryWatcher struct {
	WatchList    sync.Map
	queryCaches  *swiss.Map[string, queryCache]
	queryCacheMu sync.RWMutex
}

// NewQueryWatcher initializes a new QueryWatcher
func NewQueryWatcher() *QueryWatcher {
	return &QueryWatcher{
		WatchList:   sync.Map{},                       // WatchList is a map of queries to their respective clients, type: map[DSQLQuery]*sync.Map[int]struct{}
		queryCaches: swiss.New[string, queryCache](0), // queryCaches is a map of queries to their respective data caches
	}
}

func (w *QueryWatcher) Run(ctx context.Context) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		w.listenForSubscriptions(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		w.watchKeys(ctx)
	}()

	<-ctx.Done()
	wg.Wait()
}

// listenForSubscriptions listens for query subscriptions and unsubscribes
func (w *QueryWatcher) listenForSubscriptions(ctx context.Context) {
	for {
		select {
		case event := <-WatchSubscriptionChan:
			if event.subscribe {
				w.AddWatcher(event.query, event.clientFd)
			} else {
				w.RemoveWatcher(event.query, event.clientFd)
			}
		case <-ctx.Done():
			return
		}
	}
}

// watchKeys watches for changes in keys and notifies clients
func (w *QueryWatcher) watchKeys(ctx context.Context) {
	for {
		select {
		case event := <-WatchChan:
			w.WatchList.Range(func(key, value interface{}) bool {
				query := key.(DSQLQuery)
				clients := value.(*sync.Map)

				if !WildCardMatch(query.KeyRegex, event.Key) {
					return true
				}

				// Add this key to the query store
				// TODO: We can implement more finegrained locking here which locks only the particular query's store.
				w.queryCacheMu.Lock()
				store, ok := w.queryCaches.Get(query.String())
				if !ok {
					log.Warnf("Query not found in queryCaches: %s", query)
					w.queryCacheMu.Unlock()
					return true
				}
				switch event.Operation {
				case constants.Set:
					((*swiss.Map[string, *Obj])(store)).Put(event.Key, event.Value)
				case constants.Del:
					((*swiss.Map[string, *Obj])(store)).Delete(event.Key)
				default:
					log.Warnf("Unknown operation: %s", event.Operation)
				}
				w.queryCacheMu.Unlock()

				// Execute the query
				w.queryCacheMu.RLock()
				queryResult, err := ExecuteQuery(&query, store)
				w.queryCacheMu.RUnlock()
				if err != nil {
					log.Error(err)
					return true
				}

				encodedResult := Encode(CreatePushResponse(&query, &queryResult), false)
				clients.Range(func(clientKey, _ interface{}) bool {
					clientFd := clientKey.(int)
					_, err := syscall.Write(clientFd, encodedResult)
					if err != nil {
						w.RemoveWatcher(query, clientFd)
					}
					return true
				})

				return true
			})
		case <-ctx.Done():
			return
		}
	}
}

// AddWatcher adds a client as a watcher to a query.
func (w *QueryWatcher) AddWatcher(query DSQLQuery, clientFd int) { //nolint:gocritic
	clients, _ := w.WatchList.LoadOrStore(query, &sync.Map{})
	clients.(*sync.Map).Store(clientFd, struct{}{})
	w.queryCacheMu.Lock()
	w.queryCaches.Put(query.String(), newQueryCache())
	w.queryCacheMu.Unlock()
}

// RemoveWatcher removes a client from the watchlist for a query.
func (w *QueryWatcher) RemoveWatcher(query DSQLQuery, clientFd int) { //nolint:gocritic
	if clients, ok := w.WatchList.Load(query); ok {
		clients.(*sync.Map).Delete(clientFd)
		log.Info(fmt.Printf("Removed client %d from query %s", clientFd, query))
		// If no more clients for this query, remove the query from WatchList
		if w.clientCount(clients.(*sync.Map)) == 0 {
			w.WatchList.Delete(query)

			//	Remove this Query's cached data.
			w.queryCacheMu.Lock()
			w.queryCaches.Delete(query.String())
			w.queryCacheMu.Unlock()

			log.Info(fmt.Printf("Removed query %s from WatchList", query))
		}
	}
}

// Helper function to count clients
func (w *QueryWatcher) clientCount(clients *sync.Map) int {
	count := 0
	clients.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
