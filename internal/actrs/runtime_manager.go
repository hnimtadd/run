package actrs

import (
	"fmt"

	"github.com/hnimtadd/run/internal/message"
	"github.com/hnimtadd/run/internal/store"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
)

type RuntimeManager struct {
	store  store.Store
	cache  store.ModCacher
	ctx    cluster.GrainContext
	lookup map[string]*actor.PID // map pid with runtime
}

func (r *RuntimeManager) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *cluster.ClusterInit:
		fmt.Println("start runtime manager")
		r.ctx = cluster.NewGrainContext(ctx, msg.Identity, msg.Cluster)
	case *actor.Stop:
		return
	case *actor.Stopping:
		return
	case *actor.Stopped:
		return
	case *message.RequestRuntimeMessage:
		fmt.Println("receive request runtime")
		pid, ok := r.lookup[msg.DeploymentID]
		if !ok {
			pid = r.SpawnRuntime()
			r.lookup[msg.DeploymentID] = pid
		}
		fmt.Println("will return runtime with pid", pid.Id)
		ctx.Respond(pid)
	}
}

func (r *RuntimeManager) SpawnRuntime() *actor.PID {
	props := actor.PropsFromProducer(
		NewRuntime(&RuntimeConfig{r.store, r.cache}),
	)
	pid := r.ctx.Spawn(props)
	return pid
}

func NewRuntimeManager(cfg *RuntimeManagerConfig) actor.Producer {
	return func() actor.Actor {
		return &RuntimeManager{
			lookup: make(map[string]*actor.PID),
			store:  cfg.Store,
			cache:  cfg.Cache,
		}
	}
}

var KindRuntimeManager = "kind-runtime-manager"

type RuntimeManagerConfig struct {
	Store store.Store
	Cache store.ModCacher
}

func NewRuntimeManagerKind(cfg *RuntimeManagerConfig, opts ...actor.PropsOption) *cluster.Kind {
	props := actor.PropsFromProducer(NewRuntimeManager(cfg), opts...)
	return cluster.NewKind(KindRuntimeManager, props)
}
