package utils_test

import (
	"testing"

	"github.com/hnimtadd/run/internal/utils"

	"github.com/stretchr/testify/require"
)

func Test_GetObjectNameFromLocation(t *testing.T) {
	validLocation := "http://localhost:9000/bucket/path"
	expectedObjectName := "path"
	currentObjectName, err := utils.GetObjectNameFromLocation(validLocation)
	require.Nil(t, err)
	require.Equal(t, expectedObjectName, currentObjectName)

	inValidLocation := "localhost:9000/bucket/path"
	expectedObjectName = ""
	currentObjectName, err = utils.GetObjectNameFromLocation(inValidLocation)
	require.NotNil(t, err)
	require.Equal(t, expectedObjectName, currentObjectName)
}
