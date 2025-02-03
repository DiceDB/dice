package ironhawk

import (
	"log/slog"
	"strconv"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/iothread"
)

var (
	keyFPMap    map[string]map[uint32]bool             // keyFPMap is a map of Key -> [fingerprint1, fingerprint2, ...]
	fpThreadMap map[uint32]map[*iothread.IOThread]bool // fpConnMap is a map of fingerprint -> [client1Chan, client2Chan, ...]
)

func init() {
	keyFPMap = make(map[string]map[uint32]bool)
	fpThreadMap = make(map[uint32]map[*iothread.IOThread]bool)
}

func HandleWatch(c *cmd.Cmd, t *iothread.IOThread) {
	fp := c.GetFingerprint()
	key := c.Key()
	slog.Debug("creating a new subscription",
		slog.String("key", key),
		slog.String("cmd", c.String()),
		slog.Any("fingerprint", fp))

	if _, ok := keyFPMap[key]; !ok {
		keyFPMap[key] = make(map[uint32]bool)
	}
	keyFPMap[key][fp] = true

	if _, ok := fpThreadMap[fp]; !ok {
		fpThreadMap[fp] = make(map[*iothread.IOThread]bool)
	}
	fpThreadMap[fp][t] = true
}

func HandleUnwatch(c *cmd.Cmd, t *iothread.IOThread) {
	if len(c.C.Args) != 1 {
		return
	}

	_fp, err := strconv.ParseUint(c.C.Args[0], 10, 32)
	if err != nil {
		return
	}
	fp := uint32(_fp)

	delete(fpThreadMap[fp], t)
	if len(fpThreadMap[fp]) == 0 {
		delete(fpThreadMap, fp)
	}

	for key, fpMap := range keyFPMap {
		if _, ok := fpMap[fp]; ok {
			delete(keyFPMap[key], fp)
		}
		if len(keyFPMap[key]) == 0 {
			delete(keyFPMap, key)
		}
	}
}

func CleanupThreadWatchSubscriptions(t *iothread.IOThread) {
	for fp, threadMap := range fpThreadMap {
		if _, ok := threadMap[t]; ok {
			delete(fpThreadMap[fp], t)
		}
		if len(fpThreadMap[fp]) == 0 {
			delete(fpThreadMap, fp)
		}
	}
}

// func (m *Manager) handleWatchEvent(event dstore.CmdWatchEvent) {
// 	// Check if any watch commands are listening to updates on this key.
// 	fingerprints, exists := m.querySubscriptionMap[event.AffectedKey]
// 	if !exists {
// 		return
// 	}

// 	affectedCommands, cmdExists := affectedCmdMap[event.Cmd]
// 	if !cmdExists {
// 		slog.Error("Received a watch event for an unknown command type",
// 			slog.String("cmd", event.Cmd))
// 		return
// 	}

// 	// iterate through all command fingerprints that are listening to this key
// 	for fingerprint := range fingerprints {
// 		cmdToExecute := m.fingerprintCmdMap[fingerprint]
// 		// Check if the command associated with this fingerprint actually needs to be executed for this event.
// 		// For instance, if the event is a SET, only GET commands need to be executed. This also
// 		// helps us handle cases where a key might get updated by an unrelated command which makes it
// 		// incompatible with the watched command.
// 		if _, affected := affectedCommands[cmdToExecute.Cmd]; affected {
// 			m.notifyClients(fingerprint, cmdToExecute)
// 		}
// 	}
// }

// // notifyClients sends cmd to all clients listening to this fingerprint, so that they can execute it.
// func (m *Manager) notifyClients(fingerprint uint32, diceDBCmd *cmd.DiceDBCmd) {
// 	clients, exists := m.tcpSubscriptionMap[fingerprint]
// 	if !exists {
// 		slog.Warn("No clients found for fingerprint",
// 			slog.Uint64("fingerprint", uint64(fingerprint)))
// 		return
// 	}

// 	for clientChan := range clients {
// 		clientChan <- diceDBCmd
// 	}
// }
