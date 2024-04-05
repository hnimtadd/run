package types

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type Deployment struct {
	Hash string `json:"hash" bson:"hash"`
	// Deprecated, will move to use types.Blob instead, since we will save it to blob db,
	// Could use BlobID instead and get blob from blobstored
	Blob        []byte            `json:"blob" bson:"blob"`
	BlobID      uuid.UUID         `json:"blobID" bson:"blobID"`
	CreatedAt   int64             `json:"createdAt" bson:"createdAt"` // unix timestamp
	ID          uuid.UUID         `json:"id" bson:"_id"`
	EndpointID  uuid.UUID         `json:"endpointID" bson:"endpointID"`
	Environment map[string]string `json:"environment" bson:"environment"`
	Format      LogFormat         `json:"logFormat" bson:"format"`
}

func NewDeployment(endpoint *Endpoint, blob []byte, environment ...map[string]string) (*Deployment, error) {
	md5hash := md5.Sum(blob)
	deploymentHash := hex.EncodeToString(md5hash[:])
	deploymentID := uuid.New()

	var env map[string]string
	if len(environment) == 1 {
		env = environment[0]
	}

	deployment := &Deployment{
		ID:          deploymentID,
		Blob:        blob,
		Hash:        deploymentHash,
		EndpointID:  endpoint.ID,
		Environment: env,
		CreatedAt:   time.Now().Unix(),
	}
	return deployment, nil
}
