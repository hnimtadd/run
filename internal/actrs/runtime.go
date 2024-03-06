package actrs

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hnimtadd/run/internal/message"
	"github.com/hnimtadd/run/internal/runtime"
	"github.com/hnimtadd/run/pb/v1"

	"github.com/asynkron/protoactor-go/actor"
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

func (r *Runtime) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		r.started = time.Now()
	case *message.RequestRuntimeMessage:
		if err := r.Initialize(msg); err != nil {
			fmt.Printf("cannot init runtime, %v", err)
		}
	}
}

func (r *Runtime) Initialize(msg *message.RequestRuntimeMessage) error {
	r.deploymentID = uuid.MustParse(msg.DeploymentID)
	r.started = time.Now()

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

func NewRuntime() actor.Actor {
	return &Runtime{}
}
