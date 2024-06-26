package runtime

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type Args struct {
	Stdout       io.Writer
	Cache        wazero.CompilationCache
	Engine       string
	Blob         []byte
	DeploymentID uuid.UUID
}

type Runtime struct {
	ctx          context.Context
	mod          wazero.CompiledModule
	runtime      wazero.Runtime
	stdout       io.Writer
	engine       string
	blob         []byte
	deploymentID uuid.UUID
}

func New(ctx context.Context, args Args) (*Runtime, error) {
	config := wazero.NewRuntimeConfig().
		WithCompilationCache(args.Cache).
		WithCloseOnContextDone(true)
	r := wazero.NewRuntimeWithConfig(ctx, config)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	mod, err := r.CompileModule(ctx, args.Blob)
	if err != nil {
		return nil, fmt.Errorf("runtime: failed to compile module, err: %v", err)
	}

	return &Runtime{
		engine:       args.Engine,
		stdout:       args.Stdout,
		runtime:      r,
		mod:          mod,
		blob:         args.Blob,
		ctx:          ctx,
		deploymentID: args.DeploymentID,
	}, nil
}

func (r *Runtime) Invoke(stdin io.Reader, env map[string]string, args ...string) error {
	modConf := wazero.
		NewModuleConfig().
		WithStdin(stdin).
		WithStdout(r.stdout).
		WithArgs(args...)

	for key, value := range env {
		modConf = modConf.WithEnv(key, value)
	}

	_, err := r.runtime.InstantiateModule(r.ctx, r.mod, modConf)
	return err
}

func (r *Runtime) Close() error {
	return r.runtime.Close(r.ctx)
}

func (r *Runtime) GetRuntime() string {
	return r.engine
}
