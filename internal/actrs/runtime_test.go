package actrs_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/hnimtadd/run/internal/actrs"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"
	pb "github.com/hnimtadd/run/pbs/gopb/v1"
	"github.com/stretchr/testify/require"
)

func TestRuntimeActrs(t *testing.T) {
	b, err := os.ReadFile("./../_testdata/python/index.wasm")
	require.Nil(t, err)
	memoryStore := store.NewMemoryStore()
	cacheStore := store.NewMemoryModCacher()
	endpoint, err := types.NewEndpoint("test_endpoint", "python", map[string]string{})
	require.Nil(t, err)
	require.Nil(t, memoryStore.CreateEndpoint(endpoint))

	deployment, err := types.NewDeployment(endpoint)
	require.Nil(t, err)
	require.Nil(t, memoryStore.CreateDeployment(deployment))

	blobMetadata, err := types.NewRawBlobMetadata(deployment, b)
	require.Nil(t, err)
	require.Nil(t, memoryStore.CreateBlobMetadata(blobMetadata))

	blobMetadata, err = memoryStore.AddDeploymentBlob(blobMetadata, b)
	require.Nil(t, err)
	require.NotNil(t, blobMetadata)

	runtime := actrs.Runtime{
		Store:      memoryStore,
		LogStore:   memoryStore,
		BlobStore:  memoryStore,
		Cache:      cacheStore,
		Deployment: blobMetadata.DeploymentID,
		StdOut:     new(bytes.Buffer),
	}
	require.Nil(t, runtime.Initialize(&pb.HTTPRequest{
		DeploymentId: deployment.ID.String(),
		Runtime:      "python",
	}))

	req := pb.HTTPRequest{
		Method:       "GET",
		Url:          "/url",
		Runtime:      "python",
		DeploymentId: deployment.ID.String(),
	}

	runtime.Handle(nil, &req)
}
