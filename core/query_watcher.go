// Package core provides core functionality for the DiceDB query watching system.
package core

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/dicedb/dice/internal/constants"

	"github.com/charmbracelet/log"
	"github.com/cockroachdb/swiss"
)

type (
	cacheStore *swiss.Map[string, *Obj]

	KeyValue struct {
		Key   string
		Value *Obj
	}

	// WatchEvent represents a change in a watched key.
	WatchEvent struct {
		Key       string
		Operation string
		Value     *Obj
	}

	// WatchSubscription represents a subscription to watch a query.
	WatchSubscription struct {
		Subscribe bool             // true for subscribe, false for unsubscribe
		Query     DSQLQuery        // query to watch
		ClientFD  int              // client file descriptor
		CacheChan chan *[]KeyValue // channel to receive cache data for this query
	}

	// AdhocQueryResult represents the result of an adhoc query.
	AdhocQueryResult struct {
		Result *[]DSQLQueryResultRow
		Err    error
	}

	// AdhocQuery represents an adhoc query request.
	AdhocQuery struct {
		Query        DSQLQuery
		ResponseChan chan AdhocQueryResult
	}

	// QueryWatcher watches for changes in keys and notifies clients.
	QueryWatcher struct {
		WatchList    sync.Map                       // WatchList is a map of queries to their respective clients, type: map[DSQLQuery]*sync.Map[int]struct{}
		QueryCache   *swiss.Map[string, cacheStore] // QueryCache is a map of queries to their respective data caches
		QueryCacheMu sync.RWMutex
	}
)

var (
	// WatchSubscriptionChan is the channel to receive updates about query subscriptions.
	WatchSubscriptionChan chan WatchSubscription

	// AdhocQueryChan is the channel to receive adhoc queries.
	AdhocQueryChan chan AdhocQuery
)

// NewQueryWatcher initializes a new QueryWatcher.
func NewQueryWatcher() *QueryWatcher {
	WatchSubscriptionChan = make(chan WatchSubscription)
	AdhocQueryChan = make(chan AdhocQuery, 1000)

	return &QueryWatcher{
		WatchList:  sync.Map{},
		QueryCache: swiss.New[string, cacheStore](0),
	}
}

func newCacheStore() cacheStore {
	return swiss.New[string, *Obj](0)
}

// Run starts the QueryWatcher's main loops.
func (w *QueryWatcher) Run(ctx context.Context, watchChan <-chan WatchEvent) {
	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		w.listenForSubscriptions(ctx)
	}()

	go func() {
		defer wg.Done()
		w.watchKeys(ctx, watchChan)
	}()

	go func() {
		defer wg.Done()
		w.serveAdhocQueries(ctx)
	}()

	<-ctx.Done()
	wg.Wait()
}

// listenForSubscriptions listens for query subscriptions and unsubscriptions.
func (w *QueryWatcher) listenForSubscriptions(ctx context.Context) {
	for {
		select {
		case event := <-WatchSubscriptionChan:
			if event.Subscribe {
				w.addWatcher(&event.Query, event.ClientFD, event.CacheChan)
			} else {
				w.removeWatcher(&event.Query, event.ClientFD)
			}
		case <-ctx.Done():
			return
		}
	}
}

// watchKeys watches for changes in keys and notifies clients.
func (w *QueryWatcher) watchKeys(ctx context.Context, watchChan <-chan WatchEvent) {
	for {
		select {
		case event := <-watchChan:
			w.processWatchEvent(event)
		case <-ctx.Done():
			return
		}
	}
}

// processWatchEvent processes a single watch event.
func (w *QueryWatcher) processWatchEvent(event WatchEvent) {
	w.WatchList.Range(func(key, value interface{}) bool {
		query := key.(DSQLQuery)
		clients := value.(*sync.Map)

		if !WildCardMatch(query.KeyRegex, event.Key) {
			return true
		}

		w.updateQueryCache(&query, event)
		queryResult, err := w.runQuery(&query)
		if err != nil {
			log.Error(err)
			return true
		}

		w.notifyClients(&query, clients, queryResult)
		return true
	})
}

