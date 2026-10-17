package runtime

import (
	"context"
	"os"
)

// Actor identifies who performed a gated action.
type Actor string

const (
	ActorHuman = Actor("human")
)

// CurrentActor resolves CLI/MCP actor from env (APPLYDIFF_ACTOR) or defaults to human.
func CurrentActor() Actor {
	if v := os.Getenv("APPLYDIFF_ACTOR"); v != "" {
		return Actor(v)
	}
	return ActorHuman
}

func (a *App) recordAction(ctx context.Context, jobID, action string, actor Actor, detail string) {
	if a == nil || a.Store == nil || a.Store.TableMissing(ctx, "action_log") {
		return
	}
	_ = a.Store.LogAction(ctx, jobID, action, string(actor), detail)
}
