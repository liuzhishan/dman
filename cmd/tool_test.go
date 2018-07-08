package cmd

import (
	"os"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	logInfo("%d", time.Now().Unix())
}
