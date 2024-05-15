package types

import (
	"time"

	"github.com/google/uuid"
)

type Deployment struct {
	ID          uuid.UUID         `json:"id" bson:"_id"`
	Hash        string            `json:"hash" bson:"hash"` /* Deprecated, this field will move to types.Blob*/
	BlobID      uuid.UUID         `json:"blobID" bson:"blobID"`
	CreatedAt   int64             `json:"createdAt" bson:"createdAt"` // unix timestamp
	EndpointID  uuid.UUID         `json:"endpointID" bson:"endpointID"`
	Environment map[string]string `json:"environment" bson:"environment"`
	Format      LogFormat         `json:"logFormat" bson:"format"`
}

func NewDeployment(endpoint *Endpoint, environment ...map[string]string) (*Deployment, error) {
	deploymentID := uuid.New()

	var env map[string]string
	if len(environment) == 1 {
		env = environment[0]
	}

	deployment := &Deployment{
		ID:          deploymentID,
		EndpointID:  endpoint.ID,
		Environment: env,
		CreatedAt:   time.Now().Unix(),
	}
	return deployment, nil
}
