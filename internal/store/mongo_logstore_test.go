package store_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"
	"github.com/hnimtadd/run/internal/utils"
	"github.com/stretchr/testify/require"
)

var testColLog = "logs"

func TestAppendLog(t *testing.T) {
	utils.SkipCI(t)
	db := getMongoDatabase(t)

	logCol := db.Collection(testColLog)
	defer cleanCollection(t, logCol)

	store := store.MongoLogStore{
		LogCol: logCol,
	}

	requestID := uuid.New()
	deploymentID := uuid.New()
	logEvent := types.RequestLog{
		RequestID:    requestID,
		DeploymentID: deploymentID,
		Contents:     []string{"this is first line", "this is second line"},
	}
	err := store.AppendLog(&logEvent)
	require.Nil(t, err)

	getEvent, err := store.GetLogByRequestID(requestID.String())
	require.Nil(t, err)
	require.Equal(t, logEvent.Contents, getEvent.Contents)
	require.Equal(t, logEvent.DeploymentID, getEvent.DeploymentID)
	require.Equal(t, logEvent.RequestID, getEvent.RequestID)

	events, err := store.GetLogOfDeployment(deploymentID.String())
	require.Nil(t, err)
	require.Equal(t, 1, len(events))
	event := events[0]
	require.Equal(t, logEvent.Contents, event.Contents)
	require.Equal(t, logEvent.DeploymentID, event.DeploymentID)
	require.Equal(t, logEvent.RequestID, getEvent.RequestID)

	failedGetEvent, err := store.GetLogByRequestID(uuid.Nil.String())
	require.NotNil(t, err)
	require.Nil(t, failedGetEvent)
}
