package store

import (
	"context"
	"time"

	"github.com/hnimtadd/run/internal/errors"
	"github.com/hnimtadd/run/internal/types"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	EndpointColName   = "endpoints"
	DeploymentColName = "deployments"
)

type MongoStore struct {
	EndpointCol   *mongo.Collection
	DeploymentCol *mongo.Collection
}

func (m MongoStore) CreateEndpoint(endpoint *types.Endpoint) error {
	_, err := m.EndpointCol.InsertOne(context.Background(), endpoint)
	return err
}

func (m MongoStore) UpdateEndpoint(endpointID string, params UpdateEndpointParams) error {
	endpoint, err := m.GetEndpointByID(endpointID)
	if err != nil {
		return err
	}
	if endpoint.ActiveDeploymentID.String() != params.ActiveDeployID.String() {
		return errors.ErrRequestInvalidDeploymentID
	}

	filter := bson.M{"_id": endpoint.ID}
	update := bson.M{"$set": bson.M{"environment": params.Environment}}
	return m.EndpointCol.FindOneAndUpdate(context.Background(), filter, update).Err()
}

func (m MongoStore) GetEndpointByID(endpointID string) (*types.Endpoint, error) {
	endpointUID, err := uuid.Parse(endpointID)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": endpointUID}
	endpoint := new(types.Endpoint)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()
	if err := m.EndpointCol.FindOne(ctx, filter).Decode(endpoint); err != nil {
		return nil, err
	}
	return endpoint, nil
}

func (m MongoStore) GetEndpoints() ([]*types.Endpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cur, err := m.EndpointCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var endpoints []*types.Endpoint
	if err := cur.All(ctx, &endpoints); err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (m MongoStore) CreateDeployment(deploy *types.Deployment) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := m.DeploymentCol.InsertOne(ctx, deploy)
	return err
}

func (m MongoStore) GetDeploymentByID(deploymentID string) (*types.Deployment, error) {
	deploymentUID, err := uuid.Parse(deploymentID)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": deploymentUID}
	deployment := new(types.Deployment)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()
	if err := m.DeploymentCol.FindOne(ctx, filter).Decode(deployment); err != nil {
		return nil, err
	}
	return deployment, nil
}

func (m MongoStore) GetDeployments() ([]*types.Deployment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cur, err := m.DeploymentCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var deployments []*types.Deployment
	if err := cur.All(ctx, &deployments); err != nil {
		return nil, err
	}
	return deployments, nil
}

func (m MongoStore) GetDeploymentByEndpointID(endpointID string) ([]*types.Deployment, error) {
	endpointUID, err := uuid.Parse(endpointID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	filter := bson.M{"endpointID": endpointUID}
	cur, err := m.DeploymentCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var deployments []*types.Deployment
	if err := cur.All(ctx, &deployments); err != nil {
		return nil, err
	}
	return deployments, nil
}

func NewMongoStore(db *mongo.Database) (Store, error) {
	return &MongoStore{
		DeploymentCol: db.Collection(DeploymentColName),
		EndpointCol:   db.Collection(EndpointColName),
	}, nil
}
