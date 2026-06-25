# Matching Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the single Review page with a full Matching section (Pipeline / Review / Settings) with configurable similarity threshold, pipeline stats, market search, and downloadable embed worker.

**Architecture:** Five new Go backend endpoints serve stats, settings, market search, and embed script. The UI gets a collapsible "Matching" sidebar group with three sub-pages. A Python embed worker script is served by the backend for download.

**Tech Stack:** Go 1.22 (pgx, pgvector), Next.js 14 (App Router, Tailwind), Python 3.9+ (sentence-transformers, bge-large-en-v1.5)

**Embed script note:** The `GET /api/v1/matching/embed-script` endpoint serves the script from a Go string constant in `pkg/matching/embed_worker_data.go`. The `tools/embed_worker.py` file is a user-facing copy; keep both in sync.

## Global Constraints

- DB schema uses `VECTOR(1024)` for embeddings
- Embedding model is `BAAI/bge-large-en-v1.5` (outputs 1024-dim vectors)
- Sidebar uses click-to-expand collapsible groups (not hover dropdowns)
- All new API endpoints prefixed with `/api/v1/matching/`
- Auth: no auth for now (same as existing endpoints)
- YAGNI — don't add features beyond what's specified

---

### Task 1: matching_config table + settings endpoints

**Files:**
- Modify: `pkg/matching/store.go`
- Modify: `pkg/matching/handler.go`
- Test: `pkg/matching/store_test.go`

**Interfaces:**
- Produces: `Store.GetSimilarityThreshold(ctx) (float64, error)` — reads from `matching_config` table
- Produces: `Store.SetSimilarityThreshold(ctx, val float64) error` — upserts into `matching_config` table
- Produces: `GET /api/v1/matching/settings` → `{"similarity_threshold": 0.80}`
- Produces: `POST /api/v1/matching/settings` body `{"similarity_threshold": 0.85}` → `{"ok": true}`

- [ ] **Step 1: Add matching_config DDL to CreateTables**

In `pkg/matching/store.go`, add to the `CreateTables` SQL batch:

```go
CREATE TABLE IF NOT EXISTS matching_config (
    key   VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL
);
INSERT INTO matching_config (key, value) VALUES ('similarity_threshold', '0.80')
ON CONFLICT (key) DO NOTHING;
```

- [ ] **Step 2: Add store methods**

Add to `Store` in `pkg/matching/store.go`:

```go
func (s *Store) GetSimilarityThreshold(ctx context.Context) (float64, error) {
    var val string
    err := s.pg.P().QueryRow(ctx,
        `SELECT value FROM matching_config WHERE key = 'similarity_threshold'`,
    ).Scan(&val)
    if err != nil {
        return 0.80, fmt.Errorf("get similarity threshold: %w", err)
    }
    f, err := strconv.ParseFloat(val, 64)
    if err != nil {
        return 0.80, nil
    }
    return f, nil
}

func (s *Store) SetSimilarityThreshold(ctx context.Context, val float64) error {
    _, err := s.pg.P().Exec(ctx,
        `INSERT INTO matching_config (key, value) VALUES ('similarity_threshold', $1)
         ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
        fmt.Sprintf("%.4f", val),
    )
    if err != nil {
        return fmt.Errorf("set similarity threshold: %w", err)
    }
    return nil
}
```

Add import for `"strconv"`.

- [ ] **Step 3: Add settings handlers**

Add to `pkg/matching/handler.go`:

```go
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
```

- [ ] **Step 4: Wire routes**

Add to `WireRoutes`:

```go
mux.HandleFunc("GET /api/v1/matching/settings", h.GetSettings)
mux.HandleFunc("POST /api/v1/matching/settings", h.PostSettings)
```

- [ ] **Step 5: Run tests to verify**

Run: `cd Arby && go test ./pkg/matching/ -run TestJoinFloats -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/matching/store.go pkg/matching/handler.go
git commit -m "feat: add matching_config table and settings endpoints"
```

---

### Task 2: Stats endpoint

**Files:**
- Modify: `pkg/matching/store.go`
- Modify: `pkg/matching/handler.go`

**Interfaces:**
- Produces: `Store.GetStats(ctx) (*Stats, error)` — aggregates across tables
- Produces: `GET /api/v1/matching/stats` → `{"unembedded": N, "embedded": N, "pending_candidates": N, "reviewed_candidates": N, "pairs_pending_approval": N, "pairs_approved": N, "pairs_rejected": N}`

- [ ] **Step 1: Define stats type and store method**

Add to `pkg/matching/store.go`:

```go
type Stats struct {
    Unembedded          int `json:"unembedded"`
    Embedded            int `json:"embedded"`
    PendingCandidates   int `json:"pending_candidates"`
    ReviewedCandidates  int `json:"reviewed_candidates"`
    PairsPendingApproval int `json:"pairs_pending_approval"`
    PairsApproved       int `json:"pairs_approved"`
    PairsRejected       int `json:"pairs_rejected"`
}

func (s *Store) GetStats(ctx context.Context) (*Stats, error) {
    var st Stats
    err := s.pg.P().QueryRow(ctx, `
        SELECT
            (SELECT COUNT(*) FROM markets WHERE embedding IS NULL AND status = 'OPEN' AND COALESCE(title, '') <> '' AND description IS NOT NULL AND TRIM(description) <> ''),
            (SELECT COUNT(*) FROM markets WHERE embedding IS NOT NULL AND status = 'OPEN'),
            (SELECT COUNT(*) FROM match_candidates WHERE status = 'PENDING'),
            (SELECT COUNT(*) FROM match_candidates WHERE status = 'REVIEWED'),
            (SELECT COUNT(*) FROM match_pairs WHERE status = 'PENDING_APPROVAL'),
            (SELECT COUNT(*) FROM match_pairs WHERE status = 'APPROVED'),
            (SELECT COUNT(*) FROM match_pairs WHERE status = 'REJECTED')
    `).Scan(
        &st.Unembedded, &st.Embedded, &st.PendingCandidates, &st.ReviewedCandidates,
        &st.PairsPendingApproval, &st.PairsApproved, &st.PairsRejected,
    )
    if err != nil {
        return nil, fmt.Errorf("get stats: %w", err)
    }
    return &st, nil
}
```

- [ ] **Step 2: Add handler and wire route**

Add to `pkg/matching/handler.go`:

```go
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
```

Wire in `WireRoutes`:

```go
mux.HandleFunc("GET /api/v1/matching/stats", h.GetStats)
```

- [ ] **Step 3: Build check**

Run: `cd Arby && go build ./cmd/polybot`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add pkg/matching/store.go pkg/matching/handler.go
git commit -m "feat: add matching stats endpoint"
```

