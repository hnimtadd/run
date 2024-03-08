package actrs

import (
	"github.com/hnimtadd/run/internal/message"
	"github.com/hnimtadd/run/internal/store"

	"github.com/asynkron/protoactor-go/actor"
)

type RuntimeManager struct {
	store  store.Store
	cache  store.ModCacher
	lookup map[string]*actor.PID // map pid with runtime
}

func (r *RuntimeManager) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		return
	case *actor.Stop:
		return
	case *actor.Stopping:
		return
	case *actor.Stopped:
		return
	case *message.RequestRuntimeMessage:
		pid, ok := r.lookup[msg.DeploymentID]
		if !ok {
			pid = r.SpawnRuntime(ctx)
			r.lookup[msg.DeploymentID] = pid
		}
		ctx.Respond(pid)
	}
}

func (r *RuntimeManager) SpawnRuntime(ctx actor.Context) *actor.PID {
	props := actor.PropsFromProducer(NewRuntime(r.store, r.cache))
	pid := ctx.Spawn(props)
	return pid
}

func NewRuntimeManager(store store.Store, cache store.ModCacher) actor.Producer {
	return func() actor.Actor {
		return &RuntimeManager{
			lookup: make(map[string]*actor.PID),
			store:  store,
			cache:  cache,
		}
	}
}
