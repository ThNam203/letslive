package services

import (
	"os"
	"testing"

	"sen1or/letslive/shared/pkg/logger"
)

func TestMain(m *testing.M) {
	logger.Init(logger.Error)
	os.Exit(m.Run())
}