// updateQueryCache updates the query cache based on the watch event.
func (w *QueryWatcher) updateQueryCache(query *DSQLQuery, event WatchEvent) {
	w.QueryCacheMu.Lock()
	defer w.QueryCacheMu.Unlock()

	store, ok := w.QueryCache.Get(query.String())
	if !ok {
		log.Warnf("Query not found in cacheStore: %s", query)
		return
	}

	switch event.Operation {
	case constants.Set:
		((*swiss.Map[string, *Obj])(store)).Put(event.Key, event.Value)
	case constants.Del:
		((*swiss.Map[string, *Obj])(store)).Delete(event.Key)
	default:
		log.Warnf("Unknown operation: %s", event.Operation)
	}
}

// notifyClients notifies all clients watching a query about the new result.
func (w *QueryWatcher) notifyClients(query *DSQLQuery, clients *sync.Map, queryResult *[]DSQLQueryResultRow) {
	encodedResult := Encode(CreatePushResponse(query, queryResult), false)
	clients.Range(func(clientKey, _ interface{}) bool {
		clientFD := clientKey.(int)
		_, err := syscall.Write(clientFD, encodedResult)
		if err != nil {
			w.removeWatcher(query, clientFD)
		}
		return true
	})
}

// serveAdhocQueries listens for adhoc queries, executes them, and sends the result back to the client.
func (w *QueryWatcher) serveAdhocQueries(ctx context.Context) {
	for {
		select {
		case query := <-AdhocQueryChan:
			result, err := w.runQuery(&query.Query)
			query.ResponseChan <- AdhocQueryResult{Result: result, Err: err}
		case <-ctx.Done():
			return
		}
	}
}

// addWatcher adds a client as a watcher to a query.
func (w *QueryWatcher) addWatcher(query *DSQLQuery, clientFD int, cacheChan chan *[]KeyValue) {
	clients, _ := w.WatchList.LoadOrStore(*query, &sync.Map{})
	clients.(*sync.Map).Store(clientFD, struct{}{})

	w.QueryCacheMu.Lock()
	defer w.QueryCacheMu.Unlock()

	cache := newCacheStore()
	// Hydrate the cache with data from all shards.
	// TODO: We need to ensure we receive cache data from all shards once we have multithreading in place.
	//  For now we only expect one update.
	kvs := <-cacheChan
	for _, kv := range *kvs {
		((*swiss.Map[string, *Obj])(cache)).Put(kv.Key, kv.Value)
	}

	w.QueryCache.Put(query.String(), cache)
}

// removeWatcher removes a client from the watchlist for a query.
func (w *QueryWatcher) removeWatcher(query *DSQLQuery, clientFD int) {
	if clients, ok := w.WatchList.Load(*query); ok {
		clients.(*sync.Map).Delete(clientFD)
		log.Info(fmt.Sprintf("client '%d' no longer watching query: %s", clientFD, query))

		// If no more clients for this query, remove the query from WatchList
		if w.clientCount(clients.(*sync.Map)) == 0 {
			w.WatchList.Delete(*query)

			// Remove this Query's cached data.
			w.QueryCacheMu.Lock()
			w.QueryCache.Delete(query.String())
			w.QueryCacheMu.Unlock()

			log.Info(fmt.Sprintf("no longer watching query: %s", query))
		}
	}
}

// clientCount returns the number of clients watching a query.
func (w *QueryWatcher) clientCount(clients *sync.Map) int {
	count := 0
	clients.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// runQuery executes a query on its respective cache.
func (w *QueryWatcher) runQuery(query *DSQLQuery) (*[]DSQLQueryResultRow, error) {
	w.QueryCacheMu.RLock()
	defer w.QueryCacheMu.RUnlock()

	store, ok := w.QueryCache.Get(query.String())
	if !ok {
		return nil, fmt.Errorf("query was not found in the cache: %s", query)
	}

	result, err := ExecuteQuery(query, store)
	return &result, err
}
