package matching

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

type matchPairResponse struct {
	ID           string  `json:"id"`
	CandidateID  string  `json:"candidate_id"`
	MarketATitle string  `json:"market_a_title"`
	MarketBTitle string  `json:"market_b_title"`
	VenueA       string  `json:"venue_a"`
	VenueB       string  `json:"venue_b"`
	Category     string  `json:"category"`
	IsSameEvent  bool    `json:"is_same_event"`
	Relationship string  `json:"relationship"`
	Confidence   float64 `json:"confidence"`
	Reasoning    string  `json:"reasoning"`
	LegAModel    string  `json:"leg_a_model"`
	LegBModel    string  `json:"leg_b_model"`
	Status       string  `json:"status"`
}

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

type unembeddedResponse struct {
	Markets []MarketResponse `json:"markets"`
}

type MarketResponse struct {
	ID            string `json:"id"`
	Venue         string `json:"venue"`
	VenueMarketID string `json:"venue_market_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Category      string `json:"category"`
}

type embeddingsRequest struct {
	Embeddings []embeddingItem `json:"embeddings"`
}

type embeddingItem struct {
	ID     string    `json:"id"`
	Vector []float64 `json:"vector"`
}

type embeddingsResponse struct {
	Updated int `json:"updated"`
}

func (h *Handler) GetUnembedded(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 64
	if limitStr != "" {
		v, err := strconv.Atoi(limitStr)
		if err == nil && v > 0 {
			if v > 256 {
				limit = 256
			} else {
				limit = v
			}
		}
	}

	markets, err := h.store.GetUnembeddedMarkets(r.Context(), limit)
	if err != nil {
		slog.Error("get unembedded markets", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := unembeddedResponse{Markets: make([]MarketResponse, 0, len(markets))}
	for _, m := range markets {
		resp.Markets = append(resp.Markets, MarketResponse{
			ID:            m.ID,
			Venue:         m.Venue,
			VenueMarketID: m.VenueMarketID,
			Title:         m.Title,
			Description:   m.Description,
			Category:      m.Category,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("encode unembedded response", "error", err)
	}
}

func (h *Handler) PostEmbeddings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("read embeddings body", "error", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req embeddingsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		slog.Error("decode embeddings body", "error", err)
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	updated := 0
	for _, e := range req.Embeddings {
		if err := h.store.UpsertEmbedding(ctx, e.ID, e.Vector); err != nil {
			slog.Error("upsert embedding", "error", err, "id", e.ID)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		updated++
	}

	slog.InfoContext(ctx, "embeddings upserted", "count", updated)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(embeddingsResponse{Updated: updated}); err != nil {
		slog.Error("encode embeddings response", "error", err)
	}
}

func (h *Handler) GetMatchPairs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	pairs, err := h.store.GetMatchPairsWithDetails(r.Context(), status)
	if err != nil {
		slog.Error("get match pairs with details", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if pairs == nil {
		pairs = []matchPairResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"pairs": pairs}); err != nil {
		slog.Error("encode match pairs response", "error", err)
	}
}

func (h *Handler) ApproveMatchPair(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateMatchPairStatus(r.Context(), id, "APPROVED"); err != nil {
		slog.Error("approve match pair", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (h *Handler) RejectMatchPair(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateMatchPairStatus(r.Context(), id, "REJECTED"); err != nil {
		slog.Error("reject match pair", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (h *Handler) WireRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/markets/unembedded", h.GetUnembedded)
	mux.HandleFunc("POST /api/v1/markets/embeddings", h.PostEmbeddings)
	mux.HandleFunc("GET /api/v1/matching/pairs", h.GetMatchPairs)
	mux.HandleFunc("POST /api/v1/matching/pairs/{id}/approve", h.ApproveMatchPair)
	mux.HandleFunc("POST /api/v1/matching/pairs/{id}/reject", h.RejectMatchPair)
}
