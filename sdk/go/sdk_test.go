package sdk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_responseWriter_Header(t *testing.T) {
	w := newResponseWriter()
	require.NotNil(t, w.header)
	require.NotNil(t, w.Header())

	key := "key"
	val := "val"
	w.Header().Add(key, val)

	require.Contains(t, w.Header().Values(key), val)
	require.Equal(t, w.Header().Get(key), val)
}
