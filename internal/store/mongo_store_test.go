package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"
	"github.com/hnimtadd/run/internal/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	testColEndpoint   = "endpoints"
	testColBlob       = "blobs"
	testColDeployment = "deployments"
	testDatabase      = "test-raptor"
	mongoClient       *mongo.Client
	mongoURL          = "mongodb://localhost:27018/testRaptor?authSource=admin"
)

func TestMongoStore_CreateDeployment(t *testing.T) {
	utils.SkipCI(t)
	db := getMongoDatabase(t)
	endpointCol := db.Collection(testColEndpoint)
	deploymentCol := db.Collection(testColDeployment)
	defer cleanCollection(t, endpointCol)
	defer cleanCollection(t, deploymentCol)
	endpoint, err := types.NewEndpoint("endpoint1", "go", make(map[string]string))
	require.Nil(t, err)

	deployment, err := types.NewDeployment(endpoint, nil, make(map[string]string))
	require.Nil(t, err)
	mongoStore := store.MongoStore{
		EndpointCol:   endpointCol,
		DeploymentCol: deploymentCol,
	}
	require.Nil(t, mongoStore.CreateEndpoint(endpoint))
	require.Nil(t, mongoStore.CreateDeployment(deployment))
	require.NotNil(t, mongoStore.CreateDeployment(deployment))
}

func TestMongoStore_CreateEndpoint(t *testing.T) {
	utils.SkipCI(t)
	db := getMongoDatabase(t)
	endpointCol := db.Collection(testColEndpoint)
	deploymentCol := db.Collection(testColDeployment)
	defer cleanCollection(t, endpointCol)
	defer cleanCollection(t, deploymentCol)
	endpoint, err := types.NewEndpoint("endpoint2", "go", make(map[string]string))
	require.Nil(t, err)

	mongoStore := store.MongoStore{
		EndpointCol:   endpointCol,
		DeploymentCol: deploymentCol,
	}

	require.Nil(t, mongoStore.CreateEndpoint(endpoint))
	require.NotNil(t, mongoStore.CreateEndpoint(endpoint))
}

func TestMongoStore_GetDeploymentByEndpointID(t *testing.T) {
	utils.SkipCI(t)
	db := getMongoDatabase(t)
	endpointCol := db.Collection(testColEndpoint)
	deploymentCol := db.Collection(testColDeployment)
	defer cleanCollection(t, endpointCol)
	defer cleanCollection(t, deploymentCol)
	endpoint, err := types.NewEndpoint("endpoint3", "go", make(map[string]string))
	require.Nil(t, err)
	deployment, err := types.NewDeployment(endpoint, nil, make(map[string]string))
	require.Nil(t, err)

	mongoStore := store.MongoStore{
		EndpointCol:   endpointCol,
		DeploymentCol: deploymentCol,
	}

	require.Nil(t, mongoStore.CreateEndpoint(endpoint))
	require.Nil(t, mongoStore.CreateDeployment(deployment))
	deployments, err := mongoStore.GetDeploymentsByEndpointID(endpoint.ID.String())
	require.Nil(t, err)
	require.NotNil(t, deployments)
	require.Equal(t, 1, len(deployments))
	require.Equal(t, *deployment, *deployments[0])

	nilDeployments, err := mongoStore.GetDeploymentsByEndpointID("ahlsfj")
	require.NotNil(t, err)
	require.Nil(t, nilDeployments)
}

func TestMongoStore_GetDeploymentByID(t *testing.T) {
	utils.SkipCI(t)
	db := getMongoDatabase(t)
	endpointCol := db.Collection(testColEndpoint)
	deploymentCol := db.Collection(testColDeployment)
	defer cleanCollection(t, endpointCol)
	defer cleanCollection(t, deploymentCol)
	endpoint, err := types.NewEndpoint("endpoint4", "go", make(map[string]string))
	require.Nil(t, err)
	deployment, err := types.NewDeployment(endpoint, nil, make(map[string]string))
	require.Nil(t, err)

	mongoStore := store.MongoStore{
		EndpointCol:   endpointCol,
		DeploymentCol: deploymentCol,
	}

	require.Nil(t, mongoStore.CreateEndpoint(endpoint))
	require.Nil(t, mongoStore.CreateDeployment(deployment))
	validDeployment, err := mongoStore.GetDeploymentByID(deployment.ID.String())
	require.Nil(t, err)
	require.NotNil(t, validDeployment)
	require.Equal(t, *deployment, *validDeployment)

	invalidDeployment, err := mongoStore.GetEndpointByID(uuid.Nil.String())
	require.NotNil(t, err)
	require.Nil(t, invalidDeployment)
}

//func TestMongoStore_GetDeployments(t *testing.T) {
//}

