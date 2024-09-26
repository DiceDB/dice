package querywatcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"syscall"
	"time"

	"github.com/dicedb/dice/internal/comm"

	"github.com/ohler55/ojg/jp"

	"github.com/dicedb/dice/internal/object"

	"github.com/dicedb/dice/internal/sql"

	"github.com/cockroachdb/swiss"
	"github.com/dicedb/dice/internal/clientio"
	dstore "github.com/dicedb/dice/internal/store"
)

type (
	cacheStore *swiss.Map[string, *object.Obj]

	// WatchSubscription represents a subscription to watch a query.
	WatchSubscription struct {
		Subscribe bool          // true for subscribe, false for unsubscribe
		Query     sql.DSQLQuery // query to watch
		ClientFD  int           // client file descriptor
		CacheChan chan *[]struct {
			Key   string
			Value *object.Obj
		} // channel to receive cache data for this query
		QwatchClientChan   chan comm.QwatchResponse // Generic channel for HTTP/Websockets etc.
		ClientIdentifierID uint32                   // Helps identify qwatch client on httpserver side
	}

	// AdhocQueryResult represents the result of an adhoc query.
	AdhocQueryResult struct {
		Result      *[]sql.QueryResultRow
		Fingerprint string
		Err         error
	}

	// AdhocQuery represents an adhoc query request.
	AdhocQuery struct {
		Query        sql.DSQLQuery
		ResponseChan chan AdhocQueryResult
	}

	// QueryManager watches for changes in keys and notifies clients.
	QueryManager struct {
		WatchList    sync.Map                       // WatchList is a map of query string to their respective clients, type: map[string]*sync.Map[int]struct{}
		QueryCache   *swiss.Map[string, cacheStore] // QueryCache is a map of fingerprints to their respective data caches
		QueryCacheMu sync.RWMutex
		logger       *slog.Logger
	}

	HTTPQwatchResponse struct {
		Cmd   string `json:"cmd"`
		Query string `json:"query"`
		Data  []any  `json:"data"`
	}

	ClientIdentifier struct {
		ClientIdentifierID int
		IsHTTPClient       bool
	}
)

var (
	// WatchSubscriptionChan is the channel to receive updates about query subscriptions.
	WatchSubscriptionChan chan WatchSubscription

	// AdhocQueryChan is the channel to receive adhoc queries.
	AdhocQueryChan chan AdhocQuery
)

func NewClientIdentifier(clientIdentifierID int, isHTTPClient bool) ClientIdentifier {
	return ClientIdentifier{
		ClientIdentifierID: clientIdentifierID,
		IsHTTPClient:       isHTTPClient,
	}
}

// NewQueryManager initializes a new QueryManager.
func NewQueryManager(logger *slog.Logger) *QueryManager {
	WatchSubscriptionChan = make(chan WatchSubscription)
	AdhocQueryChan = make(chan AdhocQuery, 1000)
	return &QueryManager{
		WatchList:  sync.Map{},
		QueryCache: swiss.New[string, cacheStore](0),
		logger:     logger,
	}
}

