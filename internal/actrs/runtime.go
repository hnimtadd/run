package actrs

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hnimtadd/run/internal/runtime"
	"github.com/hnimtadd/run/internal/shared"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/pb/v1"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
	"google.golang.org/protobuf/proto"
)

type Runtime struct {
	started      time.Time
	store        store.Store
	cache        store.ModCacher
	runtime      *runtime.Runtime
	stdout       *bytes.Buffer
	managerPID   *actor.PID
	deploymentID uuid.UUID
}

func (r *Runtime) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		fmt.Print("booted runtime at", ctx.Self().Id)
		r.started = time.Now()
	case *actor.Stopped:
		timeUsed := time.Since(r.started)
		fmt.Println("stopped runtime", timeUsed.Seconds())
	case *pb.HTTPRequest:
		fmt.Println("runtime handling request", "request_id", msg.Id)
		if r.runtime == nil {
			if err := r.Initialize(msg); err != nil {
				fmt.Println("cannot initialized runtime", err)
			}
		}
		r.managerPID = ctx.Sender()
		// Handle the HTTP request that is forwarded from the WASM server actor.
		r.Handle(ctx, msg)
	}
}

func (r *Runtime) Initialize(msg *pb.HTTPRequest) error {
	deploy, err := r.store.GetDeploymentByID(msg.DeploymentId)
	if err != nil {
		return fmt.Errorf("runtime: could not find deployment (%s)", r.deploymentID)
	}

	r.deploymentID = deploy.ID
	modCache, err := r.cache.Get(deploy.ID)
	if err != nil {
		fmt.Println("cache not hit")
		modCache = wazero.NewCompilationCache()
	}

	r.stdout = new(bytes.Buffer)
	args := runtime.Args{
		Cache:        modCache,
		DeploymentID: deploy.ID,
		Engine:       msg.Runtime,
		Stdout:       r.stdout,
		Blob:         deploy.Blob,
	}

	run, err := runtime.New(context.Background(), args)
	if err != nil {
		return err
	}
	r.runtime = run

	go func() {
		_ = r.cache.Put(deploy.ID, modCache)
	}()
	return nil
}

func (r *Runtime) Handle(ctx actor.Context, msg *pb.HTTPRequest) {
	if r.deploymentID != uuid.MustParse(msg.DeploymentId) {
		responseError(ctx, http.StatusInternalServerError, "deploymentID must match with runtime deployment ID", msg.Id)
		return
	}

	bufBytes, err := proto.Marshal(msg)
	if err != nil {
		responseError(ctx, http.StatusInternalServerError, "cannot marshal request", msg.Id)
		return
	}

	// TODO: clean this
	var args []string

	req := bytes.NewReader(bufBytes)
	if err := r.runtime.Invoke(req, msg.GetEnv(), args...); err != nil {
		// log error
		responseError(ctx, http.StatusInternalServerError, "invoke error: "+err.Error(), msg.Id)
		return
	}

	// TODO: return logs from sandbox
	status, body, err := shared.ParseStdout(r.stdout)
	if err != nil {
		responseError(ctx, http.StatusInternalServerError, "cannot parse output"+err.Error(), msg.Id)
		return
	}
	rsp := &pb.HTTPResponse{
		Body:      body,
		Code:      int32(status),
		RequestId: msg.Id,
	}

	// TODO: in the future, we should track the runtime metric of the request (duration, status)
	responseHTTPResponse(ctx, rsp)
	r.stdout.Reset()
}

func responseHTTPResponse(ctx actor.Context, response *pb.HTTPResponse) {
	ctx.Respond(response)
}

func responseError(ctx actor.Context, code int32, msg string, id string) {
	ctx.Respond(&pb.HTTPResponse{
		Body:      []byte(msg),
		Code:      code,
		RequestId: id,
	})
}

func NewRuntime(cfg *RuntimeConfig) actor.Producer {
	return func() actor.Actor {
		return &Runtime{
			store: cfg.Store,
			cache: cfg.Cache,
		}
	}
}

var KindRuntime = "kind-runtime"

type RuntimeConfig struct {
	Store store.Store
	Cache store.ModCacher
}

func NewRuntimeKind(cfg *RuntimeConfig, opts ...actor.PropsOption) *cluster.Kind {
	props := actor.PropsFromProducer(NewRuntime(cfg), opts...)
	return cluster.NewKind(KindRuntime, props)
}
