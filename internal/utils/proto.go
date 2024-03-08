package utils

import "github.com/hnimtadd/run/pb/v1"

func MakeProtoRequest(requestID string) *pb.HTTPRequest {
	return &pb.HTTPRequest{
		Id: requestID,
	}
}