func TestMongoStore_GetEndpointByID(t *testing.T) {
	utils.SkipCI(t)
	db := getMongoDatabase(t)
	endpointCol := db.Collection(testColEndpoint)
	deploymentCol := db.Collection(testColDeployment)
	defer cleanCollection(t, endpointCol)
	defer cleanCollection(t, deploymentCol)
	endpoint, err := types.NewEndpoint("endpoint2", "go", make(map[string]string))
	require.Nil(t, err)

	mongoStore := store.MongoStore{
		EndpointCol:   endpointCol,
		DeploymentCol: deploymentCol,
	}

	require.Nil(t, mongoStore.CreateEndpoint(endpoint))
	require.NotNil(t, mongoStore.CreateEndpoint(endpoint))

	validEndpoint, err := mongoStore.GetEndpointByID(endpoint.ID.String())
	require.Nil(t, err)
	require.NotNil(t, validEndpoint)

	invalidEndpoint, err := mongoStore.GetEndpointByID("invalid id")
	require.NotNil(t, err)
	require.Nil(t, invalidEndpoint)

	invalidUUID := uuid.NewString()
	for {
		if invalidUUID != endpoint.ID.String() {
			break
		}
		invalidUUID = uuid.NewString()
	}
	invalidEndpoint, err = mongoStore.GetEndpointByID(invalidUUID)
	require.NotNil(t, err)
	require.Nil(t, invalidEndpoint)
}

func TestMongoStore_GetEndpoints(t *testing.T) {
	utils.SkipCI(t)
	db := getMongoDatabase(t)
	endpointCol := db.Collection(testColEndpoint)
	deploymentCol := db.Collection(testColDeployment)
	defer cleanCollection(t, endpointCol)
	defer cleanCollection(t, deploymentCol)
	endpoint, err := types.NewEndpoint("endpoint", "go", make(map[string]string))
	require.Nil(t, err)

	mongoStore := store.MongoStore{
		EndpointCol:   endpointCol,
		DeploymentCol: deploymentCol,
	}

	require.Nil(t, mongoStore.CreateEndpoint(endpoint))
	anotherEndpoint, err := types.NewEndpoint("endpoint another", "go", make(map[string]string))
	require.Nil(t, err)
	require.NotNil(t, anotherEndpoint)
	require.Nil(t, mongoStore.CreateEndpoint(anotherEndpoint))

	endpoints, err := mongoStore.GetEndpoints()
	require.Nil(t, err)
	require.Equal(t, 2, len(endpoints))
}

func TestMongoStore_UpdateActiveDeploymentOfEndpoint(t *testing.T) {
	utils.SkipCI(t)
	db := getMongoDatabase(t)
	endpointCol := db.Collection(testColEndpoint)
	deploymentCol := db.Collection(testColDeployment)
	defer cleanCollection(t, endpointCol)
	defer cleanCollection(t, deploymentCol)
	endpoint, err := types.NewEndpoint("endpoint2", "go", make(map[string]string))
	require.Nil(t, err)

	mongoStore := store.MongoStore{
		EndpointCol:   endpointCol,
		DeploymentCol: deploymentCol,
	}

	require.Nil(t, mongoStore.CreateEndpoint(endpoint))
	require.NotNil(t, mongoStore.CreateEndpoint(endpoint))

	deployment, err := types.NewDeployment(endpoint, nil, make(map[string]string))
	require.Nil(t, err)
	require.NotNil(t, deployment)

	require.Nil(t, mongoStore.CreateDeployment(deployment))
	require.Nil(t, mongoStore.UpdateActiveDeploymentOfEndpoint(endpoint.ID.String(), deployment.ID.String()))

	newEndpoint, err := mongoStore.GetEndpointByID(endpoint.ID.String())
	require.Nil(t, err)
	require.NotNil(t, newEndpoint)

	require.Equal(t, deployment.ID.String(), newEndpoint.ActiveDeploymentID.String())
}

func TestMongoStore_CreateBlobMetadata(t *testing.T) {
	utils.SkipCI(t)
	db := getMongoDatabase(t)
	blobCol := db.Collection(testColBlob)
	defer cleanCollection(t, blobCol)

	mongoStore := store.MongoStore{
		BlobCol: blobCol,
	}
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

	err = mongoStore.CreateBlobMetadata(blobMetadata)
	require.Nil(t, err)

	// Duplicated
	err = mongoStore.CreateBlobMetadata(blobMetadata)
	require.NotNil(t, err)
}

func TestMongoStore_GetBlobMetadataByDeploymentID(t *testing.T) {
	utils.SkipCI(t)
	db := getMongoDatabase(t)
	blobCol := db.Collection(testColBlob)
	defer cleanCollection(t, blobCol)

	mongoStore := store.MongoStore{
		BlobCol: blobCol,
	}
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

	failedBlobMetadata, err := mongoStore.GetBlobMetadataByDeploymentID(deployment.ID.String())
	require.NotNil(t, err)
	require.Nil(t, failedBlobMetadata)

	err = mongoStore.CreateBlobMetadata(blobMetadata)
	require.Nil(t, err)

	successBlobMetadata, err := mongoStore.GetBlobMetadataByDeploymentID(deployment.ID.String())
	require.Nil(t, err)
	require.NotNil(t, successBlobMetadata)

	require.Equal(t, *blobMetadata, *successBlobMetadata)
}

func getMongoDatabase(t *testing.T) *mongo.Database {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if mongoClient == nil {
		fmt.Println(mongoURL)
		opt := options.Client().ApplyURI(mongoURL)
		mongoClient, err = mongo.Connect(ctx, opt)
		require.Nil(t, err)
	}
	require.Nil(t, mongoClient.Ping(context.Background(), nil))
	return mongoClient.Database(testDatabase)
}

func cleanCollection(t *testing.T, col *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	require.Nil(t, col.Drop(ctx))
}
