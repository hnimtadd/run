package store

import (
	"context"
	"fmt"
	"time"

	"github.com/hnimtadd/run/internal/types"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	EndpointColName   = "endpoints"
	DeploymentColName = "deployments"
)

type MongoStore struct {
	EndpointCol   *mongo.Collection
	DeploymentCol *mongo.Collection
}

func (m MongoStore) UpdateActiveDeploymentOfEndpoint(endpointID string, deploymentID string) error {
	endpoint, err := m.GetEndpointByID(endpointID)
	if err != nil {
		return err
	}

	deploymentUID, err := uuid.Parse(deploymentID)
	if err != nil {
		return err
	}

	_, err = m.GetDeploymentByID(deploymentID)
	if err != nil {
		// here, we could meet the case where user want to update nil uuid
		if deploymentID != uuid.Nil.String() {
			return err
		}
	}

	filter := bson.M{"_id": endpoint.ID}
	update := bson.M{"$set": bson.M{"activeDeploymentID": deploymentUID}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err = m.EndpointCol.UpdateOne(ctx, filter, update)
	return err
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

	filter := bson.M{"_id": endpoint.ID}
	currEnv := endpoint.Environment

	for k, v := range params.Environment {
		currEnv[k] = v
	}

	update := bson.M{"$set": bson.M{"environment": currEnv}}
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

func (m MongoStore) GetDeploymentsByEndpointID(endpointID string) ([]*types.Deployment, error) {
	endpointUID, err := uuid.Parse(endpointID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	filter := bson.M{"endpointID": endpointUID}

	// find by ascending order by createdAt timestamp
	cur, err := m.DeploymentCol.Find(ctx, filter, options.Find().SetSort(bson.M{"createdAt": 1}))
	if err != nil {
		return nil, err
	}
	var deployments []*types.Deployment
	if err := cur.All(ctx, &deployments); err != nil {
		return nil, err
	}
	return deployments, nil
}

func (m MongoStore) DeleteDeployment(deploymentID string) error {
	deployment, err := m.GetDeploymentByID(deploymentID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": deployment.ID}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cur, err := m.DeploymentCol.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if cur.DeletedCount != int64(1) {
		return fmt.Errorf("store: unexpected error, expected delete 1 document, got %d", cur.DeletedCount)
	}
	return nil
}

func NewMongoStore(db *mongo.Database) (Store, error) {
	return &MongoStore{
		DeploymentCol: db.Collection(DeploymentColName),
		EndpointCol:   db.Collection(EndpointColName),
	}, nil
}
