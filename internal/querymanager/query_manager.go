package querymanager

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"syscall"
	"time"

	"github.com/dicedb/dice/internal/comm"
	"github.com/dicedb/dice/internal/common"

	"github.com/ohler55/ojg/jp"

	"github.com/dicedb/dice/internal/object"

	"github.com/dicedb/dice/internal/sql"

	"github.com/dicedb/dice/internal/clientio"
	dstore "github.com/dicedb/dice/internal/store"
)

type (
	CacheStore common.ITable[string, *object.Obj]

	// QuerySubscription represents a subscription to watch a query.
	QuerySubscription struct {
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

	// Manager watches for changes in keys and notifies clients.
	Manager struct {
		WatchList    sync.Map                          // WatchList is a map of query string to their respective clients, type: map[string]*sync.Map[int]struct{}
		QueryCache   common.ITable[string, CacheStore] // QueryCache is a map of fingerprints to their respective data caches
		QueryCacheMu sync.RWMutex
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
	// QuerySubscriptionChan is the channel to receive updates about query subscriptions.
	QuerySubscriptionChan chan QuerySubscription

	// AdhocQueryChan is the channel to receive adhoc queries.
	AdhocQueryChan chan AdhocQuery
)

func NewClientIdentifier(clientIdentifierID int, isHTTPClient bool) ClientIdentifier {
	return ClientIdentifier{
		ClientIdentifierID: clientIdentifierID,
		IsHTTPClient:       isHTTPClient,
	}
}

func NewQueryCacheStoreRegMap() common.ITable[string, CacheStore] {
	return &common.RegMap[string, CacheStore]{
		M: make(map[string]CacheStore),
	}
}

func NewQueryCacheStore() common.ITable[string, CacheStore] {
	return NewQueryCacheStoreRegMap()
}

func NewCacheStoreRegMap() CacheStore {
	return &common.RegMap[string, *object.Obj]{
		M: make(map[string]*object.Obj),
	}
}

func NewCacheStore() CacheStore {
	return NewCacheStoreRegMap()
}

// NewQueryManager initializes a new Manager.
func NewQueryManager() *Manager {
	QuerySubscriptionChan = make(chan QuerySubscription)
	AdhocQueryChan = make(chan AdhocQuery, 1000)
	return &Manager{
		WatchList:  sync.Map{},
		QueryCache: NewQueryCacheStore(),
	}
}

// Run starts the Manager's main loops.
func (m *Manager) Run(ctx context.Context, watchChan <-chan dstore.QueryWatchEvent) {
	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		m.listenForSubscriptions(ctx)
	}()

	go func() {
		defer wg.Done()
		m.watchKeys(ctx, watchChan)
	}()

	go func() {
		defer wg.Done()
		m.serveAdhocQueries(ctx)
	}()

	<-ctx.Done()
	wg.Wait()
}

// listenForSubscriptions listens for query subscriptions and unsubscriptions.
func (m *Manager) listenForSubscriptions(ctx context.Context) {
	for {
		select {
		case event := <-QuerySubscriptionChan:
			var client ClientIdentifier
			if event.QwatchClientChan != nil {
				client = NewClientIdentifier(int(event.ClientIdentifierID), true)
			} else {
				client = NewClientIdentifier(event.ClientFD, false)
			}

			if event.Subscribe {
				m.addWatcher(&event.Query, client, event.QwatchClientChan, event.CacheChan)
			} else {
				m.removeWatcher(&event.Query, client, event.QwatchClientChan)
			}
		case <-ctx.Done():
			return
		}
	}
}

// watchKeys watches for changes in keys and notifies clients.
func (m *Manager) watchKeys(ctx context.Context, watchChan <-chan dstore.QueryWatchEvent) {
	for {
		select {
		case event := <-watchChan:
			m.processWatchEvent(event)
		case <-ctx.Done():
			return
		}
	}
}

// processWatchEvent processes a single watch event.
func (m *Manager) processWatchEvent(event dstore.QueryWatchEvent) {
	// Iterate over the watchlist to go through the query string
	// and the corresponding client connections to that query string
	m.WatchList.Range(func(key, value interface{}) bool {
		queryString := key.(string)
		clients := value.(*sync.Map)

		query, err := sql.ParseQuery(queryString)
		if err != nil {
			slog.Error(
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

		m.updateQueryCache(query.Fingerprint, event)

		queryResult, err := m.runQuery(&query)
		if err != nil {
			slog.Error(err.Error())
			return true
		}

		m.notifyClients(&query, clients, queryResult)
		return true
	})
}

// updateQueryCache updates the query cache based on the watch event.
func (m *Manager) updateQueryCache(queryFingerprint string, event dstore.QueryWatchEvent) {
	m.QueryCacheMu.Lock()
	defer m.QueryCacheMu.Unlock()

	store, ok := m.QueryCache.Get(queryFingerprint)
	if !ok {
		slog.Warn("Fingerprint not found in CacheStore", slog.String("fingerprint", queryFingerprint))
		return
	}

	switch event.Operation {
	case dstore.Set:
		store.Put(event.Key, &event.Value)
	case dstore.Del:
		store.Delete(event.Key)
	default:
		slog.Warn("Unknown operation", slog.String("operation", event.Operation))
	}
}

func (m *Manager) notifyClients(query *sql.DSQLQuery, clients *sync.Map, queryResult *[]sql.QueryResultRow) {
	encodedResult := clientio.Encode(GenericWatchResponse(sql.Qwatch, query.String(), *queryResult), false)

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
			go m.sendWithRetry(query, clientFD, encodedResult)
		default:
			slog.Warn("Invalid Client, response channel invalid.")
		}

		return true
	})
}