---

### Task 3: Market search + pairs-by-market filter

**Files:**
- Modify: `pkg/matching/store.go`
- Modify: `pkg/matching/handler.go`

**Interfaces:**
- Produces: `Store.SearchMarkets(ctx, venue, query string, limit int) ([]Market, error)` — text search using `ILIKE`
- Produces: `Store.GetMatchPairsByMarket(ctx, marketID string) ([]matchPairResponse, error)` — pairs involving a market
- Produces: `GET /api/v1/matching/markets/search?venue=KALSHI&q=fed&limit=10` → `{"markets": [...]}`
- Produces: `GET /api/v1/matching/pairs?market_id=<uuid>` → `{"pairs": [...]}`

- [ ] **Step 1: Add SearchMarkets store method**

Add to `pkg/matching/store.go`:

```go
func (s *Store) SearchMarkets(ctx context.Context, venue, query string, limit int) ([]Market, error) {
    if limit <= 0 || limit > 100 {
        limit = 20
    }
    rows, err := s.pg.P().Query(ctx, `
        SELECT id, venue, venue_market_id, COALESCE(title,''), COALESCE(description,''), COALESCE(category,'')
        FROM markets
        WHERE venue = $1
          AND (title ILIKE '%' || $2 || '%' OR venue_market_id ILIKE '%' || $2 || '%')
          AND status = 'OPEN'
        ORDER BY
            CASE WHEN title ILIKE $2 || '%' THEN 0 ELSE 1 END,
            last_updated_at DESC
        LIMIT $3
    `, venue, query, limit)
    if err != nil {
        return nil, fmt.Errorf("search markets: %w", err)
    }
    defer rows.Close()

    var result []Market
    for rows.Next() {
        var m Market
        if err := rows.Scan(&m.ID, &m.Venue, &m.VenueMarketID, &m.Title, &m.Description, &m.Category); err != nil {
            return nil, fmt.Errorf("search markets scan: %w", err)
        }
        result = append(result, m)
    }
    return result, rows.Err()
}
```

- [ ] **Step 2: Add GetMatchPairsByMarket store method**

Add to `pkg/matching/store.go`:

```go
func (s *Store) GetMatchPairsByMarket(ctx context.Context, marketID string) ([]matchPairResponse, error) {
    rows, err := s.pg.P().Query(ctx, `
        SELECT
            mp.id,
            mp.candidate_id,
            ma.title AS market_a_title,
            mb.title AS market_b_title,
            ma.venue AS venue_a,
            mb.venue AS venue_b,
            COALESCE(mc.category, '') AS category,
            COALESCE(mp.is_same_event, false) AS is_same_event,
            COALESCE(mp.relationship, '') AS relationship,
            COALESCE(mp.confidence, 0) AS confidence,
            COALESCE(mp.reasoning, '') AS reasoning,
            COALESCE(mp.leg_a_model, '') AS leg_a_model,
            COALESCE(mp.leg_b_model, '') AS leg_b_model,
            mp.status
        FROM match_pairs mp
        JOIN match_candidates mc ON mc.id = mp.candidate_id
        JOIN markets ma ON ma.id = mc.market_a_id
        JOIN markets mb ON mb.id = mc.market_b_id
        WHERE mc.market_a_id = $1 OR mc.market_b_id = $1
        ORDER BY mp.reviewed_at DESC, mc.created_at DESC
    `, marketID)
    if err != nil {
        return nil, fmt.Errorf("get pairs by market: %w", err)
    }
    defer rows.Close()

    var result []matchPairResponse
    for rows.Next() {
        var p matchPairResponse
        if err := rows.Scan(
            &p.ID, &p.CandidateID, &p.MarketATitle, &p.MarketBTitle,
            &p.VenueA, &p.VenueB, &p.Category, &p.IsSameEvent,
            &p.Relationship, &p.Confidence, &p.Reasoning,
            &p.LegAModel, &p.LegBModel, &p.Status,
        ); err != nil {
            return nil, fmt.Errorf("pairs by market scan: %w", err)
        }
        result = append(result, p)
    }
    return result, rows.Err()
}
```

- [ ] **Step 3: Add handlers and wire routes**

Add to `pkg/matching/handler.go`:

```go
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

func (h *Handler) GetMatchPairsByMarket(w http.ResponseWriter, r *http.Request) {
    marketID := r.URL.Query().Get("market_id")
    if marketID == "" {
        http.Error(w, "market_id param required", http.StatusBadRequest)
        return
    }

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
```

Add `"strconv"` import if not already present. Wire in `WireRoutes`:

```go
mux.HandleFunc("GET /api/v1/matching/markets/search", h.SearchMarkets)
```

The `GetMatchPairsByMarket` handler is added on the existing `/api/v1/matching/pairs` route alongside the existing `GetMatchPairs`. Since Go 1.22 ServeMux uses prefix matching for the same method+path, we need to differentiate by query param in a single handler. Update the existing `GetMatchPairs` to check for the `market_id` param:

```go
func (h *Handler) GetMatchPairs(w http.ResponseWriter, r *http.Request) {
    if marketID := r.URL.Query().Get("market_id"); marketID != "" {
        h.GetMatchPairsByMarket(w, r)
        return
    }
    // ... existing code
}
```

- [ ] **Step 4: Build check**

Run: `cd Arby && go build ./cmd/polybot`
Expected: Build succeeds

- [ ] **Step 5: Commit**

```bash
git add pkg/matching/store.go pkg/matching/handler.go
git commit -m "feat: add market search and pairs-by-market endpoints"
```

---

### Task 4: Embed script endpoint

**Files:**
- Modify: `pkg/matching/handler.go` — add embed script handler
- Create: `tools/embed_worker.py` — the actual Python script

**Interfaces:**
- Produces: `GET /api/v1/matching/embed-script` → Python file as `text/x-python` attachment

- [ ] **Step 1: Create the embed worker script**

Create `tools/embed_worker.py` (see Task 6 for full content — this is a placeholder reference). The script is a standalone Python tool that:
1. Fetches `GET /api/v1/markets/unembedded?limit=N` from a configurable host
2. Loads `BAAI/bge-large-en-v1.5` via sentence-transformers
3. POSTs embeddings back to `POST /api/v1/markets/embeddings`
4. Loops until the backlog is clear

- [ ] **Step 2: Embed script content in Go file**

