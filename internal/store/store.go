package store

import (
	"github.com/hnimtadd/run/internal/types"
)

type (
	Store interface {
		CreateEndpoint(endpoint *types.Endpoint) error
		UpdateEndpoint(endpointID string, params UpdateEndpointParams) error
		GetEndpointByID(endpointID string) (*types.Endpoint, error)
		GetEndpoints() ([]*types.Endpoint, error)
		UpdateActiveDeploymentOfEndpoint(endpointID string, deploymentID string) error

		CreateDeployment(deploy *types.Deployment) error
		GetDeploymentByID(deploymentID string) (*types.Deployment, error)
		GetDeployments() ([]*types.Deployment, error)
		DeleteDeployment(deploymentID string) error

		GetDeploymentsByEndpointID(endpointID string) ([]*types.Deployment, error)
	}
	UpdateEndpointParams struct {
		Environment map[string]string
	}

	LogStore interface {
		AppendLog(log *types.RequestLog) error
		GetLogByRequestID(requestID string) (*types.RequestLog, error)
		GetLogsOfRequest(deploymentID string, requestID string) (*types.RequestLog, error)
		GetLogOfDeployment(deploymentID string) ([]*types.RequestLog, error)
	}
	MetricStore interface {
		AddEndpointMetric(endpointID string, metrics types.RequestMetric) error
		GetMetricByEndpointID(endpointID string) (types.RuntimeMetric, error)
	}

	BlobStore interface {
		AddDeploymentBlob(*types.Blob) (*types.Blob, error)
		GetDeploymentBlobByDeploymentID(deploymentID string) (*types.Blob, error)
		GetBDeploymentBlobByBlobID(blobID string) (*types.Blob, error)
	}
)
