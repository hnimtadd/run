package actrs

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hnimtadd/run/internal/runtime"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/pb/v1"

	"github.com/asynkron/protoactor-go/actor"
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
	deploy, err := r.store.GetDeployment(msg.DeploymentId)
	if err != nil {
		return fmt.Errorf("runtime: could not find deployment (%s)", r.deploymentID)
	}

	r.deploymentID = deploy.ID
	modCache, err := r.cache.Get(deploy.ID)
	if err != nil {
		fmt.Println("cache not hit")
		modCache = wazero.NewCompilationCache()
	}

	args := runtime.Args{
		Cache:        modCache,
		DeploymentID: deploy.ID,
		Engine:       msg.Runtime,
		Stdout:       r.stdout,
	}

	args.Blob = deploy.Blob

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

	// TODO: runtime should return status, http response, log instead of http response only
	rsp := new(pb.HTTPResponse)
	if err := proto.Unmarshal(r.stdout.Bytes(), rsp); err != nil {
		responseError(ctx, http.StatusInternalServerError, "invoke error: "+err.Error(), msg.Id)
		return
	}

	// TODO: in the future, we should track the runtime metric of the request (duration, status)
	responseHTTPResponse(ctx, rsp)
	r.stdout.Reset()
}

func responseHTTPResponse(ctx actor.Context, response *pb.HTTPResponse) {
	ctx.Send(ctx.Parent(), response)
}

func responseError(ctx actor.Context, code int32, msg string, id string) {
	ctx.Send(ctx.Parent(), &pb.HTTPResponse{
		Body:      []byte(msg),
		Code:      code,
		RequestId: id,
	})
}

func NewRuntime(store store.Store, cache store.ModCacher) actor.Producer {
	return func() actor.Actor {
		return &Runtime{
			store: store,
			cache: cache,
		}
	}
}
