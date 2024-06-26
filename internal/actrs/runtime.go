package actrs

import (
	"bytes"
	"context"
	"encoding/json"
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
	pb "github.com/hnimtadd/run/pbs/gopb/v1"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
	"google.golang.org/protobuf/proto"
)

type Runtime struct {
	Started     time.Time
	Store       store.Store
	LogStore    store.LogStore
	BlobStore   store.BlobStore
	MetricStore store.MetricStore
	Cache       store.ModCacher
	Runtime     *runtime.Runtime
	StdOut      *bytes.Buffer
	ManagerPID  *actor.PID
	Deployment  uuid.UUID
	_format     types.LogFormat
}

func (r *Runtime) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		slog.Info("runtime started", "node", "runtime")
		r.Started = time.Now()

	case *actor.Stopped:
		timeUsed := time.Since(r.Started)
		slog.Info("runtime started", "node", "runtime", "online duration", timeUsed)

	case *pb.HTTPRequest:
		slog.Info("incoming request", "request", msg.Id)
		if r.Runtime == nil {
			if err := r.Initialize(msg); err != nil {
				slog.Info("cannot initialized runtime", "node", "runtime", "msg", err.Error())
			}
		}
		r.ManagerPID = ctx.Sender()
		// Handle the HTTP request that is forwarded from the WASM server actor.
		r.Handle(ctx, msg)
	}
}

func (r *Runtime) Initialize(msg *pb.HTTPRequest) error {
	deploy, err := r.Store.GetDeploymentByID(msg.DeploymentId)
	if err != nil {
		slog.Error("runtime: could not find deployment ", "msg", err.Error())
		return err
	}

	blobMetadata, err := r.Store.GetBlobMetadataByDeploymentID(msg.DeploymentId)
	if err != nil {
		slog.Error("cannot get blob information  from store", "msg", err.Error())
		return err
	}

	blob, err := r.BlobStore.GetDeploymentBlobByURI(blobMetadata.Location)
	if err != nil {
		slog.Error("cannot get deployment blob from blobStore", "msg", err.Error())
		return err
	}

	r.Deployment = deploy.ID
	r._format = deploy.Format
	modCache, err := r.Cache.Get(deploy.ID)
	if err != nil {
		modCache = wazero.NewCompilationCache()
	}
	r.StdOut = new(bytes.Buffer)

	args := runtime.Args{
		Stdout:       r.StdOut,
		DeploymentID: deploy.ID,
		Blob:         blob.Data,
		Engine:       msg.Runtime,
		Cache:        modCache,
	}

	run, err := runtime.New(context.Background(), args)
	if err != nil {
		slog.Error("failed to create runtime", "msg", err.Error())
		return err
	}

	r.Runtime = run

	err = r.Cache.Put(deploy.ID, modCache)
	if err != nil {
		log.Println("cannot put cache", err)
	}
	return nil
}

func (r *Runtime) Handle(ctx actor.Context, req *pb.HTTPRequest) {
	if req.Runtime != r.Runtime.GetRuntime() {
		slog.Error("invalid runtime found in request", "node", "runtime")
		responseError(ctx, req, http.StatusBadRequest, "invalid runtime found in request", req.Id)
		return
	}
	switch req.Runtime {
	case "python":
		r.HandlePythonRuntime(ctx, req)
		return
	case "go":
		r.HandleGoRuntime(ctx, req)
		return
	default:
		responseError(ctx, req, http.StatusBadRequest, fmt.Sprintf("this runtime (%s) is not support", req.Runtime), req.Id)
		return
	}
}

func (r *Runtime) HandleGoRuntime(ctx actor.Context, req *pb.HTTPRequest) {
	if r.Deployment != uuid.MustParse(req.DeploymentId) {
		responseError(ctx, req, http.StatusInternalServerError, "deploymentID must match with runtime deployment ID", req.Id)
		return
	}

	start := time.Now()
	bufBytes, err := proto.Marshal(req)
	if err != nil {
		responseError(ctx, req, http.StatusInternalServerError, "cannot marshal request", req.Id)
		return
	}

	if err := r.Runtime.Invoke(bytes.NewReader(bufBytes), req.GetEnv()); err != nil {
		// request_log.go error
		responseError(ctx, req, http.StatusInternalServerError, "invoke error: "+err.Error(), req.Id)
		return
	}

	logs, body, err := shared.ParseStdout(r.StdOut)
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
		reqLogs := types.NewRequestLog(r.Deployment, requestUID, lines)
		if err := r.LogStore.AppendLog(reqLogs); err != nil {
			slog.Error("failed to add log to server", "request", req.Id, "msg", err.Error())
		}
	}
	duration := time.Since(start)
	// Calculate metric for current request
	requestMetric := types.CreateRequestMetric(req.Id, int(rsp.Code), duration)

	// update metric of this deployment

	responseHTTPWithMetrics(ctx, req, rsp, &requestMetric)
	r.StdOut.Reset()
}

