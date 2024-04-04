package actrs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/hnimtadd/run/internal/message"
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
		slog.Info("runtime started", "node", "runtime")

		r.started = time.Now()

	case *actor.Stopped:
		timeUsed := time.Since(r.started)
		slog.Info("runtime started", "node", "runtime", "online duration", timeUsed)

	case *pb.HTTPRequest:
		slog.Info("incoming request", "request", msg.Id)
		if r.runtime == nil {
			if err := r.Initialize(msg); err != nil {
				slog.Info("cannot initialized runtime", "node", "runtime", "msg", err.Error())
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
		responseError(ctx, req, http.StatusInternalServerError, "deploymentID must match with runtime deployment ID", req.Id)
		return
	}

	start := time.Now()
	bufBytes, err := proto.Marshal(req)
	if err != nil {
		responseError(ctx, req, http.StatusInternalServerError, "cannot marshal request", req.Id)
		return
	}

	if err := r.runtime.Invoke(bytes.NewReader(bufBytes), req.GetEnv()); err != nil {
		// request_log.go error
		responseError(ctx, req, http.StatusInternalServerError, "invoke error: "+err.Error(), req.Id)
		return
	}

	logs, body, err := shared.ParseStdout(r.stdout)
	if err != nil {
		slog.Error("cannot parse output ", "request", req.Id, "msg", err.Error())
		responseError(ctx, req, http.StatusInternalServerError, "cannot parse output "+err.Error(), req.Id)
		return
	}

	rsp := new(pb.HTTPResponse)
	if err := proto.Unmarshal(body, rsp); err != nil {
		slog.Error("cannot unmarshal output ", "request", req.Id, "msg", err.Error())
		responseError(ctx, req, http.StatusInternalServerError, "cannot unmarshal output "+err.Error(), req.Id)
		return
	}
	rsp.RequestId = req.Id

	// TODO: runtime metrics, write request_log.go to metric server
	lines, err := shared.ParseLog(logs)
	if err == nil {
		requestUID, _ := uuid.Parse(req.Id)
		reqLogs := types.NewRequestLog(r.deploymentID, requestUID, lines)
		if err := r.logStore.AppendLog(reqLogs); err != nil {
			slog.Error("failed to add log to server", "request", req.Id, "msg", err.Error())
		}
	}
	duration := time.Since(start)
	// Calculate metric for current request
	requestMetric := types.CreateRequestMetric(req.Id, int(rsp.Code), duration)

	// save request metric to store
	fmt.Println(requestMetric)

	// update metric of this deployment

	responseHTTPWithMetrics(ctx, req, rsp, &requestMetric)
	r.stdout.Reset()
}

func responseHTTPWithMetrics(ctx actor.Context, request *pb.HTTPRequest, response *pb.HTTPResponse, metric *types.RequestMetric) {
	ctx.Respond(&message.ResponseWithMetric{
		Response: response,
		MetricMessage: &message.MetricMessage{
			DeploymentID: request.DeploymentId,
			RequestID:    request.Id,
			Metric:       *metric,
		},
	})
}

func responseError(ctx actor.Context, request *pb.HTTPRequest, code int32, msg string, id string) {
	rsp := &pb.HTTPResponse{
		Body:      []byte(msg),
		Code:      code,
		RequestId: id,
	}
	responseHTTPWithMetrics(ctx, request, rsp, nil)
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
