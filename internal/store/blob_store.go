package store

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hnimtadd/run/internal/types"
	"github.com/hnimtadd/run/internal/utils"

	"github.com/minio/minio-go/v7"
)

type MinioBlobStore struct {
	client     *minio.Client
	bucketName string
}

func NewMinioBlobStore(client *minio.Client, bucketName string) (BlobStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	found, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := client.EnableVersioning(ctx, bucketName); err != nil {
			slog.Error("cannot enable versioning for bucket", "err", err, "bucket", bucketName)
		}
	}()

	blobStore := &MinioBlobStore{
		client:     client,
		bucketName: bucketName,
	}
	if found {
		return blobStore, nil
	}
	if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
		return nil, err
	}
	return blobStore, nil
}

// AddDeploymentBlob implements BlobStore.
func (m *MinioBlobStore) AddDeploymentBlob(blob *types.BlobMetadata, data []byte) (*types.BlobMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	info, err := m.client.PutObject(ctx, m.bucketName, utils.CreateBlobObjectName(blob), bytes.NewReader(data), -1, minio.PutObjectOptions{
		UserMetadata: map[string]string{
			"deploymentID": blob.DeploymentID.String(),
			"hash":         blob.Hash,
		},
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return nil, err
	}
	// if info.ETag != blob.Hash {
	// 	// check if the uploaded blob was received successfully by serve
	// 	return nil, errors.Newf(
	// 		"checksum failed, server receive wrong file, given: %s, got: %s",
	// 		blob.Hash,
	// 		info.ETag,
	// 	)
	// }
	blob.Location = info.Location
	blob.VersionID = info.VersionID
	fmt.Println("info", info.Key)
	return blob, nil
}

// GetDeploymentBlobByURI implements BlobStore.
func (m *MinioBlobStore) GetDeploymentBlobByURI(location string, versionID ...string) (*types.BlobObject, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	opts := minio.GetObjectOptions{}
	if len(versionID) == 1 {
		opts.VersionID = versionID[0]
	}

	fmt.Println(location)
	objectName, err := utils.GetObjectNameFromLocation(location)
	if err != nil {
		return nil, err
	}
	object, err := m.client.GetObject(ctx, m.bucketName, objectName, opts)
	if err != nil {
		return nil, err
	}

	info, err := object.Stat()
	if err != nil {
		return nil, err
	}

	return &types.BlobObject{
		Data:         object,
		Etag:         info.ETag,
		UserMetadata: info.UserMetadata,
	}, nil
}

func (m *MinioBlobStore) DeleteDeploymentBlob(location string, versionID ...string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	opts := minio.StatObjectOptions{}
	if len(versionID) == 1 {
		opts.VersionID = versionID[0]
	}
	objectName, err := utils.GetObjectNameFromLocation(location)
	if err != nil {
		return false, err
	}
	_, err = m.client.StatObject(ctx, m.bucketName, objectName, opts)
	if err != nil {
		return false, err
	}
	removeOpts := minio.RemoveObjectOptions{}
	if len(versionID) == 1 {
		opts.VersionID = versionID[0]
	}

	err = m.client.RemoveObject(ctx, m.bucketName, objectName, removeOpts)
	return true, err
}