Create `pkg/matching/embed_worker_data.go` containing the full script as a const string. The script is defined inline in the Go file — no `//go:embed` path issues.

Add handler in `pkg/matching/handler.go` referencing the const:

```go
func (h *Handler) GetEmbedScript(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/x-python")
    w.Header().Set("Content-Disposition", "attachment; filename=embed_worker.py")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(embedWorkerScript))
}
```

Wire in `WireRoutes`:

```go
mux.HandleFunc("GET /api/v1/matching/embed-script", h.GetEmbedScript)
```

Note: `tools/embed_worker.py` is kept as a user-facing convenience copy. The canonical source is `pkg/matching/embed_worker_data.go`. Both must stay in sync.

- [ ] **Step 3: Build check**

Run: `cd Arby && go build ./cmd/polybot`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add pkg/matching/handler.go tools/embed_worker.py
git commit -m "feat: add embed script download endpoint"
```

---

### Task 5: Update CandidateDiscoverer and Reviewer to use configurable threshold

**Files:**
- Modify: `pkg/matching/candidate.go`
- Modify: `pkg/matching/reviewer.go`

**Interfaces:**
- Consumes: `Store.GetSimilarityThreshold(ctx) (float64, error)`

- [ ] **Step 1: Update CandidateDiscoverer.RunOnce**

In `pkg/matching/candidate.go`, modify `RunOnce` to read threshold from store instead of using the hardcoded field:

```go
func (d *CandidateDiscoverer) RunOnce(ctx context.Context) error {
    threshold, err := d.store.GetSimilarityThreshold(ctx)
    if err != nil {
        threshold = d.similarityThreshold // fallback to constructor default
    }

    // ... rest of the function, change references from d.similarityThreshold to threshold
```

Find every use of `d.similarityThreshold` in `RunOnce` and replace with `threshold` (the local variable).

- [ ] **Step 2: Update Reviewer.RunOnce**

In `pkg/matching/reviewer.go`, modify `RunOnce` to filter pending candidates by similarity threshold:

```go
func (r *Reviewer) RunOnce(ctx context.Context) error {
    threshold, err := r.store.GetSimilarityThreshold(ctx)
    if err != nil {
        threshold = r.confidenceThreshold // fallback
    }

    candidates, err := r.store.GetPendingCandidatesWithMarkets(ctx, r.batchSize*4)
    // ... existing code ...
```

After fetching candidates, filter them:

```go
var filtered []matching.CandidateWithMarkets
for _, c := range candidates {
    if c.Similarity >= threshold {
        filtered = append(filtered, c)
    }
}
candidates = filtered
```

Or better, modify `GetPendingCandidatesWithMarkets` to accept a threshold parameter. Update the store method signature:

```go
func (s *Store) GetPendingCandidatesWithMarkets(ctx context.Context, limit int, minSimilarity float64) ([]CandidateWithMarkets, error)
```

And update the SQL to include `AND mc.similarity >= $2`.

- [ ] **Step 3: Build check**

Run: `cd Arby && go build ./cmd/polybot`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add pkg/matching/candidate.go pkg/matching/reviewer.go pkg/matching/store.go
git commit -m "feat: read similarity threshold from config in discoverer and reviewer"
```

---

### Task 6: Python embed worker script

**Files:**
- Create: `tools/embed_worker.py`

- [ ] **Step 1: Write `tools/embed_worker.py`**

Full script content:

```python
#!/usr/bin/env python3
"""
embed_worker.py — standalone embedding worker for Arby

Fetches unembedded markets via the Arby REST API, embeds them locally
using BAAI/bge-large-en-v1.5, and writes embeddings back via the API.

Usage:
    pip install requests sentence-transformers torch
    python embed_worker.py --host https://arby.nostrabotus.com --batch-size 64
"""

from __future__ import annotations

import argparse
import logging
import sys
import time
from typing import Any

import requests
from sentence_transformers import SentenceTransformer

logging.basicConfig(
    format="%(asctime)s  %(levelname)-8s  %(message)s",
    datefmt="%H:%M:%S",
    level=logging.INFO,
)
log = logging.getLogger("embed_worker")

MODEL_NAME = "BAAI/bge-large-en-v1.5"


def load_model(batch_size: int) -> SentenceTransformer:
    log.info("Loading %s (batch_size=%d) …", MODEL_NAME, batch_size)
    model = SentenceTransformer(MODEL_NAME)
    log.info("Model ready on %s", model.device)
    return model


def fetch_unembedded(host: str, limit: int) -> list[dict[str, Any]]:
    resp = requests.get(
        f"{host}/api/v1/markets/unembedded",
        params={"limit": limit},
        timeout=30,
    )
    resp.raise_for_status()
    data = resp.json()
    return data.get("markets", [])


def post_embeddings(host: str, embeddings: list[dict[str, Any]]) -> int:
    resp = requests.post(
        f"{host}/api/v1/markets/embeddings",
        json={"embeddings": embeddings},
        timeout=60,
    )
    resp.raise_for_status()
    return resp.json().get("updated", 0)


def main() -> None:
    parser = argparse.ArgumentParser(description="Arby embedding worker")
    parser.add_argument("--host", default="http://localhost:8087", help="Arby API base URL")
    parser.add_argument("--batch-size", type=int, default=64, help="Embedding batch size")
    parser.add_argument("--fetch-limit", type=int, default=2000, help="Markets per fetch")
    parser.add_argument("--sleep", type=int, default=10, help="Seconds to wait when queue is empty")
    args = parser.parse_args()

    host = args.host.rstrip("/")
    model = load_model(args.batch_size)

    total_embedded = 0
    empty_passes = 0

    log.info("Starting embed worker against %s", host)

    while True:
        try:
            batch = fetch_unembedded(host, args.fetch_limit)
        except requests.RequestException as e:
            log.warning("Fetch failed: %s — retrying in 30s", e)
            time.sleep(30)
            continue

        if not batch:
            empty_passes += 1
            log.info(
                "No unembedded markets (pass %d). Total: %d. Sleeping %ds …",
                empty_passes, total_embedded, args.sleep,
            )
            time.sleep(args.sleep)
            continue

        empty_passes = 0
        log.info("Fetched %d unembedded markets", len(batch))

        texts = []
        for m in batch:
            title = (m.get("title") or "").strip()
            desc = (m.get("description") or "").strip()
            if desc:
                texts.append(f"{title}. {desc[:400]}")
            else:
                texts.append(title)

        log.info("Encoding %d texts …", len(texts))
        vectors = model.encode(texts, batch_size=args.batch_size, normalize_embeddings=True, show_progress_bar=True)

        embeddings = [
            {"id": m["id"], "vector": v.tolist()}
            for m, v in zip(batch, vectors)
        ]

        try:
            updated = post_embeddings(host, embeddings)
        except requests.RequestException as e:
            log.warning("Upload failed: %s — retrying in 30s", e)
            time.sleep(30)
            continue

        total_embedded += updated
        log.info("Wrote %d embeddings (session total: %d)", updated, total_embedded)

        if updated < len(batch):
            log.warning("Expected %d updates, got %d — some may have failed", len(batch), updated)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        log.info("Interrupted")
        sys.exit(0)
```

- [ ] **Step 2: Commit**

```bash
git add tools/embed_worker.py
git commit -m "feat: add embed worker script"
```

---

### Task 7: UI Sidebar restructuring

**Files:**
- Create: `services/ui/src/app/matching/pipeline/page.tsx`
- Create: `services/ui/src/app/matching/review/page.tsx`
- Create: `services/ui/src/app/matching/settings/page.tsx`
- Modify: `services/ui/src/components/Sidebar.tsx`

- [ ] **Step 1: Create placeholder matching pages**

Create `services/ui/src/app/matching/pipeline/page.tsx`:
```tsx
"use client"

export default function PipelinePage() {
  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-100">Pipeline</h1>
      <p className="text-sm text-muted mt-1">Loading pipeline data…</p>
    </div>
  )
}
```

Create `services/ui/src/app/matching/review/page.tsx`:
```tsx
"use client"

export default function MatchingReviewPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-100">Review</h1>
      <p className="text-sm text-muted mt-1">Loading pairs…</p>
    </div>
  )
}
```

Create `services/ui/src/app/matching/settings/page.tsx`:
```tsx
"use client"

export default function MatchingSettingsPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-100">Matching Settings</h1>
      <p className="text-sm text-muted mt-1">Loading settings…</p>
    </div>
  )
}
```

- [ ] **Step 2: Update sidebar with collapsible Matching group**

Replace `services/ui/src/components/Sidebar.tsx`:

```tsx
"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { useState } from "react"

const navItems = [
  { label: "Dashboard", href: "/", icon: GridIcon },
  { label: "Markets", href: "/markets", icon: ChartIcon },
  { label: "Strategies", href: "/strategies", icon: BundleIcon },
  { label: "Bundles", href: "/bundles", icon: BundleIcon },
  { label: "Settings", href: "/settings", icon: GearIcon },
]

const matchingSubItems = [
  { label: "Pipeline", href: "/matching/pipeline" },
  { label: "Review", href: "/matching/review" },
  { label: "Settings", href: "/matching/settings" },
]

export function Sidebar() {
  const path = usePathname()
  const isMatchingActive = path.startsWith("/matching")
  const [matchingOpen, setMatchingOpen] = useState(isMatchingActive)

  return (
    <aside className="w-16 lg:w-56 bg-surface-alt border-r border-border flex flex-col shrink-0">
      <div className="h-14 flex items-center justify-center lg:justify-start lg:px-5 border-b border-border">
        <span className="text-accent font-bold text-lg hidden lg:inline">Arby</span>
        <span className="text-accent font-bold text-lg lg:hidden">A</span>
      </div>
      <nav className="flex-1 flex flex-col gap-1 p-2">
        {navItems.map((item) => {
          const active = path === item.href
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors ${
                active
                  ? "bg-accent/10 text-accent"
                  : "text-muted hover:text-gray-200 hover:bg-surface-hover"
              }`}
            >
              <item.icon active={active} />
              <span className="hidden lg:inline">{item.label}</span>
            </Link>
          )
        })}

        {/* Matching collapsible group */}
        <div className="mt-1">
          <button
            onClick={() => setMatchingOpen(!matchingOpen)}
            className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors ${
              isMatchingActive
                ? "bg-accent/10 text-accent"
                : "text-muted hover:text-gray-200 hover:bg-surface-hover"
            }`}
          >
            <LayersIcon active={isMatchingActive} />
            <span className="hidden lg:inline flex-1 text-left">Matching</span>
            <svg
              className={`hidden lg:block w-3 h-3 transition-transform ${matchingOpen ? "rotate-180" : ""}`}
              fill="none" stroke="currentColor" viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>

          {matchingOpen && (
            <div className="ml-3 mt-1 flex flex-col gap-0.5 border-l border-border pl-2">
              {matchingSubItems.map((item) => {
                const active = path === item.href
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`block px-3 py-2 rounded-lg text-sm transition-colors ${
                      active
                        ? "bg-accent/10 text-accent"
                        : "text-muted hover:text-gray-200 hover:bg-surface-hover"
                    }`}
                  >
                    <span className="hidden lg:inline">{item.label}</span>
                  </Link>
                )
              })}
            </div>
          )}
        </div>
      </nav>
      <div className="p-2 border-t border-border">
        <div className="flex items-center gap-3 px-3 py-2 text-xs text-muted">
          <div className="w-6 h-6 rounded-full bg-accent/20 flex items-center justify-center text-accent text-xs font-bold">A</div>
          <span className="hidden lg:inline truncate">Arby v0.1</span>
        </div>
      </div>
    </aside>
  )
}

// Icon components (GridIcon, ChartIcon, BundleIcon, GearIcon, LayersIcon)
// Keep existing GridIcon, ChartIcon, BundleIcon, GearIcon unchanged

function LayersIcon({ active }: { active: boolean }) {
  return (
    <svg className="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={active ? 2 : 1.5}
        d="M2.25 15.75l5.159 5.159a2.25 2.25 0 003.182 0l5.159-5.159m-1.5-1.5l1.409-1.41a2.25 2.25 0 000-3.182l-5.159-5.159a2.25 2.25 0 00-3.182 0L6.75 8.25M21 12V9m-9 0h9" />
    </svg>
  )
}
```