func (r *Runtime) HandlePythonRuntime(ctx actor.Context, req *pb.HTTPRequest) {
	// currently, we could not use protobuf with python sdk, so this handlers try to parse the request into json object, then pass it into the sandbox, the response then will be used to construct the proto response
	if r.Deployment != uuid.MustParse(req.DeploymentId) {
		slog.Info("deploymentID mismatch", "node", "runtime")
		responseError(ctx, req, http.StatusInternalServerError, "deploymentID must match with runtime deployment ID", req.Id)
		return
	}

	start := time.Now()

	// TODO: fix this json, currently we directly parse it into json
	jsonReq := map[string]any{
		"body":          req.GetBody(),
		"method":        req.GetMethod(),
		"url":           req.GetUrl(),
		"endpoint_id":   req.GetEndpointId(),
		"env":           req.GetEnv(),
		"header":        req.GetHeader(),
		"runtime":       req.GetRuntime(),
		"deployment_id": req.GetDeploymentId(),
		"id":            req.GetId(),
	}
	slog.Info("req", "req", jsonReq, "node", "runtime")

	bufBytes, err := json.Marshal(jsonReq)
	// bufBytes, err := proto.Marshal(req)
	if err != nil {
		slog.Info("cannot marshal request", "node", "runtime", "msg", err.Error())
		responseError(ctx, req, http.StatusInternalServerError, "cannot marshal request", req.Id)
		return
	}

	if err := r.Runtime.Invoke(bytes.NewReader(bufBytes), req.GetEnv()); err != nil {
		slog.Info("invoke error", "msg", err.Error(), "node", "runtime")
		// request_log.go error
		responseError(ctx, req, http.StatusInternalServerError, "invoke error: "+err.Error(), req.Id)
		return
	}

	logs, body, err := shared.ParseStdout(r.StdOut)
	if err != nil {
		slog.Error("cannot parse output ", "request", req.Id, "msg", err.Error())
		responseError(ctx, req, http.StatusInternalServerError, "cannot parse output "+err.Error(), req.Id)
		return
	}

	type sandboxResponse struct {
		Body      string              `json:"body"`
		Code      int                 `json:"code"`
		RequestID string              `json:"request_id"`
		Header    map[string][]string `json:"header"`
	}

	jsonRes := new(sandboxResponse)

	if err := json.Unmarshal(body, jsonRes); err != nil {
		slog.Error("cannot unmarshal output ", "request", req.Id, "msg", err.Error())
		responseError(ctx, req, http.StatusInternalServerError, "cannot unmarshal output "+err.Error(), req.Id)
		return
	}

	protoHeaders := map[string]*pb.HeaderFields{}
	for key, header := range jsonRes.Header {
		protoHeaders[key] = &pb.HeaderFields{
			Fields: header,
		}
	}

	rsp := &pb.HTTPResponse{
		Body:      []byte(jsonRes.Body),
		Code:      int32(jsonRes.Code),
		RequestId: req.Id,
		Header:    protoHeaders,
	}

	// TODO: runtime metrics, write request_log.go to metric server
	lines, err := shared.ParseLog(logs)
	if err == nil {
		requestUID, _ := uuid.Parse(req.Id)
		reqLogs := types.NewRequestLog(r.Deployment, requestUID, lines)
		if err := r.LogStore.AppendLog(reqLogs); err != nil {
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
	r.StdOut.Reset()
	fmt.Println(rsp)
}

func responseHTTPWithMetrics(ctx actor.Context, request *pb.HTTPRequest, response *pb.HTTPResponse, metric *types.RequestMetric) {
	if ctx == nil {
		return
	}
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
			Store:     cfg.Store,
			Cache:     cfg.Cache,
			LogStore:  cfg.LogStore,
			BlobStore: cfg.BlobStore,
		}
	}
}

var KindRuntime = "kind-runtime"

type RuntimeConfig struct {
	Store     store.Store
	LogStore  store.LogStore
	BlobStore store.BlobStore
	Cache     store.ModCacher
}

func NewRuntimeKind(cfg *RuntimeConfig, opts ...actor.PropsOption) *cluster.Kind {
	props := actor.PropsFromProducer(NewRuntime(cfg), opts...)
	return cluster.NewKind(KindRuntime, props)
}
