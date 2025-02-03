package ironhawk

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/iothread"
)

var (
	keyFPMap    map[string]map[uint32]bool             // keyFPMap is a map of Key -> [fingerprint1, fingerprint2, ...]
	fpThreadMap map[uint32]map[*iothread.IOThread]bool // fpConnMap is a map of fingerprint -> [client1Chan, client2Chan, ...]
	fpCmdMap    map[uint32]*cmd.Cmd                    // fpCmdMap is a map of fingerprint -> command
)

func init() {
	keyFPMap = make(map[string]map[uint32]bool)
	fpThreadMap = make(map[uint32]map[*iothread.IOThread]bool)
	fpCmdMap = make(map[uint32]*cmd.Cmd)
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
	fpCmdMap[fp] = c
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

	// TODO: Maintain ref count for gp -> cmd mapping
	// delete it from delete(fpCmdMap, fp) only when ref count is 0
	// check if any easier way to do this
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

func NotifyWatchers(c *cmd.Cmd, execute func(c *cmd.Cmd) (*cmd.CmdRes, error)) {
	// TODO: During first WATCH call, we are getting the response multiple times on the Client
	// Check if this is happening because of the way we are notifying the watchers
	key := c.Key()
	for fp := range keyFPMap[key] {
		_c := fpCmdMap[fp]
		if _c == nil {
			// TODO: We might want to remove the key from keyFPMap if we don't have a command for it.
			continue
		}

		r, err := execute(_c)
		if err != nil {
			slog.Error("failed to execute command as part of watch notification",
				slog.Any("cmd", _c.String()),
				slog.Any("error", err))
			continue
		}

		for thread := range fpThreadMap[fp] {
			err := thread.IoHandler.WriteSync(context.Background(), r)
			if err != nil {
				slog.Error("failed to write response to thread", slog.Any("thread", thread.ID()), slog.Any("error", err))
			}
		}

		slog.Debug("notifying watchers for key", slog.String("key", key), slog.Int("watchers", len(fpThreadMap[fp])))
	}
}