- [ ] **Step 3: Build check UI**

Run: `cd services/ui && npm run build`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add services/ui/src/components/Sidebar.tsx services/ui/src/app/matching/
git commit -m "feat: add Matching sidebar group with placeholder sub-pages"
```

---

### Task 8: Pipeline page

**Files:**
- Modify: `services/ui/src/app/matching/pipeline/page.tsx`

- [ ] **Step 1: Implement Pipeline page with funnel**

Replace `services/ui/src/app/matching/pipeline/page.tsx`:

```tsx
"use client"

import { useState, useEffect, useCallback } from "react"

type StageStats = {
  unembedded: number
  embedded: number
  pending_candidates: number
  reviewed_candidates: number
  pairs_pending_approval: number
  pairs_approved: number
  pairs_rejected: number
}

type StageCard = {
  key: keyof StageStats
  label: string
  description: string
}

const stages: StageCard[] = [
  { key: "unembedded", label: "Unembedded", description: "Markets awaiting embedding" },
  { key: "embedded", label: "Embedded", description: "Markets with vector embeddings" },
  { key: "pending_candidates", label: "Candidates", description: "Similar pairs found by ANN" },
  { key: "reviewed_candidates", label: "LLM Reviewed", description: "Pairs analyzed by LLM" },
  { key: "pairs_pending_approval", label: "Pending Approval", description: "Awaiting human review" },
]

