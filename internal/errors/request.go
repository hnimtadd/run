package errors

import "errors"

var (
	ErrInvalidRequest             = errors.New("given request is invalid")
	ErrRequestInvalidDeploymentID = errors.New("given request has invalid deploymentID")
)
