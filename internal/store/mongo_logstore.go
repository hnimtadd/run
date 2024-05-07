package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hnimtadd/run/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var LogColName = "logs"

type MongoLogStore struct {
	LogCol *mongo.Collection
}

func NewMongoLogStore(db *mongo.Database) (LogStore, error) {
	return &MongoLogStore{
		LogCol: db.Collection(LogColName),
	}, nil
}

// AppendLog implements LogStore.
func (m *MongoLogStore) AppendLog(log *types.RequestLog) error {
	log.CreatedAt = time.Now().Unix()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := m.LogCol.InsertOne(ctx, log)
	return err
}

// GetLogByRequestID implements LogStore.
func (m *MongoLogStore) GetLogByRequestID(requestID string) (*types.RequestLog, error) {
	requestUID, err := uuid.Parse(requestID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	filter := bson.M{"_id": requestUID}
	var res types.RequestLog
	if err := m.LogCol.FindOne(ctx, filter).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetLogOfDeployment implements LogStore.
func (m *MongoLogStore) GetLogOfDeployment(deploymentID string) ([]*types.RequestLog, error) {
	deploymentUID, err := uuid.Parse(deploymentID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	filter := bson.M{"deployment_id": deploymentUID}
	cur, err := m.LogCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	res := make([]*types.RequestLog, 0)
	err = cur.All(ctx, &res)
	return res, err
}

// GetLogsOfRequest implements LogStore.
func (m *MongoLogStore) GetLogsOfRequest(deploymentID string, requestID string) (*types.RequestLog, error) {
	requestUID, err := uuid.Parse(requestID)
	if err != nil {
		return nil, err
	}
	deploymentUID, err := uuid.Parse(deploymentID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	filter := bson.M{"_id": requestUID, "deployment_id": deploymentUID}
	res := new(types.RequestLog)
	err = m.LogCol.FindOne(ctx, filter).Decode(res)
	return res, err
}
