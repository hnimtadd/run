package manager

import (
	"bytes"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"

	"github.com/hnimtadd/run/internal/message"
	"github.com/hnimtadd/run/internal/runtime"
)

type Runtime struct {
	started      time.Time
	runtime      *runtime.Runtime
	stdout       *bytes.Buffer
	blob         []byte
	deploymentID uuid.UUID
}

func NewRuntime() *Runtime {
	return &Runtime{}
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

func (r *Runtime) Handle(msg *message.RequestMessage) error {
	return r.runtime.Invoke(msg.Body, msg.Env, msg.Args...)
}
