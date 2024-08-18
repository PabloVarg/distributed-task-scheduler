package scheduler

import (
	"io"
	"log/slog"
	"testing"
)

func TestLoadBalanceBehaviour(t *testing.T) {
	pool := WorkerPool{
		ids:    make([]string, 0, 2),
		logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	pool.ids = append(pool.ids, "1")
	pool.ids = append(pool.ids, "2")

	selectedCounts := make(map[string]int)
	for range 4 {
		selectedId, err := pool.nextWorker()
		if err != nil {
			t.Errorf("expected no errors")
		}

		selectedCounts[selectedId] += 1
	}

	for _, value := range selectedCounts {
		if value != 2 {
			t.Errorf("task not evenly distributed")
		}
	}
}
