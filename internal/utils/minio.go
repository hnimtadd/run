package utils

import (
	"fmt"
	"strings"

	"github.com/hnimtadd/run/internal/types"
)

func CreateBlobObjectName(blob *types.BlobMetadata) string {
	return fmt.Sprintf("%v/%v", blob.EndpointID.String(), blob.DeploymentID.String())
}

// GetObjectNameFromLocation returns specific object path after remove location
func GetObjectNameFromLocation(location string) (string, error) {
	// http://localhost:9000/bucket/path
	path := strings.Split(location, "/")
	if len(path) < 5 {
		return "", fmt.Errorf("invalid location")
	}

	return strings.Join(path[4:], "/"), nil
}
