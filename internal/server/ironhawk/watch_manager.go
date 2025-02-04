package ironhawk

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/dicedb/dice/internal/cmd"
)

type WatchManager struct {
	keyFPMap    map[string]map[uint32]bool
	fpThreadMap map[uint32]map[*IOThread]bool
	fpCmdMap    map[uint32]*cmd.Cmd
}

func NewWatchManager() *WatchManager {
	return &WatchManager{
		keyFPMap:    map[string]map[uint32]bool{},
		fpThreadMap: map[uint32]map[*IOThread]bool{},
		fpCmdMap:    map[uint32]*cmd.Cmd{},
	}
}

func (w *WatchManager) HandleWatch(c *cmd.Cmd, t *IOThread) {
	fp := c.GetFingerprint()
	key := c.Key()
	slog.Debug("creating a new subscription",
		slog.String("key", key),
		slog.String("cmd", c.String()),
		slog.Any("fingerprint", fp))

	if _, ok := w.keyFPMap[key]; !ok {
		w.keyFPMap[key] = make(map[uint32]bool)
	}
	w.keyFPMap[key][fp] = true

	if _, ok := w.fpThreadMap[fp]; !ok {
		w.fpThreadMap[fp] = make(map[*IOThread]bool)
	}
	w.fpThreadMap[fp][t] = true
	w.fpCmdMap[fp] = c
}

func (w *WatchManager) HandleUnwatch(c *cmd.Cmd, t *IOThread) {
	if len(c.C.Args) != 1 {
		return
	}

	_fp, err := strconv.ParseUint(c.C.Args[0], 10, 32)
	if err != nil {
		return
	}
	fp := uint32(_fp)

	delete(w.fpThreadMap[fp], t)
	if len(w.fpThreadMap[fp]) == 0 {
		delete(w.fpThreadMap, fp)
	}

	for key, fpMap := range w.keyFPMap {
		if _, ok := fpMap[fp]; ok {
			delete(w.keyFPMap[key], fp)
		}
		if len(w.keyFPMap[key]) == 0 {
			delete(w.keyFPMap, key)
		}
	}

	// TODO: Maintain ref count for gp -> cmd mapping
	// delete it from delete(fpCmdMap, fp) only when ref count is 0
	// check if any easier way to do this
}

func (w *WatchManager) CleanupThreadWatchSubscriptions(t *IOThread) {
	for fp, threadMap := range w.fpThreadMap {
		if _, ok := threadMap[t]; ok {
			delete(w.fpThreadMap[fp], t)
		}
		if len(w.fpThreadMap[fp]) == 0 {
			delete(w.fpThreadMap, fp)
		}
	}
}

func (w *WatchManager) NotifyWatchers(c *cmd.Cmd, shardManager *ShardManager, t *IOThread) {
	// TODO: During first WATCH call, we are getting the response multiple times on the Client
	// Check if this is happening because of the way we are notifying the watchers
	key := c.Key()
	for fp := range w.keyFPMap[key] {
		_c := w.fpCmdMap[fp]
		if _c == nil {
			// TODO: We might want to remove the key from keyFPMap if we don't have a command for it.
			continue
		}

		r, err := shardManager.Execute(_c)
		if err != nil {
			slog.Error("failed to execute command as part of watch notification",
				slog.Any("cmd", _c.String()),
				slog.Any("error", err))
			continue
		}

		for thread := range w.fpThreadMap[fp] {
			err := thread.IoHandler.WriteSync(context.Background(), r.R)
			if err != nil {
				slog.Error("failed to write response to thread", slog.Any("client_id", thread.ClientID), slog.Any("error", err))
			}
		}

		slog.Debug("notifying watchers for key", slog.String("key", key), slog.Int("watchers", len(w.fpThreadMap[fp])))
	}
}
