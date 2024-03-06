package errors

import "errors"

var (
	ErrEndpointExisted      = errors.New("given endpoint is existed")
	ErrEndpointNotExisted   = errors.New("given endpoint is not existed")
	ErrDeploymentExisted    = errors.New("given deployment is existed")
	ErrDeploymentNotExisted = errors.New("given deployment is not existed")
)
