package health

import (
	"context"
	"encoding/json"
	"net/http"
)

type Check func(ctx context.Context) error

type Health struct {
	dbCheck    Check
	redisCheck Check
}

func New(dbCheck, redisCheck Check) *Health {
	return &Health{
		dbCheck:    dbCheck,
		redisCheck: redisCheck,
	}
}

func (h *Health) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"alive"}`))
	}
}

func (h *Health) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		checks := map[string]string{}

		if h.dbCheck != nil {
			if err := h.dbCheck(r.Context()); err != nil {
				checks["database"] = err.Error()
				status = http.StatusServiceUnavailable
			} else {
				checks["database"] = "ok"
			}
		}

		if h.redisCheck != nil {
			if err := h.redisCheck(r.Context()); err != nil {
				checks["redis"] = err.Error()
				status = http.StatusServiceUnavailable
			} else {
				checks["redis"] = "ok"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": status,
			"checks": checks,
		})
	}
}
