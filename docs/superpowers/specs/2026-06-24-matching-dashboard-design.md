# Matching Dashboard — Design Spec

## Overview

Replace the current single Review page with a "Matching" section in the sidebar containing three sub-pages: Pipeline, Review, and Settings. Add new Go backend endpoints to support pipeline statistics, market search, dynamic similarity threshold configuration, and an embed-worker script download.

## Sidebar Navigation

Current: `Dashboard | Markets | Strategies | Review | Bundles | Settings`

New: `Dashboard | Markets | Strategies | Matching (▼) > Pipeline | Matching > Review | Matching > Settings | Bundles | Settings`

The Matching nav item is a click-to-expand group. Clicking "Matching" expands/collapses three sub-links below it. The "Review" top-level nav item is replaced by the sub-page under Matching. The previous "Settings" top-level nav remains separate (not moved under Matching). The Matching item uses a LayersIcon (three stacked squares).

When sidebar is collapsed to `w-16`, the Matching label shows "M" and sub-links are hidden. Hovering the icon shows the group label in a tooltip but sub-links remain inaccessible until expanded — acceptable since the collapsed state is a shortcut view.

## New Go Backend Endpoints

All under the `pkg/matching/handler.go` handler, wired via `WireRoutes`.

### `GET /api/v1/matching/stats`

Returns pipeline stage counts:

```json
{
  "unembedded": 1200,
  "embedded": 13800,
  "pending_candidates": 45,
  "reviewed_candidates": 230,
  "pairs_pending_approval": 89,
  "pairs_approved": 120,
  "pairs_rejected": 21
}
```

**SQL:** Aggregates from `markets`, `match_candidates`, and `match_pairs` tables. Unembedded = `markets WHERE embedding IS NULL AND status='OPEN'`. Embedded = `markets WHERE embedding IS NOT NULL AND status='OPEN'`. Pending candidates = `match_candidates WHERE status='PENDING'`. Reviewed = `WHERE status='REVIEWED'`.

### `GET /api/v1/matching/settings`

```json
{
  "similarity_threshold": 0.80
}
```

Values come from a new `matching_config` table (simple key/value, or a single-row config table).

### `POST /api/v1/matching/settings`

Body: `{"similarity_threshold": 0.85}`

Updates the config row. The CandidateDiscoverer reads this value on each cycle instead of using the hardcoded default. The `NewCandidateDiscoverer` constructor still gets a default, but `RunOnce` checks the store for an override.

### `GET /api/v1/matching/pairs?market_id=<uuid>`

Returns match pairs filtered to those involving the given market ID (as either market A or market B). Same response shape as the existing unfiltered endpoint.

### `GET /api/v1/matching/markets/search?venue=KALSHI&q=federal+funds+rate`

Full-text search on markets table filtered by venue. Returns:

```json
{
  "markets": [
    {
      "id": "uuid",
      "venue": "KALSHI",
      "venue_market_id": "...",
      "title": "...",
      "description": "...",
      "category": "Macro"
    }
  ]
}
```

Uses the existing pg_trgm GIN indexes on `title` and `venue_market_id`.

### `GET /api/v1/matching/embed-script`

Returns the embed_worker.py file with content-type `text/x-python` and content-disposition attachment.

## Pipeline Page (`/matching/pipeline`)

A horizontal funnel visualization with 5 stage cards:

```
Unembedded → Embedded → Candidates → LLM Reviewed → Pairs
```

Each card shows:
- Stage name
- Count (from `/api/v1/matching/stats`)
- Status icon (⚠️ if count > threshold, ✅ if done, ⏳ if in progress)

Cards are clickable — clicking shows a detail panel below with relevant entries:
- Unembedded: list of venues + counts, or a table of recent unembedded markets
- Candidates: table of pending candidates with similarity scores
- Pairs: same table as the Review page, filtered by status

State: fetches stats on load, auto-refreshes every 30 seconds.

## Review Page (`/matching/review`)

Enhanced from current:
- **Search bar** at top: dropdown for venue filter + text input for market title
- Below search: two modes
  - **Default mode**: same table as current (pending pairs awaiting approval)
  - **Search result mode**: shows the matched market + all its cross-venue candidates/pairs in a detail view
- Each pair row: Market A (venue badge, title) | Market B (venue badge, title) | Similarity score | LLM Relationship badge | Confidence | Reasoning | Actions (Approve/Reject)
- For pairs not yet reviewed by LLM: shows "Waiting for LLM review" with the candidate similarity score
- For approved/rejected pairs: shows the result with reasoning, no action buttons

## Settings Page (`/matching/settings`)

Adds to the existing AI Pair Review section:
- **Similarity Threshold** slider (0.50 – 1.00, step 0.01, default 0.80)
  - Tooltip: "Markets with similarity below this threshold will not be sent for LLM comparison"
- **Download Embed Worker** button
  - Fetches `GET /api/v1/matching/embed-script` and saves as `embed_worker.py`
  - Note: "Run on your local machine to drain the embedding backlog"

## Embed Worker Script

Adapted from the old PolyBot script. Key differences:
- Uses REST API (`https://arby.nostrabotus.com/api/v1/markets/unembedded` and `POST /api/v1/markets/embeddings`) instead of SSH tunnel + direct DB
- Accepts `--host` flag for the API base URL
- Accepts `--batch-size` and `--fetch-limit` flags
- Uses `BAAI/bge-large-en-v1.5` (1024-dim, matches schema)
- No SSH tunnel, no DB credentials needed
- Simple loop: fetch unembedded → embed locally → POST embeddings back

## Similarity Threshold Semantics

The threshold serves two purposes:

1. **Candidate discovery** — `CandidateDiscoverer` only inserts candidates where `similarity >= threshold` (the existing ANN filter). This controls what gets found in the first place.
2. **LLM review qualification** — `Reviewer.GetPendingCandidatesWithMarkets` also filters by `similarity >= threshold`. A candidate must meet the threshold to be sent to the LLM. This prevents low-confidence pairs from consuming API credits.

Both `CandidateDiscoverer.RunOnce` and `Reviewer.RunOnce` read the threshold from `matching_config` on each cycle, so changes via the Settings page take effect immediately without restart.

## Data Flow for Similarity Threshold

Current: Hardcoded at 0.80 in constructor.

New: Store adds `GetSimilarityThreshold() float64` and `SetSimilarityThreshold(val float64)` methods reading/writing `matching_config` table:

```sql
CREATE TABLE IF NOT EXISTS matching_config (
    key   VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL
);
INSERT INTO matching_config (key, value) VALUES ('similarity_threshold', '0.80')
ON CONFLICT (key) DO NOTHING;
```

Both discovery and reviewer filter candidates by this threshold.

## Implementation Order

1. Backend: `matching_config` table, settings endpoints, stats endpoint, market search endpoint, pairs-by-market filter
2. Backend: Update `CandidateDiscoverer` and `Reviewer` to read thresholds from config
3. Backend: Embed script endpoint
4. UI: Sidebar restructuring (Matching dropdown)
5. UI: Pipeline page
6. UI: Enhanced Review page with search
7. UI: Settings page additions (threshold slider + download button)
8. Deploy

## Testing

- Backend: Unit tests for new store methods (mocking DB)
- Frontend: Manual testing via deployed instance
- Embed script: Test locally against dev instance
