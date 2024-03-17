package types

import "github.com/google/uuid"

//const (
//	LogFormatString LogFormat = iota
//	LogFormatJSON
//)

type LogFormat byte

type RequestLog struct {
	RequestID    uuid.UUID
	DeploymentID uuid.UUID
	Contents     []string
	CreatedAt    int64
}