func newCacheStore() cacheStore {
	return swiss.New[string, *object.Obj](0)
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
			var client ClientIdentifier
			if event.QwatchClientChan != nil {
				client = NewClientIdentifier(int(event.ClientIdentifierID), true)
			} else {
				client = NewClientIdentifier(event.ClientFD, false)
			}

			if event.Subscribe {
				w.addWatcher(&event.Query, client, event.QwatchClientChan, event.CacheChan)
			} else {
				w.removeWatcher(&event.Query, client, event.QwatchClientChan)
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

		query, err := sql.ParseQuery(queryString)
		if err != nil {
			w.logger.Error(
				"error parsing query",
				slog.String("query", queryString),
			)
			return true
		}

		// Check if the key matches the regex
		if query.Where != nil {
			matches, err := sql.EvaluateWhereClause(query.Where, sql.QueryResultRow{Key: event.Key, Value: event.Value}, make(map[string]jp.Expr))
			if err != nil || !matches {
				return true
			}
		}

		w.updateQueryCache(query.Fingerprint, event)

		queryResult, err := w.runQuery(&query)
		if err != nil {
			w.logger.Error(err.Error())
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
		w.logger.Warn("Fingerprint not found in cacheStore", slog.String("fingerprint", queryFingerprint))
		return
	}

	switch event.Operation {
	case dstore.Set:
		((*swiss.Map[string, *object.Obj])(store)).Put(event.Key, &event.Value)
	case dstore.Del:
		((*swiss.Map[string, *object.Obj])(store)).Delete(event.Key)
	default:
		w.logger.Warn("Unknown operation", slog.String("operation", event.Operation))
	}
}

func (w *QueryManager) notifyClients(query *sql.DSQLQuery, clients *sync.Map, queryResult *[]sql.QueryResultRow) {
	encodedResult := clientio.Encode(clientio.CreatePushResponse(query, queryResult), false)

	clients.Range(func(clientKey, clientVal interface{}) bool {
		// Identify the type of client and respond accordingly
		switch clientIdentifier := clientKey.(ClientIdentifier); {
		case clientIdentifier.IsHTTPClient:
			qwatchClientResponseChannel := clientVal.(chan comm.QwatchResponse)
			qwatchClientResponseChannel <- comm.QwatchResponse{
				ClientIdentifierID: uint32(clientIdentifier.ClientIdentifierID),
				Result:             encodedResult,
				Error:              nil,
			}
		case !clientIdentifier.IsHTTPClient:
			// We use a retry mechanism here as the client's socket may be temporarily unavailable for writes due to the
			// high number of writes that are possible in qwatch. Without this mechanism, the client may be removed from the
			// watchlist prematurely.
			// TODO:
			//  1. Replace with thread pool to prevent launching an unbounded number of goroutines.
			//  2. Each client's writes should be sent in a serialized manner, maybe a per-client queue should be maintained
			//   here. A single queue-per-client is also helpful when the client file descriptor is closed and the queue can
			//   just be destroyed.
			clientFD := clientIdentifier.ClientIdentifierID
			// This is a regular client, use clientFD to send the response
			go w.sendWithRetry(query, clientFD, encodedResult)
		default:
			w.logger.Warn("Invalid Client, response channel invalid.")
		}

		return true
	})
}

// sendWithRetry writes data to a client file descriptor with retries. It writes with an exponential backoff.
func (w *QueryManager) sendWithRetry(query *sql.DSQLQuery, clientFD int, data []byte) {
	maxRetries := 20
	retryDelay := 20 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		_, err := syscall.Write(clientFD, data)
		if err == nil {
			return
		}

		if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
			time.Sleep(retryDelay)
			retryDelay *= 2 // exponential backoff
			continue
		}

		w.logger.Error(
			"error writing to client",
			slog.Int("client", clientFD),
			slog.Any("error", err),
		)
		w.removeWatcher(query, NewClientIdentifier(clientFD, false), nil)
		return
	}
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
func (w *QueryManager) addWatcher(query *sql.DSQLQuery, clientIdentifier ClientIdentifier,
	qwatchClientChan chan comm.QwatchResponse, cacheChan chan *[]struct {
		Key   string
		Value *object.Obj
	}) {
	queryString := query.String()

	clients, _ := w.WatchList.LoadOrStore(queryString, &sync.Map{})
	if qwatchClientChan != nil {
		clients.(*sync.Map).Store(clientIdentifier, qwatchClientChan)
	} else {
		clients.(*sync.Map).Store(clientIdentifier, struct{}{})
	}

	w.QueryCacheMu.Lock()
	defer w.QueryCacheMu.Unlock()

	cache := newCacheStore()
	// Hydrate the cache with data from all shards.
	// TODO: We need to ensure we receive cache data from all shards once we have multithreading in place.
	//  For now we only expect one update.
	kvs := <-cacheChan
	for _, kv := range *kvs {
		((*swiss.Map[string, *object.Obj])(cache)).Put(kv.Key, kv.Value)
	}

	w.QueryCache.Put(query.Fingerprint, cache)
}

// removeWatcher removes a client from the watchlist for a query.
func (w *QueryManager) removeWatcher(query *sql.DSQLQuery, clientIdentifier ClientIdentifier,
	qwatchClientChan chan comm.QwatchResponse) {
	queryString := query.String()
	if clients, ok := w.WatchList.Load(queryString); ok {
		if qwatchClientChan != nil {
			clients.(*sync.Map).Delete(clientIdentifier)
			w.logger.Debug("HTTP client no longer watching query",
				slog.Any("clientIdentifierId", clientIdentifier.ClientIdentifierID),
				slog.Any("queryString", queryString))
		} else {
			clients.(*sync.Map).Delete(clientIdentifier)
			w.logger.Debug("client no longer watching query",
				slog.Int("client", clientIdentifier.ClientIdentifierID),
				slog.String("query", queryString))
		}

		// If no more clients for this query, remove the query from WatchList
		if w.clientCount(clients.(*sync.Map)) == 0 {
			w.WatchList.Delete(queryString)

			// Remove this Query's cached data.
			w.QueryCacheMu.Lock()
			w.QueryCache.Delete(query.Fingerprint)
			w.QueryCacheMu.Unlock()

			w.logger.Debug("no longer watching query", slog.String("query", queryString))
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
func (w *QueryManager) runQuery(query *sql.DSQLQuery) (*[]sql.QueryResultRow, error) {
	w.QueryCacheMu.RLock()
	defer w.QueryCacheMu.RUnlock()

	store, ok := w.QueryCache.Get(query.Fingerprint)
	if !ok {
		return nil, fmt.Errorf("fingerprint was not found in the cache: %s", query.Fingerprint)
	}

	result, err := sql.ExecuteQuery(query, store)
	return &result, err
}