// sendWithRetry writes data to a client file descriptor with retries. It writes with an exponential backoff.
func (m *Manager) sendWithRetry(query *sql.DSQLQuery, clientFD int, data []byte) {
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

		slog.Error(
			"error writing to client",
			slog.Int("client", clientFD),
			slog.Any("error", err),
		)
		m.removeWatcher(query, NewClientIdentifier(clientFD, false), nil)
		return
	}
}

// serveAdhocQueries listens for adhoc queries, executes them, and sends the result back to the client.
func (m *Manager) serveAdhocQueries(ctx context.Context) {
	for {
		select {
		case query := <-AdhocQueryChan:
			result, err := m.runQuery(&query.Query)
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
func (m *Manager) addWatcher(query *sql.DSQLQuery, clientIdentifier ClientIdentifier,
	qwatchClientChan chan comm.QwatchResponse, cacheChan chan *[]struct {
		Key   string
		Value *object.Obj
	}) {
	queryString := query.String()

	clients, _ := m.WatchList.LoadOrStore(queryString, &sync.Map{})
	if qwatchClientChan != nil {
		clients.(*sync.Map).Store(clientIdentifier, qwatchClientChan)
	} else {
		clients.(*sync.Map).Store(clientIdentifier, struct{}{})
	}

	m.QueryCacheMu.Lock()
	defer m.QueryCacheMu.Unlock()

	cache := NewCacheStore()
	// Hydrate the cache with data from all shards.
	// TODO: We need to ensure we receive cache data from all shards once we have multithreading in place.
	//  For now we only expect one update.
	kvs := <-cacheChan
	for _, kv := range *kvs {
		cache.Put(kv.Key, kv.Value)
	}

	m.QueryCache.Put(query.Fingerprint, cache)
}

// removeWatcher removes a client from the watchlist for a query.
func (m *Manager) removeWatcher(query *sql.DSQLQuery, clientIdentifier ClientIdentifier,
	qwatchClientChan chan comm.QwatchResponse) {
	queryString := query.String()
	if clients, ok := m.WatchList.Load(queryString); ok {
		if qwatchClientChan != nil {
			clients.(*sync.Map).Delete(clientIdentifier)
			slog.Debug("HTTP client no longer watching query",
				slog.Any("clientIdentifierId", clientIdentifier.ClientIdentifierID),
				slog.Any("queryString", queryString))
		} else {
			clients.(*sync.Map).Delete(clientIdentifier)
			slog.Debug("client no longer watching query",
				slog.Int("client", clientIdentifier.ClientIdentifierID),
				slog.String("query", queryString))
		}

		// If no more clients for this query, remove the query from WatchList
		if m.clientCount(clients.(*sync.Map)) == 0 {
			m.WatchList.Delete(queryString)

			// Remove this Query's cached data.
			m.QueryCacheMu.Lock()
			m.QueryCache.Delete(query.Fingerprint)
			m.QueryCacheMu.Unlock()

			slog.Debug("no longer watching query", slog.String("query", queryString))
		}
	}
}

// clientCount returns the number of clients watching a query.
func (m *Manager) clientCount(clients *sync.Map) int {
	count := 0
	clients.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// runQuery executes a query on its respective cache.
func (m *Manager) runQuery(query *sql.DSQLQuery) (*[]sql.QueryResultRow, error) {
	m.QueryCacheMu.RLock()
	defer m.QueryCacheMu.RUnlock()

	store, ok := m.QueryCache.Get(query.Fingerprint)
	if !ok {
		return nil, fmt.Errorf("fingerprint was not found in the cache: %s", query.Fingerprint)
	}

	result, err := sql.ExecuteQuery(query, store)
	return &result, err
}
