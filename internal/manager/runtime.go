package manager

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/hnimtadd/run/internal/message"
	"github.com/hnimtadd/run/internal/runtime"
	"github.com/hnimtadd/run/pb/v1"

	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
	"google.golang.org/protobuf/proto"
)

type Runtime struct {
	started time.Time
	runtime *runtime.Runtime
	stdout  *bytes.Buffer
	blob    []byte
	// store
	deploymentID uuid.UUID
}

func NewRuntime() *Runtime {
	return &Runtime{
		stdout: new(bytes.Buffer),
	}
}

func (r *Runtime) Initialize(msg *message.RequestRuntimeMessage) error {
	r.deploymentID = uuid.MustParse(msg.DeploymentID)
	r.started = time.Now()

	// Must check deploy and cache here
	// deploy, err := r.store.GetDeployment(r.deploymentID)
	// if err != nil {
	// 	return fmt.Errorf("runtime: could not find deployment (%s)", r.deploymentID)
	// }

	modCache := wazero.NewCompilationCache()

	args := runtime.Args{
		Cache:        modCache,
		DeploymentID: r.deploymentID,
		Engine:       msg.Runtime,
		Stdout:       r.stdout,
	}

	args.Blob = r.blob

	run, err := runtime.New(context.Background(), args)
	if err != nil {
		return err
	}
	r.runtime = run

	return nil
}

func (r *Runtime) Handle(ctx context.Context, msg *pb.HTTPRequest) {
	if r.deploymentID != uuid.MustParse(msg.DeploymentId) {
		responseError(ctx, http.StatusBadRequest, "deploymentID must match with runtime id", msg.Id)
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
	responseHTTPResponse(ctx, rsp)
	r.stdout.Reset()

	// TODO: in the future, we should track the runtime metric of the request (duration, status)
}

func responseHTTPResponse(ctx context.Context, response *pb.HTTPResponse) {
	responseCh, ok := ctx.Value(message.ContextKey(response.RequestId)).(chan<- *pb.HTTPResponse)
	if !ok {
		log.Panic("context key not found")
		return
	}
	responseCh <- response
}

func responseError(ctx context.Context, code int32, msg string, id string) {
	responseCh, ok := ctx.Value(message.ContextKey(id)).(chan<- *pb.HTTPResponse)
	if !ok {
		log.Panic("context key not found")
	}

	responseCh <- &pb.HTTPResponse{
		Body:      []byte(msg),
		Code:      code,
		RequestId: id,
	}
}
