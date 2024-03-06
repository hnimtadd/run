package actrs

import (
	"github.com/asynkron/protoactor-go/actor"
	"github.com/hnimtadd/run/internal/message"
)

var KindRuntimeManager = "runtime-manager"

type RuntimeManager struct {
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
			pid = r.SpawnRuntime(ctx, msg)
		}
		ctx.Respond(pid)
	}
}

func (r *RuntimeManager) SpawnRuntime(ctx actor.Context, msg *message.RequestRuntimeMessage) *actor.PID {
	props := actor.PropsFromProducer(NewRuntime)
	pid := ctx.Spawn(props)
	ctx.Request(pid, msg)
	return pid
}

func NewRuntimeManager() actor.Actor {
	return &RuntimeManager{
		lookup: make(map[string]*actor.PID),
	}
}
