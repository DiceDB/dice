package querywatcher

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/cockroachdb/swiss"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/regex"
	dstore "github.com/dicedb/dice/internal/store"
)

type (
	cacheStore *swiss.Map[string, *dstore.Obj]

	// WatchSubscription represents a subscription to watch a query.
	WatchSubscription struct {
		Subscribe bool                    // true for subscribe, false for unsubscribe
		Query     DSQLQuery               // query to watch
		ClientFD  int                     // client file descriptor
		CacheChan chan *[]dstore.KeyValue // channel to receive cache data for this query
	}

	// AdhocQueryResult represents the result of an adhoc query.
	AdhocQueryResult struct {
		Result      *[]dstore.DSQLQueryResultRow
		Fingerprint string
		Err         error
	}

	// AdhocQuery represents an adhoc query request.
	AdhocQuery struct {
		Query        DSQLQuery
		ResponseChan chan AdhocQueryResult
	}

	// QueryManager watches for changes in keys and notifies clients.
	QueryManager struct {
		WatchList    sync.Map                       // WatchList is a map of query string to their respective clients, type: map[string]*sync.Map[int]struct{}
		QueryCache   *swiss.Map[string, cacheStore] // QueryCache is a map of fingerprints to their respective data caches
		QueryCacheMu sync.RWMutex
	}
)

var (
	// WatchSubscriptionChan is the channel to receive updates about query subscriptions.
	WatchSubscriptionChan chan WatchSubscription

	// AdhocQueryChan is the channel to receive adhoc queries.
	AdhocQueryChan chan AdhocQuery
)

// NewQueryManager initializes a new QueryManager.
func NewQueryManager() *QueryManager {
	WatchSubscriptionChan = make(chan WatchSubscription)
	AdhocQueryChan = make(chan AdhocQuery, 1000)
	return &QueryManager{
		WatchList:  sync.Map{},
		QueryCache: swiss.New[string, cacheStore](0),
	}
}

func newCacheStore() cacheStore {
	return swiss.New[string, *dstore.Obj](0)
}

// Run starts the QueryManager's main loops.
func (w *QueryManager) Run(ctx context.Context, watchChan <-chan dstore.WatchEvent) {
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
func (w *QueryManager) listenForSubscriptions(ctx context.Context) {
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
func (w *QueryManager) watchKeys(ctx context.Context, watchChan <-chan dstore.WatchEvent) {
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
func (w *QueryManager) processWatchEvent(event dstore.WatchEvent) {
	// Iterate over the watchlist to go through the query string
	// and the corresponding client connections to that query string
	w.WatchList.Range(func(key, value interface{}) bool {
		queryString := key.(string)
		clients := value.(*sync.Map)

		query, err := ParseQuery(queryString)
		if err != nil {
			log.Error(fmt.Sprintf("Error parsing query: %s", queryString))
			return true
		}

		// Check if the key matches the regex
		if !regex.WildCardMatch(query.KeyRegex, event.Key) {
			return true
		}

		w.updateQueryCache(query.Fingerprint, event)

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
func (w *QueryManager) updateQueryCache(queryFingerprint string, event dstore.WatchEvent) {
	w.QueryCacheMu.Lock()
	defer w.QueryCacheMu.Unlock()

	store, ok := w.QueryCache.Get(queryFingerprint)
	if !ok {
		log.Warnf("Fingerprint not found in cacheStore: %s", queryFingerprint)
		return
	}

	switch event.Operation {
	case dstore.Set:
		((*swiss.Map[string, *dstore.Obj])(store)).Put(event.Key, event.Value)
	case dstore.Del:
		((*swiss.Map[string, *dstore.Obj])(store)).Delete(event.Key)
	default:
		log.Warnf("Unknown operation: %s", event.Operation)
	}
}

// notifyClient notifies all clients watching a query about the new result.
func (w *QueryManager) notifyClients(query *DSQLQuery, clients *sync.Map, queryResult *[]dstore.DSQLQueryResultRow) {
	encodedResult := clientio.Encode(CreatePushResponse(query, queryResult), false)
	// Send the result to the client
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
func (w *QueryManager) serveAdhocQueries(ctx context.Context) {
	for {
		select {
		case query := <-AdhocQueryChan:
			result, err := w.runQuery(&query.Query)
			query.ResponseChan <- AdhocQueryResult{
				Result:      result,
				Fingerprint: query.Query.Fingerprint,
				Err:         err,
			}
		case <-ctx.Done():
			return
		}
	}
}

// addWatcher adds a client as a watcher to a query.
func (w *QueryManager) addWatcher(query *DSQLQuery, clientFD int, cacheChan chan *[]dstore.KeyValue) {
	queryString := query.String()

	clients, _ := w.WatchList.LoadOrStore(queryString, &sync.Map{})
	clients.(*sync.Map).Store(clientFD, struct{}{})

	w.QueryCacheMu.Lock()
	defer w.QueryCacheMu.Unlock()

	cache := newCacheStore()
	// Hydrate the cache with data from all shards.
	// TODO: We need to ensure we receive cache data from all shards once we have multithreading in place.
	//  For now we only expect one update.
	kvs := <-cacheChan
	for _, kv := range *kvs {
		((*swiss.Map[string, *dstore.Obj])(cache)).Put(kv.Key, kv.Value)
	}

	w.QueryCache.Put(query.Fingerprint, cache)
}

// removeWatcher removes a client from the watchlist for a query.
func (w *QueryManager) removeWatcher(query *DSQLQuery, clientFD int) {
	queryString := query.String()
	if clients, ok := w.WatchList.Load(queryString); ok {
		clients.(*sync.Map).Delete(clientFD)
		log.Info(fmt.Sprintf("client '%d' no longer watching query: %s", clientFD, queryString))

		// If no more clients for this query, remove the query from WatchList
		if w.clientCount(clients.(*sync.Map)) == 0 {
			w.WatchList.Delete(queryString)

			// Remove this Query's cached data.
			w.QueryCacheMu.Lock()
			w.QueryCache.Delete(query.Fingerprint)
			w.QueryCacheMu.Unlock()

			log.Info(fmt.Sprintf("no longer watching query: %s", queryString))
		}
	}
}

// clientCount returns the number of clients watching a query.
func (w *QueryManager) clientCount(clients *sync.Map) int {
	count := 0
	clients.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// runQuery executes a query on its respective cache.
func (w *QueryManager) runQuery(query *DSQLQuery) (*[]dstore.DSQLQueryResultRow, error) {
	w.QueryCacheMu.RLock()
	defer w.QueryCacheMu.RUnlock()

	store, ok := w.QueryCache.Get(query.Fingerprint)
	if !ok {
		return nil, fmt.Errorf("regex was not found in the cache: %s", query.KeyRegex)
	}

	result, err := ExecuteQuery(query, store)
	return &result, err
}
