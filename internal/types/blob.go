package types

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"time"

	"github.com/google/uuid"
)

// BlobMetadata belongs to exact 1 deployment, so BlobMetata.ID makes no sense here.
type BlobMetadata struct {
	DeploymentID uuid.UUID `json:"deploymentID" bson:"_id"`          // deploymentID is unique so blobMetadata with deploymentID is _id should be unique too
	Hash         string    `json:"hash" bson:"hash"`                 // md5 hash
	VersionID    string    `json:"version_id" bson:"version_id"`     // this field will be setted after blob putted to object storage
	Location     string    `json:"storage_location" bson:"location"` // this field will be setted after blob putted to object storage
	CreatedAt    int64     `json:"createdAt" bson:"createdAt"`       // Unix timestamp
}

type BlobObject struct {
	Data         io.Reader
	UserMetadata map[string]string
	Etag         string
}

func NewRawBlobMetadata(deployment *Deployment, blob []byte) (*BlobMetadata, error) {
	md5Hash := md5.Sum(blob)
	blobHash := hex.EncodeToString(md5Hash[:])

	return &BlobMetadata{
		Hash:         blobHash,
		DeploymentID: deployment.ID,
		CreatedAt:    time.Now().Unix(),
	}, nil
}
