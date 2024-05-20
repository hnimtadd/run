package runtime_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hnimtadd/run/internal/runtime"
	"github.com/hnimtadd/run/internal/shared"
	pb "github.com/hnimtadd/run/pbs/gopb/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"google.golang.org/protobuf/proto"
)

func TestRuntime_InvokeGoCode(t *testing.T) {
	b, err := os.ReadFile("./../_testdata/go/helloworld.wasm")
	require.Nil(t, err)

	req := &pb.HTTPRequest{
		Method: "GET",
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

	log, body, err := shared.ParseStdout(out)
	require.Nil(t, err)

	rsp := new(pb.HTTPResponse)
	require.Nil(t, proto.Unmarshal(body, rsp))
	require.NotNil(t, rsp)

	require.Equal(t, http.StatusOK, int(rsp.Code))
	require.Equal(t, "Hello world!", string(rsp.Body))
	require.Nil(t, r.Close())
	lines, err := shared.ParseLog(log)
	fmt.Println(lines)
	require.Nil(t, err)
	require.Equal(t, 1, len(lines))
	require.Equal(t, lines[0], "hello, this is a request_log.go")
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

func TestRuntime_InvokeGoCodeExample(t *testing.T) {
	b, err := os.ReadFile("./../../examples/go/example.wasm")
	require.Nil(t, err)

	req := &pb.HTTPRequest{
		Method: "GET",
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

	log, body, err := shared.ParseStdout(out)
	require.Nil(t, err)
	rsp := new(pb.HTTPResponse)
	require.Nil(t, proto.Unmarshal(body, rsp))

	require.Equal(t, http.StatusOK, int(rsp.Code))
	require.Equal(t, "<html>login page: <a href=\"/login\" /><br />Dashboard page: <a href=\"/dashboard\" /></html>", string(rsp.Body))
	require.Nil(t, r.Close())
	lines, err := shared.ParseLog(log)
	fmt.Println(lines)
	require.Nil(t, err)
	require.Equal(t, 1, len(lines))
	require.Equal(t, lines[0], "enter index")
}
