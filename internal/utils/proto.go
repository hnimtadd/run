package utils

import pb "github.com/hnimtadd/run/pbs/gopb/v1"

func MakeProtoRequest(requestID string) *pb.HTTPRequest {
	return &pb.HTTPRequest{
		Id: requestID,
	}
}