const summaryStages: StageCard[] = [
  { key: "pairs_approved", label: "Approved", description: "Human-approved pairs" },
  { key: "pairs_rejected", label: "Rejected", description: "Human-rejected pairs" },
]

export default function PipelinePage() {
  const [stats, setStats] = useState<StageStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(")

  const fetchStats = useCallback(async () => {
    try {
      const res = await fetch("/api/matching/stats")
      if (!res.ok) throw new Error("Failed to fetch")
      const data = await res.json()
      setStats(data)
      setError("")
    } catch (e) {
      setError("Failed to load pipeline stats")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchStats()
    const interval = setInterval(fetchStats, 30000)
    return () => clearInterval(interval)
  }, [fetchStats])

  const getStageIcon = (key: keyof StageStats, value: number) => {
    if (key === "unembedded" && value > 0) return "⚠️"
    if (key === "embedded") return "✅"
    if (key === "pending_candidates" && value > 0) return "⏳"
    return "✅"
  }

  if (loading) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <h1 className="text-2xl font-bold text-gray-100">Pipeline</h1>
        <p className="text-sm text-muted mt-4 animate-pulse">Loading pipeline data…</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <h1 className="text-2xl font-bold text-gray-100">Pipeline</h1>
        <p className="text-sm text-red-400 mt-4">{error}</p>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-100">Pipeline</h1>
        <p className="text-sm text-muted mt-1">Matching pipeline status — auto-refreshes every 30s</p>
      </div>

      {/* Funnel stages */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        {stages.map((stage, i) => {
          const value = stats ? stats[stage.key] : 0
          return (
            <div key={stage.key} className="relative bg-surface-alt rounded-xl border border-border p-5">
              {i < stages.length - 1 && (
                <div className="hidden lg:block absolute top-1/2 -right-3 w-6 h-6 text-muted">
                  <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </div>
              )}
              <div className="text-3xl font-bold text-gray-100 mb-1">{value.toLocaleString()}</div>
              <div className="text-sm font-semibold text-gray-200">{stage.label}</div>
              <div className="text-xs text-muted mt-1">{stage.description}</div>
            </div>
          )
        })}
      </div>

      {/* Summary row */}
      <div className="grid grid-cols-2 gap-4 max-w-md">
        {summaryStages.map((stage) => {
          const value = stats ? stats[stage.key] : 0
          return (
            <div key={stage.key} className="bg-surface-alt rounded-xl border border-border p-4">
              <div className={`text-2xl font-bold mb-1 ${
                stage.key === "pairs_approved" ? "text-green" : "text-red"
              }`}>
                {value.toLocaleString()}
              </div>
              <div className="text-sm font-semibold text-gray-200">{stage.label}</div>
              <div className="text-xs text-muted mt-1">{stage.description}</div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Add API proxy route**

Create `services/ui/src/app/api/matching/stats/route.ts`:

```ts
import { NextResponse } from "next/server"

const POLYBOT_HOST = process.env.POLYBOT_HOST || "http://polybot:8086"

export async function GET() {
  const res = await fetch(`${POLYBOT_HOST}/api/v1/matching/stats`)
  const data = await res.json()
  return NextResponse.json(data)
}
```

- [ ] **Step 3: Build check**

Run: `cd services/ui && npm run build`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add services/ui/src/app/matching/pipeline/page.tsx services/ui/src/app/api/matching/stats/route.ts
git commit -m "feat: implement pipeline page with funnel visualization"
```

---

### Task 9: Enhanced Review page

**Files:**
- Modify: `services/ui/src/app/matching/review/page.tsx`
- Create: `services/ui/src/app/api/matching/markets/search/route.ts`
- Create: `services/ui/src/app/api/matching/pairs/route.ts` (update existing)

- [ ] **Step 1: Create market search API proxy route**

Create `services/ui/src/app/api/matching/markets/search/route.ts`:

```ts
import { NextRequest, NextResponse } from "next/server"

const POLYBOT_HOST = process.env.POLYBOT_HOST || "http://polybot:8086"

export async function GET(req: NextRequest) {
  const venue = req.nextUrl.searchParams.get("venue")
  const q = req.nextUrl.searchParams.get("q")
  if (!venue || !q) {
    return NextResponse.json({ markets: [] })
  }
  const res = await fetch(
    `${POLYBOT_HOST}/api/v1/matching/markets/search?venue=${encodeURIComponent(venue)}&q=${encodeURIComponent(q)}&limit=10`,
  )
  const data = await res.json()
  return NextResponse.json(data)
}
```

- [ ] **Step 2: Update pairs API proxy route to pass all query params**

Update `services/ui/src/app/api/matching/pairs/route.ts` to forward all query params through:

```ts
import { NextRequest, NextResponse } from "next/server"

const GO_API_URL = process.env.GO_API_URL || "http://polybot:8086"

export async function GET(req: NextRequest) {
  const params = req.nextUrl.searchParams.toString()
  const url = `${GO_API_URL}/api/v1/matching/pairs${params ? `?${params}` : ""}`
  const res = await fetch(url)
  const data = await res.json()
  return NextResponse.json(data)
}
```

- [ ] **Step 3: Implement the enhanced review page**

Replace `services/ui/src/app/matching/review/page.tsx`:

```tsx
"use client"

import { useState, useEffect } from "react"

type MatchPair = {
  id: string
  candidate_id: string
  market_a_title: string
  market_b_title: string
  venue_a: string
  venue_b: string
  category: string
  is_same_event: boolean
  relationship: string
  confidence: number
  reasoning: string
  leg_a_model: string
  leg_b_model: string
  status: string
}

type MarketResult = {
  id: string
  venue: string
  venue_market_id: string
  title: string
  description: string
  category: string
}

export default function MatchingReviewPage() {
  const [pairs, setPairs] = useState<MatchPair[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  // Search state
  const [searchVenue, setSearchVenue] = useState("KALSHI")
  const [searchQuery, setSearchQuery] = useState("")
  const [searchResults, setSearchResults] = useState<MarketResult[]>([])
  const [searching, setSearching] = useState(false)
  const [selectedMarket, setSelectedMarket] = useState<MarketResult | null>(null)
  const [marketPairs, setMarketPairs] = useState<MatchPair[]>([])
  const [marketPairsLoading, setMarketPairsLoading] = useState(false)

  // Load all pending pairs on mount
  useEffect(() => {
    fetch("/api/matching/pairs?status=PENDING_APPROVAL")
      .then((r) => r.json())
      .then((data) => setPairs(data.pairs ?? []))
      .catch(() => setError("Failed to load"))
      .finally(() => setLoading(false))
  }, [])

  const handleApprove = async (id: string) => {
    setActionLoading(id)
    try {
      await fetch("/api/matching/pairs/approve", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id }),
      })
      setPairs((prev) => prev.filter((p) => p.id !== id))
      setMarketPairs((prev) =>
        prev.map((p) => (p.id === id ? { ...p, status: "APPROVED" } : p))
      )
    } catch {
      setError("Failed to approve")
    } finally {
      setActionLoading(null)
    }
  }

  const handleReject = async (id: string) => {
    setActionLoading(id)
    try {
      await fetch("/api/matching/pairs/reject", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id }),
      })
      setPairs((prev) => prev.filter((p) => p.id !== id))
      setMarketPairs((prev) =>
        prev.map((p) => (p.id === id ? { ...p, status: "REJECTED" } : p))
      )
    } catch {
      setError("Failed to reject")
    } finally {
      setActionLoading(null)
    }
  }

  const handleSearch = async () => {
    if (!searchQuery.trim()) return
    setSearching(true)
    setSelectedMarket(null)
    setMarketPairs([])
    try {
      const res = await fetch(
        `/api/matching/markets/search?venue=${encodeURIComponent(searchVenue)}&q=${encodeURIComponent(searchQuery)}`
      )
      const data = await res.json()
      setSearchResults(data.markets ?? [])
    } catch {
      setError("Search failed")
    } finally {
      setSearching(false)
    }
  }

  const handleSelectMarket = async (market: MarketResult) => {
    setSelectedMarket(market)
    setMarketPairsLoading(true)
    try {
      const res = await fetch(`/api/matching/pairs?market_id=${market.id}`)
      const data = await res.json()
      setMarketPairs(data.pairs ?? [])
    } catch {
      setError("Failed to load market pairs")
    } finally {
      setMarketPairsLoading(false)
    }
  }

  const confidenceColor = (c: number) => {
    if (c >= 0.9) return "text-green"
    if (c >= 0.7) return "text-amber"
    return "text-muted"
  }

  const relationshipBadge = (r: string) => {
    if (r === "EQUIVALENT") return "bg-green/20 text-green"
    if (r === "INVERSE") return "bg-amber/20 text-amber"
    return "bg-gray-500/20 text-gray-400"
  }

  const statusBadge = (s: string) => {
    if (s === "PENDING_APPROVAL") return "bg-amber/20 text-amber"
    if (s === "APPROVED") return "bg-green/20 text-green"
    if (s === "REJECTED") return "bg-red/20 text-red"
    return "bg-gray-500/20 text-gray-400"
  }

  const venueBadge = (v: string) => {
    if (v === "KALSHI") return "bg-accent/20 text-accent"
    if (v === "POLYMARKET") return "bg-amber/20 text-amber"
    return "bg-gray-500/20 text-gray-400"
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-100">Review</h1>
        <p className="text-sm text-muted mt-1">
          Search for a market to see all its comparisons, or review pending pairs below
        </p>
      </div>

      {/* Search bar */}
      <div className="flex items-center gap-3">
        <select
          value={searchVenue}
          onChange={(e) => setSearchVenue(e.target.value)}
          className="bg-surface border border-border rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
        >
          <option value="KALSHI">Kalshi</option>
          <option value="POLYMARKET">Polymarket</option>
        </select>
        <input
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSearch()}
          placeholder="Search by market title..."
          className="flex-1 bg-surface border border-border rounded-lg px-3 py-2 text-sm text-gray-200 placeholder-muted outline-none"
        />
        <button
          onClick={handleSearch}
          disabled={searching}
          className="px-4 py-2 text-sm rounded-lg bg-accent text-white hover:bg-accent-hover disabled:opacity-40 transition-colors"
        >
          {searching ? "..." : "Search"}
        </button>
      </div>

      {/* Search results */}
      {searchResults.length > 0 && (
        <div className="bg-surface-alt rounded-xl border border-border overflow-hidden">
          <div className="px-4 py-3 border-b border-border text-xs text-muted uppercase tracking-wider font-semibold">
            Search results
          </div>
          <div className="divide-y divide-border">
            {searchResults.map((m) => (
              <button
                key={m.id}
                onClick={() => handleSelectMarket(m)}
                className={`w-full text-left px-4 py-3 hover:bg-surface/50 transition-colors ${
                  selectedMarket?.id === m.id ? "bg-accent/5" : ""
                }`}
              >
                <div className="flex items-center gap-2">
                  <span className={`text-xs px-1.5 py-0.5 rounded-full ${venueBadge(m.venue)}`}>
                    {m.venue}
                  </span>
                  <span className="text-sm text-gray-200 truncate">{m.title}</span>
                </div>
                <p className="text-xs text-muted mt-1 truncate">{m.description}</p>
              </button>
            ))}
          </div>
        </div>
      )}

      {searchResults.length === 0 && searchQuery && !searching && (
        <p className="text-sm text-muted">No markets found for "{searchQuery}"</p>
      )}

      {/* Market detail: pairs involving this market */}
      {selectedMarket && (
        <div className="bg-surface-alt rounded-xl border border-border overflow-hidden">
          <div className="px-4 py-3 border-b border-border flex items-center justify-between">
            <span className="text-xs text-muted uppercase tracking-wider font-semibold">
              Comparisons for {selectedMarket.title}
            </span>
            <span className="text-xs text-muted">{marketPairs.length} pairs</span>
          </div>

          {marketPairsLoading && (
            <div className="px-4 py-8 text-sm text-muted text-center animate-pulse">
              Loading pairs…
            </div>
          )}

          {!marketPairsLoading && marketPairs.length === 0 && (
            <div className="px-4 py-8 text-sm text-muted text-center">
              No comparisons found for this market.
            </div>
          )}

          {!marketPairsLoading && marketPairs.length > 0 && (
            <table className="w-full text-sm text-left border-collapse">
              <thead>
                <tr className="border-b border-border text-xs text-muted uppercase tracking-wider">
                  <th className="px-4 py-3 font-medium">Market A</th>
                  <th className="px-4 py-3 font-medium">Market B</th>
                  <th className="px-4 py-3 font-medium">Relationship</th>
                  <th className="px-4 py-3 font-medium">Confidence</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Reasoning</th>
                  <th className="px-4 py-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {marketPairs.map((p) => (
                  <tr key={p.id} className="border-b border-border last:border-0 hover:bg-surface/50 transition-colors">
                    <td className="px-4 py-3">
                      <div className="text-gray-200 max-w-[200px] truncate" title={p.market_a_title}>{p.market_a_title}</div>
                      <span className={`inline-block text-[10px] px-1.5 py-0.5 rounded-full mt-1 ${venueBadge(p.venue_a)}`}>{p.venue_a}</span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="text-gray-200 max-w-[200px] truncate" title={p.market_b_title}>{p.market_b_title}</div>
                      <span className={`inline-block text-[10px] px-1.5 py-0.5 rounded-full mt-1 ${venueBadge(p.venue_b)}`}>{p.venue_b}</span>
                    </td>
                    <td className="px-4 py-3">
                      {p.relationship && (
                        <span className={`text-xs px-2 py-0.5 rounded-full ${relationshipBadge(p.relationship)}`}>{p.relationship}</span>
                      )}
                      {!p.relationship && <span className="text-xs text-muted">—</span>}
                    </td>
                    <td className={`px-4 py-3 text-xs font-semibold ${confidenceColor(p.confidence)}`}>
                      {p.confidence > 0 ? `${(p.confidence * 100).toFixed(0)}%` : "—"}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-0.5 rounded-full ${statusBadge(p.status)}`}>{p.status}</span>
                    </td>
                    <td className="px-4 py-3 text-xs text-muted max-w-[200px] truncate" title={p.reasoning}>
                      {p.reasoning || "Waiting for LLM review"}
                    </td>
                    <td className="px-4 py-3">
                      {p.status === "PENDING_APPROVAL" && (
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => handleApprove(p.id)}
                            disabled={actionLoading === p.id}
                            className="px-3 py-1 text-xs rounded-lg bg-green/20 text-green hover:bg-green/30 disabled:opacity-40 transition-colors"
                          >
                            {actionLoading === p.id ? "..." : "Approve"}
                          </button>
                          <button
                            onClick={() => handleReject(p.id)}
                            disabled={actionLoading === p.id}
                            className="px-3 py-1 text-xs rounded-lg bg-red/20 text-red hover:bg-red/30 disabled:opacity-40 transition-colors"
                          >
                            {actionLoading === p.id ? "..." : "Reject"}
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Pending pairs section */}
      <div>
        <h2 className="text-sm font-semibold text-gray-200 mb-3">
          Pending Approval ({pairs.length})
        </h2>

        {loading && <div className="text-sm text-muted animate-pulse">Loading pairs...</div>}
        {error && <div className="text-sm text-red-400">{error}</div>}

        {!loading && !error && pairs.length === 0 && (
          <div className="flex flex-col items-center justify-center py-16 text-muted">
            <svg className="w-12 h-12 mb-4 opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p className="text-sm">No pending pairs to review</p>
            <p className="text-xs mt-1">New pairs will appear here once discovered and reviewed by the AI.</p>
          </div>
        )}

        {pairs.length > 0 && (
          <div className="bg-surface-alt rounded-xl border border-border overflow-hidden">
            <table className="w-full text-sm text-left border-collapse">
              <thead>
                <tr className="border-b border-border text-xs text-muted uppercase tracking-wider">
                  <th className="px-4 py-3 font-medium">Market A</th>
                  <th className="px-4 py-3 font-medium">Market B</th>
                  <th className="px-4 py-3 font-medium">Rel.</th>
                  <th className="px-4 py-3 font-medium">Conf.</th>
                  <th className="px-4 py-3 font-medium">Models</th>
                  <th className="px-4 py-3 font-medium">Reasoning</th>
                  <th className="px-4 py-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {pairs.map((p) => (
                  <tr key={p.id} className="border-b border-border last:border-0 hover:bg-surface/50 transition-colors">
                    <td className="px-4 py-3">
                      <div className="text-gray-200 max-w-[200px] truncate" title={p.market_a_title}>{p.market_a_title}</div>
                      <span className={`inline-block text-[10px] px-1.5 py-0.5 rounded-full mt-1 ${venueBadge(p.venue_a)}`}>{p.venue_a}</span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="text-gray-200 max-w-[200px] truncate" title={p.market_b_title}>{p.market_b_title}</div>
                      <span className={`inline-block text-[10px] px-1.5 py-0.5 rounded-full mt-1 ${venueBadge(p.venue_b)}`}>{p.venue_b}</span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-0.5 rounded-full ${relationshipBadge(p.relationship)}`}>{p.relationship}</span>
                    </td>
                    <td className={`px-4 py-3 text-xs font-semibold ${confidenceColor(p.confidence)}`}>
                      {(p.confidence * 100).toFixed(0)}%
                    </td>
                    <td className="px-4 py-3 text-xs text-muted">
                      <span className="font-mono">{p.leg_a_model}</span>
                      <span className="mx-1">→</span>
                      <span className="font-mono">{p.leg_b_model}</span>
                    </td>
                    <td className="px-4 py-3 text-xs text-muted max-w-[200px] truncate" title={p.reasoning}>{p.reasoning}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => handleApprove(p.id)}
                          disabled={actionLoading === p.id}
                          className="px-3 py-1 text-xs rounded-lg bg-green/20 text-green hover:bg-green/30 disabled:opacity-40 transition-colors"
                        >
                          {actionLoading === p.id ? "..." : "Approve"}
                        </button>
                        <button
                          onClick={() => handleReject(p.id)}
                          disabled={actionLoading === p.id}
                          className="px-3 py-1 text-xs rounded-lg bg-red/20 text-red hover:bg-red/30 disabled:opacity-40 transition-colors"
                        >
                          {actionLoading === p.id ? "..." : "Reject"}
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Build check**

Run: `cd services/ui && npm run build`
Expected: Build succeeds

- [ ] **Step 5: Commit**

```bash
git add services/ui/src/app/matching/review/page.tsx services/ui/src/app/api/matching/markets/search/route.ts services/ui/src/app/api/matching/pairs/route.ts
git commit -m "feat: implement enhanced review page with market search"
```

---

### Task 10: Settings page additions

**Files:**
- Create: `services/ui/src/app/matching/settings/page.tsx` update
- Create: `services/ui/src/app/api/matching/settings/route.ts`

- [ ] **Step 1: Create settings API proxy route**

Create `services/ui/src/app/api/matching/settings/route.ts`:

```ts
import { NextRequest, NextResponse } from "next/server"

const POLYBOT_HOST = process.env.POLYBOT_HOST || "http://polybot:8086"

export async function GET() {
  const res = await fetch(`${POLYBOT_HOST}/api/v1/matching/settings`)
  const data = await res.json()
  return NextResponse.json(data)
}

export async function POST(req: NextRequest) {
  const body = await req.json()
  const res = await fetch(`${POLYBOT_HOST}/api/v1/matching/settings`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  const data = await res.json()
  return NextResponse.json(data)
}
```

- [ ] **Step 2: Implement matching settings page with threshold slider + download button**

Replace `services/ui/src/app/matching/settings/page.tsx`:

```tsx
"use client"

import { useState, useEffect } from "react"

export default function MatchingSettingsPage() {
  const [threshold, setThreshold] = useState(0.8)
  const [originalThreshold, setOriginalThreshold] = useState(0.8)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    fetch("/api/matching/settings")
      .then((r) => r.json())
      .then((data) => {
        const t = data.similarity_threshold ?? 0.8
        setThreshold(t)
        setOriginalThreshold(t)
      })
      .catch(() => setError("Failed to load settings"))
      .finally(() => setLoading(false))
  }, [])

  const handleSave = async () => {
    setSaving(true)
    setSaved(false)
    setError("")
    try {
      const res = await fetch("/api/matching/settings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ similarity_threshold: threshold }),
      })
      if (!res.ok) throw new Error("Save failed")
      setOriginalThreshold(threshold)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch {
      setError("Failed to save settings")
    } finally {
      setSaving(false)
    }
  }

  const handleDownload = async () => {
    try {
      const res = await fetch("/api/matching/embed-script")
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = "embed_worker.py"
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch {
      setError("Failed to download script")
    }
  }

  if (loading) {
    return (
      <div className="p-6 max-w-3xl mx-auto">
        <h1 className="text-2xl font-bold text-gray-100">Matching Settings</h1>
        <p className="text-sm text-muted mt-4 animate-pulse">Loading settings…</p>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-3xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-100">Matching Settings</h1>
        <p className="text-sm text-muted mt-1">Configure the matching pipeline behavior</p>
      </div>

      {error && (
        <div className="text-sm text-red-400 bg-red/10 border border-red/20 rounded-lg px-4 py-3">{error}</div>
      )}

      {/* Similarity Threshold */}
      <div className="bg-surface-alt rounded-xl border border-border p-6">
        <h2 className="text-base font-semibold text-gray-100 mb-1">Similarity Threshold</h2>
        <p className="text-sm text-muted mb-4">
          Markets with similarity below this threshold will not be sent for LLM comparison.
          Changes take effect on the next discovery cycle.
        </p>

        <div className="flex items-center gap-4">
          <input
            type="range"
            min="0.5"
            max="1"
            step="0.01"
            value={threshold}
            onChange={(e) => setThreshold(parseFloat(e.target.value))}
            className="flex-1 accent-accent"
          />
          <span className="text-lg font-bold text-gray-100 w-16 text-right tabular-nums">
            {(threshold * 100).toFixed(0)}%
          </span>
        </div>

        <div className="flex items-center gap-2 mt-4">
          <button
            onClick={handleSave}
            disabled={saving || threshold === originalThreshold}
            className="px-4 py-2 text-sm rounded-lg bg-accent text-white hover:bg-accent-hover disabled:opacity-40 transition-colors"
          >
            {saving ? "Saving..." : "Save"}
          </button>
          {saved && <span className="text-xs text-green">Saved</span>}
        </div>
      </div>

      {/* Embed Worker Download */}
      <div className="bg-surface-alt rounded-xl border border-border p-6">
        <h2 className="text-base font-semibold text-gray-100 mb-1">Embed Worker</h2>
        <p className="text-sm text-muted mb-4">
          Download the standalone embedding script to run on your local machine.
          It uses BAAI/bge-large-en-v1.5 via sentence-transformers.
        </p>
        <p className="text-xs text-muted mb-4">
          Usage: <code className="text-accent bg-surface px-1.5 py-0.5 rounded">python embed_worker.py --host https://arby.nostrabotus.com</code>
        </p>
        <button
          onClick={handleDownload}
          className="px-4 py-2 text-sm rounded-lg bg-accent text-white hover:bg-accent-hover transition-colors"
        >
          Download embed_worker.py
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Build check**

Run: `cd services/ui && npm run build`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add services/ui/src/app/matching/settings/page.tsx services/ui/src/app/api/matching/settings/route.ts
git commit -m "feat: add matching settings page with threshold slider and script download"
```

---

### Task 11: Deploy to EC2

**Files:** N/A — infrastructure

- [ ] **Step 1: Copy code to EC2**

```bash
# from project root
scp -i ~/.ssh/ArbAgentKeyPair1.pem -r * ec2-user@3.16.140.129:~/Arby/
```

- [ ] **Step 2: Rebuild and restart**

```bash
ssh -i ~/.ssh/ArbAgentKeyPair1.pem ec2-user@3.16.140.129 "cd ~/Arby && docker compose build polybot && docker compose up -d polybot"
```

- [ ] **Step 3: Rebuild UI**

```bash
ssh -i ~/.ssh/ArbAgentKeyPair1.pem ec2-user@3.16.140.129 "cd ~/Arby && docker compose build ui && docker compose up -d ui"
```

- [ ] **Step 4: Verify endpoints**

```bash
curl https://arby.nostrabotus.com/api/v1/matching/stats
curl https://arby.nostrabotus.com/api/v1/matching/settings
curl https://arby.nostrabotus.com/api/v1/matching/embed-script -o embed_worker.py
curl "https://arby.nostrabotus.com/api/v1/matching/markets/search?venue=KALSHI&q=federal"
```

- [ ] **Step 5: Verify UI pages load**
- Navigate to Pipeline, Review, Settings sub-pages in browser
