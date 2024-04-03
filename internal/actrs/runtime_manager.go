package actrs

import (
	"fmt"
	"log/slog"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/hnimtadd/run/internal/message"
)

type RuntimeManager struct {
	ctx    cluster.GrainContext
	lookup map[string]*actor.PID // map pid with runtime
}

func (r *RuntimeManager) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *cluster.ClusterInit:
		slog.Info("receive cluster init message", "node", "runtime manager")
		r.ctx = cluster.NewGrainContext(ctx, msg.Identity, msg.Cluster)
	case *actor.Stop:
		return
	case *actor.Stopping:
		return
	case *actor.Stopped:
		return
	case *message.RequestRuntimeMessage:
		slog.Info("receive runtime request message", "runtime", msg.Runtime, "node", "runtime manager")
		pid, ok := r.lookup[msg.DeploymentID]
		if !ok {
			pid = r.SpawnRuntime(msg)
			r.lookup[msg.DeploymentID] = pid
		}
		ctx.Respond(pid)
	}
}

func (r *RuntimeManager) SpawnRuntime(msg *message.RequestRuntimeMessage) *actor.PID {
	pid := r.ctx.Cluster().Get(fmt.Sprintf("%s/%s", msg.Runtime, msg.DeploymentID), KindRuntime)
	return pid
}

func NewRuntimeManager() actor.Producer {
	return func() actor.Actor {
		return &RuntimeManager{
			lookup: make(map[string]*actor.PID),
		}
	}
}

var KindRuntimeManager = "kind-runtime-manager"

func NewRuntimeManagerKind(opts ...actor.PropsOption) *cluster.Kind {
	props := actor.PropsFromProducer(NewRuntimeManager(), opts...)
	return cluster.NewKind(KindRuntimeManager, props)
}
