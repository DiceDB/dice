package watchmanager

import (
	"context"
	"log/slog"
	"sync"

	"github.com/dicedb/dice/internal/cmd"
	dstore "github.com/dicedb/dice/internal/store"
)

type (
	WatchSubscription struct {
		Subscribe    bool                // Subscribe is true for subscribe, false for unsubscribe. Required.
		AdhocReqChan chan *cmd.DiceDBCmd // AdhocReqChan is the channel to send adhoc requests to the io-thread. Required.
		WatchCmd     *cmd.DiceDBCmd      // WatchCmd Represents a unique key for each watch artifact, only populated for subscriptions.
		Fingerprint  uint32              // Fingerprint is a unique identifier for each watch artifact, only populated for unsubscriptions.
	}

	Manager struct {
		querySubscriptionMap     map[string]map[uint32]struct{}              // querySubscriptionMap is a map of Key -> [fingerprint1, fingerprint2, ...]
		tcpSubscriptionMap       map[uint32]map[chan *cmd.DiceDBCmd]struct{} // tcpSubscriptionMap is a map of fingerprint -> [client1Chan, client2Chan, ...]
		fingerprintCmdMap        map[uint32]*cmd.DiceDBCmd                   // fingerprintCmdMap is a map of fingerprint -> DiceDBCmd
		cmdWatchSubscriptionChan chan WatchSubscription                      // cmdWatchSubscriptionChan is the channel to send/receive watch subscription requests.
		cmdWatchChan             chan dstore.CmdWatchEvent                   // cmdWatchChan is the channel to send/receive watch events.
	}
)

var (
	affectedCmdMap = map[string]map[string]struct{}{
		dstore.Set:     {dstore.Get: struct{}{}},
		dstore.Del:     {dstore.Get: struct{}{}},
		dstore.Rename:  {dstore.Get: struct{}{}},
		dstore.ZAdd:    {dstore.ZRange: struct{}{}},
		dstore.PFADD:   {dstore.PFCOUNT: struct{}{}},
		dstore.PFMERGE: {dstore.PFCOUNT: struct{}{}},
	}
)

func NewManager(cmdWatchSubscriptionChan chan WatchSubscription, cmdWatchChan chan dstore.CmdWatchEvent) *Manager {
	return &Manager{
		querySubscriptionMap:     make(map[string]map[uint32]struct{}),
		tcpSubscriptionMap:       make(map[uint32]map[chan *cmd.DiceDBCmd]struct{}),
		fingerprintCmdMap:        make(map[uint32]*cmd.DiceDBCmd),
		cmdWatchSubscriptionChan: cmdWatchSubscriptionChan,
		cmdWatchChan:             cmdWatchChan,
	}
}

// Run starts the watch manager, listening for subscription requests and events
func (m *Manager) Run(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		m.listenForEvents(ctx)
	}()

	<-ctx.Done()
	wg.Wait()
}

func (m *Manager) listenForEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case sub := <-m.cmdWatchSubscriptionChan:
			if sub.Subscribe {
				m.handleSubscription(sub)
			} else {
				m.handleUnsubscription(sub)
			}
		case watchEvent := <-m.cmdWatchChan:
			m.handleWatchEvent(watchEvent)
		}
	}
}

// handleSubscription processes a new subscription request
func (m *Manager) handleSubscription(sub WatchSubscription) {
	fingerprint := sub.WatchCmd.GetFingerprint()
	key := sub.WatchCmd.GetKey()

	// Add fingerprint to querySubscriptionMap
	if _, exists := m.querySubscriptionMap[key]; !exists {
		m.querySubscriptionMap[key] = make(map[uint32]struct{})
	}
	m.querySubscriptionMap[key][fingerprint] = struct{}{}

	// Add DiceDBCmd to fingerprintCmdMap
	m.fingerprintCmdMap[fingerprint] = sub.WatchCmd

	// Add client channel to tcpSubscriptionMap
	if _, exists := m.tcpSubscriptionMap[fingerprint]; !exists {
		m.tcpSubscriptionMap[fingerprint] = make(map[chan *cmd.DiceDBCmd]struct{})
	}
	m.tcpSubscriptionMap[fingerprint][sub.AdhocReqChan] = struct{}{}
}

// handleUnsubscription processes an unsubscription request
func (m *Manager) handleUnsubscription(sub WatchSubscription) {
	fingerprint := sub.Fingerprint

	// Remove clientID from tcpSubscriptionMap
	if clients, ok := m.tcpSubscriptionMap[fingerprint]; ok {
		delete(clients, sub.AdhocReqChan)
		// If there are no more clients listening to this fingerprint, remove it from the map
		if len(clients) == 0 {
			// Remove the fingerprint from tcpSubscriptionMap
			delete(m.tcpSubscriptionMap, fingerprint)
		} else {
			// Other clients still subscribed, no need to remove the fingerprint altogether
			return
		}
	}

	// Remove fingerprint from querySubscriptionMap
	if diceDBCmd, ok := m.fingerprintCmdMap[fingerprint]; ok {
		key := diceDBCmd.GetKey()
		if fingerprints, ok := m.querySubscriptionMap[key]; ok {
			// Remove the fingerprint from the list of fingerprints listening to this key
			delete(fingerprints, fingerprint)
			// If there are no more fingerprints listening to this key, remove it from the map
			if len(fingerprints) == 0 {
				delete(m.querySubscriptionMap, key)
			}
		}
		// Also remove the fingerprint from fingerprintCmdMap
		delete(m.fingerprintCmdMap, fingerprint)
	}
}

func (m *Manager) handleWatchEvent(event dstore.CmdWatchEvent) {
	// Check if any watch commands are listening to updates on this key.
	fingerprints, exists := m.querySubscriptionMap[event.AffectedKey]
	if !exists {
		return
	}

	affectedCommands, cmdExists := affectedCmdMap[event.Cmd]
	if !cmdExists {
		slog.Error("Received a watch event for an unknown command type",
			slog.String("cmd", event.Cmd))
		return
	}

	// iterate through all command fingerprints that are listening to this key
	for fingerprint := range fingerprints {
		cmdToExecute := m.fingerprintCmdMap[fingerprint]
		// Check if the command associated with this fingerprint actually needs to be executed for this event.
		// For instance, if the event is a SET, only GET commands need to be executed. This also
		// helps us handle cases where a key might get updated by an unrelated command which makes it
		// incompatible with the watched command.
		if _, affected := affectedCommands[cmdToExecute.Cmd]; affected {
			m.notifyClients(fingerprint, cmdToExecute)
		}
	}
}

// notifyClients sends cmd to all clients listening to this fingerprint, so that they can execute it.
func (m *Manager) notifyClients(fingerprint uint32, diceDBCmd *cmd.DiceDBCmd) {
	clients, exists := m.tcpSubscriptionMap[fingerprint]
	if !exists {
		slog.Warn("No clients found for fingerprint",
			slog.Uint64("fingerprint", uint64(fingerprint)))
		return
	}

	for clientChan := range clients {
		clientChan <- diceDBCmd
	}
}
