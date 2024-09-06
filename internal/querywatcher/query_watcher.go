package querywatcher

import (
	"context"
	"crypto/md5"
	"encoding/hex"
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

	// QWatchManager watches for changes in keys and notifies clients.
	QWatchManager struct {
		WatchList        sync.Map                       // WatchList is a map of fingerprints to their respective clients, type: map[string]*sync.Map[int]struct{}
		QueryCache       *swiss.Map[string, cacheStore] // QueryCache is a map of fingerprints to their respective data caches
		FingerPrintCache *swiss.Map[string, *DSQLQuery] // FingerPrintCache is a map of fingerprint to the respective query
		QueryCacheMu     sync.RWMutex
	}
)

var (
	// WatchSubscriptionChan is the channel to receive updates about query subscriptions.
	WatchSubscriptionChan chan WatchSubscription

	// AdhocQueryChan is the channel to receive adhoc queries.
	AdhocQueryChan chan AdhocQuery
)

// NewQWatchManager initializes a new QWatchManager.
func NewQWatchManager() *QWatchManager {
	WatchSubscriptionChan = make(chan WatchSubscription)
	AdhocQueryChan = make(chan AdhocQuery, 1000)
	return &QWatchManager{
		WatchList:        sync.Map{},
		QueryCache:       swiss.New[string, cacheStore](0),
		FingerPrintCache: swiss.New[string, *DSQLQuery](0),
	}
}

func newCacheStore() cacheStore {
	return swiss.New[string, *dstore.Obj](0)
}

// Run starts the QWatchManager's main loops.
func (w *QWatchManager) Run(ctx context.Context, watchChan <-chan dstore.WatchEvent) {
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
func (w *QWatchManager) listenForSubscriptions(ctx context.Context) {
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
func (w *QWatchManager) watchKeys(ctx context.Context, watchChan <-chan dstore.WatchEvent) {
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
func (w *QWatchManager) processWatchEvent(event dstore.WatchEvent) {
	w.WatchList.Range(func(fingerPrint, value interface{}) bool {
		query, ok := w.FingerPrintCache.Get(fingerPrint.(string))
		if !ok {
			log.Warnf("Fingerprint not found in cacheStore: %s", fingerPrint)
			return true
		}

		clients := value.(*sync.Map)

		if !regex.WildCardMatch(query.KeyRegex, event.Key) {
			return true
		}

		w.updateQueryCache(fingerPrint.(string), event)
		queryResult, err := w.runQuery(query)
		if err != nil {
			log.Error(err)
			return true
		}

		w.notifyClients(query, clients, queryResult)
		return true
	})
}

// updateQueryCache updates the query cache based on the watch event.
func (w *QWatchManager) updateQueryCache(queryFingerprint string, event dstore.WatchEvent) {
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

// notifyClients notifies all clients watching a query about the new result.
func (w *QWatchManager) notifyClients(query *DSQLQuery, clients *sync.Map, queryResult *[]dstore.DSQLQueryResultRow) {
	encodedResult := clientio.Encode(CreatePushResponse(query, queryResult), false)
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
func (w *QWatchManager) serveAdhocQueries(ctx context.Context) {
	for {
		select {
		case query := <-AdhocQueryChan:
			result, err := w.runQuery(&query.Query)
			query.ResponseChan <- AdhocQueryResult{
				Result:      result,
				Fingerprint: generateQueryFingerprint(&query.Query),
				Err:         err,
			}
		case <-ctx.Done():
			return
		}
	}
}

// generateQueryFingerprint used to creating a common fingerprint for similar queries
func generateQueryFingerprint(query *DSQLQuery) string {
	hash := md5.Sum([]byte(query.KeyRegex))
	fingerprint := hex.EncodeToString(hash[:])
	return "f_" + fingerprint
}

// addWatcher adds a client as a watcher to a query.
func (w *QWatchManager) addWatcher(query *DSQLQuery, clientFD int, cacheChan chan *[]dstore.KeyValue) {
	queryFingerprint := generateQueryFingerprint(query)
	w.FingerPrintCache.Put(queryFingerprint, query)

	clients, _ := w.WatchList.LoadOrStore(queryFingerprint, &sync.Map{})
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

	w.QueryCache.Put(queryFingerprint, cache)
}

// removeWatcher removes a client from the watchlist for a query.
func (w *QWatchManager) removeWatcher(query *DSQLQuery, clientFD int) {
	queryFingerprint := generateQueryFingerprint(query)
	if clients, ok := w.WatchList.Load(queryFingerprint); ok {
		clients.(*sync.Map).Delete(clientFD)
		log.Info(fmt.Sprintf("client '%d' no longer watching fingerprint: %s", clientFD, queryFingerprint))

		// If no more clients for this query, remove the query from WatchList
		if w.clientCount(clients.(*sync.Map)) == 0 {
			w.WatchList.Delete(queryFingerprint)

			// Remove this Query's cached data.
			w.QueryCacheMu.Lock()
			w.QueryCache.Delete(queryFingerprint)
			w.FingerPrintCache.Delete(queryFingerprint)
			w.QueryCacheMu.Unlock()

			log.Info(fmt.Sprintf("no longer watching query: %s", query))
		}
	}
}

// clientCount returns the number of clients watching a query.
func (w *QWatchManager) clientCount(clients *sync.Map) int {
	count := 0
	clients.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// runQuery executes a query on its respective cache.
func (w *QWatchManager) runQuery(query *DSQLQuery) (*[]dstore.DSQLQueryResultRow, error) {
	w.QueryCacheMu.RLock()
	defer w.QueryCacheMu.RUnlock()

	queryFingerprint := generateQueryFingerprint(query)
	store, ok := w.QueryCache.Get(queryFingerprint)
	if !ok {
		return nil, fmt.Errorf("regex was not found in the cache: %s", query.KeyRegex)
	}

	result, err := ExecuteQuery(query, store)
	return &result, err
}
