package runtime

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hnimtadd/run/pb/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"google.golang.org/protobuf/proto"
)

func TestRuntimeInvokeGoCode(t *testing.T) {
	b, err := os.ReadFile("./../../examples/go/example.wasm")
	require.Nil(t, err)
	require.NotNil(t, b)

	req := &pb.HTTPRequest{
		Method: "get",
		Url:    "/",
		Body:   nil,
	}
	reqBytes, err := proto.Marshal(req)
	require.Nil(t, err)

	out := &bytes.Buffer{}
	args := Args{
		Stdout:       out,
		DeploymentID: uuid.New(),
		Blob:         b,
		Engine:       "go",
		Cache:        wazero.NewCompilationCache(),
	}
	r, err := New(context.Background(), args)
	require.Nil(t, err)
	require.Nil(t, r.Invoke(bytes.NewReader(reqBytes), nil))

	fmt.Println(out)
	require.NotZero(t, out.Len())
	require.Nil(t, r.Close())
}
