package runtime_test

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/hnimtadd/run/internal/runtime"
	"github.com/hnimtadd/run/internal/shared"
	"github.com/hnimtadd/run/pb/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"google.golang.org/protobuf/proto"
)

func TestRuntime_InvokeGoCode(t *testing.T) {
	b, err := os.ReadFile("./../_testdata/helloworld.wasm")
	require.Nil(t, err)

	req := &pb.HTTPRequest{
		Method: "get",
		Url:    "/",
		Body:   nil,
	}
	breq, err := proto.Marshal(req)
	require.Nil(t, err)

	out := &bytes.Buffer{}
	args := runtime.Args{
		Stdout:       out,
		DeploymentID: uuid.New(),
		Blob:         b,
		Engine:       "go",
		Cache:        wazero.NewCompilationCache(),
	}
	r, err := runtime.New(context.Background(), args)
	require.Nil(t, err)
	require.Nil(t, r.Invoke(bytes.NewReader(breq), nil))
	status, res, err := shared.ParseStdout(out)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "Hello world!", string(res))
	require.Nil(t, r.Close())
}

func TestRuntime_ModifierPassedInModuleCache(t *testing.T) {
	b, err := os.ReadFile("./../../examples/go/example.wasm")
	require.Nil(t, err)
	require.NotNil(t, b)
	out := &bytes.Buffer{}
	modCache := wazero.NewCompilationCache()
	args := runtime.Args{
		Stdout:       out,
		DeploymentID: uuid.New(),
		Blob:         b,
		Engine:       "go",
		Cache:        modCache,
	}
	r, err := runtime.New(context.Background(), args)
	require.Nil(t, err)
	require.NotNil(t, r)
	sout := new(bytes.Buffer)
	secondArgs := runtime.Args{
		Stdout:       sout,
		DeploymentID: uuid.New(),
		Blob:         b,
		Engine:       "go",
		Cache:        modCache,
	}
	sr, err := runtime.New(context.Background(), secondArgs)
	require.Nil(t, err)
	require.NotNil(t, sr)
}
