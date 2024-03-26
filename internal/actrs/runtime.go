package actrs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hnimtadd/run/internal/runtime"
	"github.com/hnimtadd/run/internal/shared"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"
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
	logStore     store.LogStore
	cache        store.ModCacher
	runtime      *runtime.Runtime
	stdout       *bytes.Buffer
	managerPID   *actor.PID
	deploymentID uuid.UUID
	_format      types.LogFormat
}

func (r *Runtime) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		fmt.Println("booted runtime at", ctx.Self().Id)
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
	r._format = deploy.Format
	modCache, err := r.cache.Get(deploy.ID)
	if err != nil {
		fmt.Println("cache not hit")
		modCache = wazero.NewCompilationCache()
	}
	r.stdout = new(bytes.Buffer)

	args := runtime.Args{
		Stdout:       r.stdout,
		DeploymentID: deploy.ID,
		Blob:         deploy.Blob,
		Engine:       msg.Runtime,
		Cache:        modCache,
	}

	run, err := runtime.New(context.Background(), args)
	if err != nil {
		return err
	}

	r.runtime = run

	err = r.cache.Put(deploy.ID, modCache)
	if err != nil {
		log.Println("cannot put cache", err)
	}
	return nil
}

func (r *Runtime) Handle(ctx actor.Context, req *pb.HTTPRequest) {
	if r.deploymentID != uuid.MustParse(req.DeploymentId) {
		responseError(ctx, http.StatusInternalServerError, "deploymentID must match with runtime deployment ID", req.Id)
		return
	}

	bufBytes, err := proto.Marshal(req)
	if err != nil {
		responseError(ctx, http.StatusInternalServerError, "cannot marshal request", req.Id)
		return
	}

	if err := r.runtime.Invoke(bytes.NewReader(bufBytes), req.GetEnv()); err != nil {
		// request_log.go error
		responseError(ctx, http.StatusInternalServerError, "invoke error: "+err.Error(), req.Id)
		return
	}

	logs, body, err := shared.ParseStdout(r.stdout)
	if err != nil {
		fmt.Println("cannot parse output "+err.Error(), req.Id)
		responseError(ctx, http.StatusInternalServerError, "cannot parse output "+err.Error(), req.Id)
		return
	}

	rsp := new(pb.HTTPResponse)
	if err := proto.Unmarshal(body, rsp); err != nil {
		fmt.Println("cannot unmarshal output "+err.Error(), req.Id)
		responseError(ctx, http.StatusInternalServerError, "cannot unmarshal output "+err.Error(), req.Id)
		return
	}
	rsp.RequestId = req.Id

	// TODO: runtime metrics, write request_log.go to metric server
	lines, err := shared.ParseLog(logs)
	if err == nil {
		reqLogs := &types.RequestLog{
			DeploymentID: r.deploymentID,
			RequestID:    uuid.MustParse(req.Id),
			Contents:     lines,
			CreatedAt:    time.Now().Unix(),
		}
		if err := r.logStore.AppendLog(reqLogs); err != nil {
			log.Println("cannot append request to server", err)
		}
		// should not response error here, should switch to string request_log.go
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
			store:    cfg.Store,
			cache:    cfg.Cache,
			logStore: cfg.LogStore,
		}
	}
}

var KindRuntime = "kind-runtime"

type RuntimeConfig struct {
	Store    store.Store
	LogStore store.LogStore
	Cache    store.ModCacher
}

func NewRuntimeKind(cfg *RuntimeConfig, opts ...actor.PropsOption) *cluster.Kind {
	props := actor.PropsFromProducer(NewRuntime(cfg), opts...)
	return cluster.NewKind(KindRuntime, props)
}
