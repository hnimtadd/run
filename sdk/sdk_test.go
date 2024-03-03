package sdk

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSdkHandle(t *testing.T) {
	r, w, err := os.Pipe()
	require.Nil(t, err)
	os.Stdout = w
	Handle(http.HandlerFunc(handleFunc))

	require.NotNil(t, r)
	fmt.Println(r)
}

func handleFunc(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hello world!")); err != nil {
		fmt.Println(err)
	}
}
