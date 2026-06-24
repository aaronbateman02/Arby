package discovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/aaronbateman02/Arby/internal/bus"
	"github.com/aaronbateman02/Arby/pkg/ingestion/discovery"
)

func TestScanner_RunAndStop(t *testing.T) {
	b := bus.New(10)
	s := discovery.NewScanner(b, time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.Run(ctx)
}
