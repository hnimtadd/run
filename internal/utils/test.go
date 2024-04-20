package utils

import (
	"os"
	"testing"
)

func SkipCI(t *testing.T) {
	if os.Getenv("ENV") == "CI" {
		t.Skip("Skipping testing in CI environment")
	}
}
