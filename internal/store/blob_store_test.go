package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"
	"github.com/hnimtadd/run/internal/utils"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/require"
)

var (
	testBucket    = "test-raptor"
	minioClient   *minio.Client
	minioUserName = "development"
	minioPassword = "development"
	minioURL      = "localhost:9000"
)

func TestBlobStore_AddDeploymentBlob(t *testing.T) {
	utils.SkipCI(t)
	minioClient := GetMinioClient(t)
	blobStore, err := store.NewMinioBlobStore(minioClient, testBucket)
	require.Nil(t, err)

	endpoint, err := types.NewEndpoint("endpoint-2", "go", nil)
	require.Nil(t, err)
	require.NotNil(t, endpoint)
	blob := []byte("hello world")

	deployment, err := types.NewDeployment(endpoint, nil)
	require.Nil(t, err)
	require.NotNil(t, deployment)

	blobMetadata, err := types.NewRawBlobMetadata(deployment, blob)
	require.Nil(t, err)
	require.NotNil(t, blobMetadata)

	newBlobMetadata, err := blobStore.AddDeploymentBlob(blobMetadata, blob)
	require.Nil(t, err)
	require.NotNil(t, newBlobMetadata)
	require.Equal(t, blobMetadata.Hash, newBlobMetadata.Hash)
	require.Equal(t, blobMetadata.DeploymentID, newBlobMetadata.DeploymentID)
	require.Equal(t, blobMetadata.CreatedAt, newBlobMetadata.CreatedAt)

	require.NotEmpty(t, blobMetadata.Location)
	CleanBucket(t)
}

func TestBlobStore_GetDeploymentBlobByURI(t *testing.T) {
	utils.SkipCI(t)
	minioClient := GetMinioClient(t)

	blobStore, err := store.NewMinioBlobStore(minioClient, testBucket)
	require.Nil(t, err)

	endpoint, err := types.NewEndpoint("endpoint-2", "go", nil)
	require.Nil(t, err)
	require.NotNil(t, endpoint)
	blob := []byte("hello world")

	deployment, err := types.NewDeployment(endpoint, nil)
	require.Nil(t, err)
	require.NotNil(t, deployment)

	blobMetadata, err := types.NewRawBlobMetadata(deployment, blob)
	require.Nil(t, err)
	require.NotNil(t, blobMetadata)

	newBlobMetadata, err := blobStore.AddDeploymentBlob(blobMetadata, blob)
	require.Nil(t, err)
	require.NotNil(t, newBlobMetadata)

	getBlobMetadata, err := blobStore.GetDeploymentBlobByURI(newBlobMetadata.Location)
	require.Nil(t, err)
	require.NotNil(t, getBlobMetadata)
	require.Equal(t, blob, getBlobMetadata.Data)

	// version 2
	anotherBlob := []byte("new blob")
	anotherblobMetadata, err := types.NewRawBlobMetadata(deployment, anotherBlob)
	require.Nil(t, err)
	require.NotNil(t, anotherBlob)

	newAnotherBlobMetadata, err := blobStore.AddDeploymentBlob(anotherblobMetadata, anotherBlob)
	require.Nil(t, err)
	require.NotNil(t, newAnotherBlobMetadata)

	getAnotherBlobMetadata, err := blobStore.GetDeploymentBlobByURI(newAnotherBlobMetadata.Location)
	require.Nil(t, err)
	require.NotNil(t, getAnotherBlobMetadata)
	require.Equal(t, anotherBlob, getAnotherBlobMetadata.Data)

	CleanBucket(t)
}

func TestBlobStore_DeleteDeploymentBlob(t *testing.T) {
	utils.SkipCI(t)
	minioClient := GetMinioClient(t)
	blobStore, err := store.NewMinioBlobStore(minioClient, testBucket)
	require.Nil(t, err)

	endpoint, err := types.NewEndpoint("endpoint-2", "go", nil)
	require.Nil(t, err)
	require.NotNil(t, endpoint)
	blob := []byte("hello world")

	deployment, err := types.NewDeployment(endpoint, nil)
	require.Nil(t, err)
	require.NotNil(t, deployment)

	blobMetadata, err := types.NewRawBlobMetadata(deployment, blob)
	require.Nil(t, err)
	require.NotNil(t, blobMetadata)

	newBlobMetadata, err := blobStore.AddDeploymentBlob(blobMetadata, blob)
	require.Nil(t, err)
	require.NotNil(t, newBlobMetadata)

	getBlobMetadata, err := blobStore.GetDeploymentBlobByURI(newBlobMetadata.Location)
	require.Nil(t, err)
	require.NotNil(t, getBlobMetadata)

	found, err := blobStore.DeleteDeploymentBlob(newBlobMetadata.Location)
	require.Nil(t, err)
	require.True(t, found)

	failedBlobMetadata, err := blobStore.GetDeploymentBlobByURI(newBlobMetadata.Location)
	require.NotNil(t, err)
	require.Nil(t, failedBlobMetadata)
	CleanBucket(t)
}

func GetMinioClient(t *testing.T) *minio.Client {
	if minioClient != nil {
		return minioClient
	}

	creds := credentials.NewStaticV4(minioUserName, minioPassword, "")
	var err error
	minioClient, err = minio.New(minioURL, &minio.Options{
		Creds:  creds,
		Secure: false,
	})
	require.Nil(t, err)
	return minioClient
}

func DeleteMinioClient() {
	minioClient = nil
}

func CleanBucket(t *testing.T) {
	client := GetMinioClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	objectCh := client.ListObjects(context.Background(), testBucket, minio.ListObjectsOptions{
		WithVersions: true,
		Recursive:    true,
	})

	for err := range client.RemoveObjects(
		context.Background(),
		testBucket,
		objectCh,
		minio.RemoveObjectsOptions{}) {
		require.Nil(t, err)
	}
	err := client.RemoveBucket(ctx, testBucket)
	require.Nil(t, err)
	DeleteMinioClient()
}
