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

	firstID := ""
	lastID := ""
	sampleIDs := make([]string, 0, 5)
	if len(req.Embeddings) > 0 {
		firstID = req.Embeddings[0].ID
		lastID = req.Embeddings[len(req.Embeddings)-1].ID
		for i := 0; i < len(req.Embeddings); i += 50 {
			sampleIDs = append(sampleIDs, req.Embeddings[i].ID[:8])
		}
	}
	slog.InfoContext(ctx, "embeddings upserted", "count", updated, "first", firstID[:8], "last", lastID[:8], "remote", r.RemoteAddr, "ua", r.UserAgent(), "samples", sampleIDs)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(embeddingsResponse{Updated: updated}); err != nil {
		slog.Error("encode embeddings response", "error", err)
	}
}

func (h *Handler) GetMatchPairs(w http.ResponseWriter, r *http.Request) {
	if marketID := r.URL.Query().Get("market_id"); marketID != "" {
		h.getMatchPairsByMarket(w, r, marketID)
		return
	}
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
	json.NewEncoder(w).Encode(map[string]any{"pairs": pairs})
}

func (h *Handler) SearchMarkets(w http.ResponseWriter, r *http.Request) {
	venue := r.URL.Query().Get("venue")
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")

	if venue == "" || query == "" {
		http.Error(w, "venue and q params required", http.StatusBadRequest)
		return
	}

	limit := 20
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}

	ctx := r.Context()
	markets, err := h.store.SearchMarkets(ctx, venue, query, limit)
	if err != nil {
		slog.Error("search markets", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]MarketResponse, 0, len(markets))
	for _, m := range markets {
		resp = append(resp, MarketResponse{
			ID:            m.ID,
			Venue:         m.Venue,
			VenueMarketID: m.VenueMarketID,
			Title:         m.Title,
			Description:   m.Description,
			Category:      m.Category,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"markets": resp})
}

func (h *Handler) getMatchPairsByMarket(w http.ResponseWriter, r *http.Request, marketID string) {
	ctx := r.Context()
	pairs, err := h.store.GetMatchPairsByMarket(ctx, marketID)
	if err != nil {
		slog.Error("get pairs by market", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if pairs == nil {
		pairs = []matchPairResponse{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"pairs": pairs})
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

type settingsResponse struct {
	SimilarityThreshold float64 `json:"similarity_threshold"`
}

type settingsUpdateRequest struct {
	SimilarityThreshold float64 `json:"similarity_threshold"`
}

func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	threshold, err := h.store.GetSimilarityThreshold(ctx)
	if err != nil {
		slog.Error("get settings", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settingsResponse{SimilarityThreshold: threshold})
}

func (h *Handler) PostSettings(w http.ResponseWriter, r *http.Request) {
	var req settingsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.SimilarityThreshold < 0 || req.SimilarityThreshold > 1 {
		http.Error(w, "similarity_threshold must be between 0 and 1", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	if err := h.store.SetSimilarityThreshold(ctx, req.SimilarityThreshold); err != nil {
		slog.Error("set settings", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := h.store.GetStats(ctx)
	if err != nil {
		slog.Error("get stats", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) GetPipelineCounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	counts, err := h.store.GetPipelineCounts(ctx)
	if err != nil {
		slog.Error("get pipeline counts", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

func (h *Handler) GetTopSimilarities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pairs, err := h.store.GetTopSimilarities(ctx, 100)
	if err != nil {
		slog.Error("get top similarities", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pairs)
}

func (h *Handler) GetEmbedScript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/x-python")
	w.Header().Set("Content-Disposition", "attachment; filename=embed_worker.py")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(embedWorkerScript))
}

func (h *Handler) WireRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/markets/unembedded", h.GetUnembedded)
	mux.HandleFunc("POST /api/v1/markets/embeddings", h.PostEmbeddings)
	mux.HandleFunc("GET /api/v1/matching/markets/search", h.SearchMarkets)
	mux.HandleFunc("GET /api/v1/matching/pairs", h.GetMatchPairs)
	mux.HandleFunc("POST /api/v1/matching/pairs/{id}/approve", h.ApproveMatchPair)
	mux.HandleFunc("POST /api/v1/matching/pairs/{id}/reject", h.RejectMatchPair)
	mux.HandleFunc("GET /api/v1/matching/settings", h.GetSettings)
	mux.HandleFunc("POST /api/v1/matching/settings", h.PostSettings)
	mux.HandleFunc("GET /api/v1/matching/stats", h.GetStats)
	mux.HandleFunc("GET /api/v1/matching/pipeline-counts", h.GetPipelineCounts)
	mux.HandleFunc("GET /api/v1/matching/top-similarities", h.GetTopSimilarities)
	mux.HandleFunc("GET /api/v1/matching/embed-script", h.GetEmbedScript)
}
