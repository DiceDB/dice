package watchmanager

import (
	"context"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	dstore "github.com/dicedb/dice/internal/store"
	"log/slog"
	"sync"
)

type (
	WatchSubscription struct {
		Subscribe          bool                       // Subscribe is true for subscribe, false for unsubscribe
		WatchCmd           cmd.RedisCmd               // WatchCmd Represents a unique key for each watch artifact, only populated for subscriptions.
		Fingerprint        uint32                     // Fingerprint is a unique identifier for each watch artifact, only populated for unsubscriptions.
		ClientFD           int                        // ClientFD is the file descriptor of the client connection
		CmdWatchClientChan chan comm.CmdWatchResponse // CmdWatchClientChan is the generic channel for HTTP/Websockets etc.
		ClientIdentifierID uint32                     // ClientIdentifierID Helps identify CmdWatch client on httpserver side
	}

	Manager struct {
		querySubscriptionMap map[string]map[uint32]bool                    // querySubscriptionMap is a map of Key -> [fingerprint1, fingerprint2, ...]
		tcpSubscriptionMap   map[uint32]map[clientio.ClientIdentifier]bool // tcpSubscriptionMap is a map of fingerprint -> [client1, client2, ...]
		fingerprintCmdMap    map[uint32]cmd.RedisCmd                       // fingerprintCmdMap is a map of fingerprint -> RedisCmd
		mu                   sync.RWMutex
		logger               *slog.Logger
	}
)

var (
	CmdWatchSubscriptionChan chan WatchSubscription
)

func NewManager(logger *slog.Logger) *Manager {
	CmdWatchSubscriptionChan = make(chan WatchSubscription)
	return &Manager{
		querySubscriptionMap: make(map[string]map[uint32]bool),
		tcpSubscriptionMap:   make(map[uint32]map[clientio.ClientIdentifier]bool),
		fingerprintCmdMap:    make(map[uint32]cmd.RedisCmd),
		logger:               logger,
	}
}

// Run starts the watch manager, listening for subscription requests and events
func (m *Manager) Run(ctx context.Context, eventChan chan dstore.CmdWatchEvent) {
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		m.listenForSubscriptions(ctx)
	}()

	go func() {
		defer wg.Done()
		m.listenForEvents(ctx, eventChan)
	}()

	wg.Wait()
}

// listenForSubscriptions handles incoming subscription requests
func (m *Manager) listenForSubscriptions(ctx context.Context) {
	for {
		select {
		case sub := <-CmdWatchSubscriptionChan:
			if sub.Subscribe {
				m.handleSubscription(sub)
			} else {
				m.handleUnsubscription(sub)
			}
		case <-ctx.Done():
			return
		}
	}
}

// handleSubscription processes a new subscription request
func (m *Manager) handleSubscription(sub WatchSubscription) {
	fingerprint := sub.WatchCmd.GetFingerprint()
	key := sub.WatchCmd.GetKey()

	client := clientio.NewClientIdentifier(sub.ClientFD, false)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add fingerprint to querySubscriptionMap
	if m.querySubscriptionMap[key] == nil {
		m.querySubscriptionMap[key] = make(map[uint32]bool)
	}
	m.querySubscriptionMap[key][fingerprint] = true

	// Add RedisCmd to fingerprintCmdMap
	m.fingerprintCmdMap[fingerprint] = sub.WatchCmd

	// Add clientID to tcpSubscriptionMap
	if m.tcpSubscriptionMap[fingerprint] == nil {
		m.tcpSubscriptionMap[fingerprint] = make(map[clientio.ClientIdentifier]bool)
	}
	m.tcpSubscriptionMap[fingerprint][client] = true
}

// handleUnsubscription processes an unsubscription request
func (m *Manager) handleUnsubscription(sub WatchSubscription) {
	fingerprint := sub.Fingerprint
	client := clientio.NewClientIdentifier(sub.ClientFD, false)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove clientID from tcpSubscriptionMap
	if clients, ok := m.tcpSubscriptionMap[fingerprint]; ok {
		delete(clients, client)
		// If there are no more clients listening to this fingerprint, remove it from the map
		if len(clients) == 0 {
			// Remove the fingerprint from tcpSubscriptionMap
			delete(m.tcpSubscriptionMap, fingerprint)
			// Also remove the fingerprint from fingerprintCmdMap
			delete(m.fingerprintCmdMap, fingerprint)
		} else {
			// Update the map with the new set of clients
			m.tcpSubscriptionMap[fingerprint] = clients
		}
	}

	// Remove fingerprint from querySubscriptionMap
	if redisCmd, ok := m.fingerprintCmdMap[fingerprint]; ok {
		key := redisCmd.GetKey()
		if fingerprints, ok := m.querySubscriptionMap[key]; ok {
			// Remove the fingerprint from the list of fingerprints listening to this key
			delete(fingerprints, fingerprint)
			// If there are no more fingerprints listening to this key, remove it from the map
			if len(fingerprints) == 0 {
				delete(m.querySubscriptionMap, key)
			} else {
				// Update the map with the new set of fingerprints
				m.querySubscriptionMap[key] = fingerprints
			}
		}
	}
}

func (m *Manager) listenForEvents(ctx context.Context, eventChan chan dstore.CmdWatchEvent) {
	affectedCmdMap := map[string]map[string]bool{"SET": {"GET": true}}
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eventChan:
			m.mu.RLock()

			// Check if any watch commands are listening to updates on this key.
			if _, ok := m.querySubscriptionMap[event.AffectedKey]; ok {
				// iterate through all command fingerprints that are listening to this key
				for fingerprint := range m.querySubscriptionMap[event.AffectedKey] {
					// Check if the command associated with this fingerprint actually needs to be executed for this event.
					// For instance, if the event is a SET, only execute GET commands need to be executed. This also
					// helps us handle cases where a key might get updated by an unrelated command which makes it
					// incompatible with the watched command.
					if affectedCommands, ok := affectedCmdMap[event.Cmd]; ok {
						if _, ok := affectedCommands[m.fingerprintCmdMap[fingerprint].Cmd]; ok {
							// TODO: execute the command, store the result, send to clients
							if clients, ok := m.tcpSubscriptionMap[fingerprint]; ok {
								for client := range clients {
									notifyClient(client, result)
								}
							}
						}
					} else {
						m.logger.Error("Received a watch event for an unknown command type",
							slog.String("cmd", event.Cmd))
					}
				}
			}

			m.mu.RUnlock()
		}
	}
}
