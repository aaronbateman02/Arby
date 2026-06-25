"use client"

import { useState, useEffect, useCallback } from "react"

type MatchPair = {
  id: string
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

type SearchMarket = {
  id: string
  title: string
  description: string
  venue: string
}

const confidenceColor = (c: number) => {
  if (c >= 0.90) return "text-green"
  if (c >= 0.70) return "text-amber"
  return "text-muted"
}

const relationshipBadge = (r: string) => {
  if (r === "EQUIVALENT") return "bg-green/20 text-green"
  if (r === "INVERSE") return "bg-amber/20 text-amber"
  return "bg-gray-500/20 text-gray-400"
}

const venueBadge = (v: string) => {
  const vl = v.toLowerCase()
  if (vl === "kalshi" || vl === "kalshi") return "bg-accent/20 text-accent"
  if (vl === "polymarket" || vl === "polymarket") return "bg-amber/20 text-amber"
  return "bg-gray-500/20 text-gray-400"
}

const statusBadge = (s: string) => {
  if (s === "PENDING_APPROVAL") return "bg-yellow-500/20 text-yellow-400"
  if (s === "APPROVED") return "bg-green/20 text-green"
  if (s === "REJECTED") return "bg-red-500/20 text-red-400"
  return "bg-gray-500/20 text-gray-400"
}

const VENUES = ["Kalshi", "Polymarket"]

export default function MatchingReviewPage() {
  const [venue, setVenue] = useState("Kalshi")
  const [query, setQuery] = useState("")
  const [searchResults, setSearchResults] = useState<SearchMarket[]>([])
  const [searchLoading, setSearchLoading] = useState(false)
  const [searchError, setSearchError] = useState("")
  const [searched, setSearched] = useState(false)

  const [selectedMarketId, setSelectedMarketId] = useState<string | null>(null)
  const [selectedMarketTitle, setSelectedMarketTitle] = useState("")
  const [marketPairs, setMarketPairs] = useState<MatchPair[]>([])
  const [marketPairsLoading, setMarketPairsLoading] = useState(false)
  const [marketPairsError, setMarketPairsError] = useState("")

  const [pendingPairs, setPendingPairs] = useState<MatchPair[]>([])
  const [pendingLoading, setPendingLoading] = useState(true)
  const [pendingError, setPendingError] = useState("")
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  useEffect(() => {
    fetch("/api/matching/pairs?status=PENDING_APPROVAL")
      .then(r => r.json())
      .then(data => setPendingPairs(data.pairs ?? []))
      .catch(() => setPendingError("Failed to load pending pairs"))
      .finally(() => setPendingLoading(false))
  }, [])

  const handleSearch = useCallback(async () => {
    if (!query.trim()) return
    setSearchLoading(true)
    setSearchError("")
    setSearched(true)
    setSelectedMarketId(null)
    setMarketPairs([])
    try {
      const res = await fetch(`/api/matching/markets/search?venue=${encodeURIComponent(venue)}&q=${encodeURIComponent(query)}`)
      const data = await res.json()
      setSearchResults(data.markets ?? [])
    } catch {
      setSearchError("Search failed")
    } finally {
      setSearchLoading(false)
    }
  }, [venue, query])

  const handleSelectMarket = useCallback(async (market: SearchMarket) => {
    setSelectedMarketId(market.id)
    setSelectedMarketTitle(market.title)
    setMarketPairsLoading(true)
    setMarketPairsError("")
    try {
      const res = await fetch(`/api/matching/pairs?market_id=${market.id}`)
      const data = await res.json()
      setMarketPairs(data.pairs ?? [])
    } catch {
      setMarketPairsError("Failed to load market pairs")
    } finally {
      setMarketPairsLoading(false)
    }
  }, [])

  const handleApprove = async (id: string) => {
    setActionLoading(id)
    try {
      await fetch("/api/matching/pairs/approve", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id }),
      })
      setPendingPairs(prev => prev.filter(p => p.id !== id))
      setMarketPairs(prev => prev.map(p => p.id === id ? { ...p, status: "APPROVED" } : p))
    } catch {
      // silently fail
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
      setPendingPairs(prev => prev.filter(p => p.id !== id))
      setMarketPairs(prev => prev.map(p => p.id === id ? { ...p, status: "REJECTED" } : p))
    } catch {
      // silently fail
    } finally {
      setActionLoading(null)
    }
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-100">Match Review</h1>
        <p className="text-sm text-muted mt-1">Search markets and review candidate match pairs</p>
      </div>

      {/* Search */}
      <div className="bg-surface-alt rounded-xl border border-border p-4">
        <div className="flex items-center gap-3">
          <select
            value={venue}
            onChange={e => setVenue(e.target.value)}
            className="bg-surface border border-border rounded-lg px-3 py-2 text-sm text-gray-200 outline-none focus:border-accent"
          >
            {VENUES.map(v => (
              <option key={v} value={v}>{v}</option>
            ))}
          </select>
          <input
            type="text"
            value={query}
            onChange={e => setQuery(e.target.value)}
            onKeyDown={e => e.key === "Enter" && handleSearch()}
            placeholder="Search markets..."
            className="flex-1 bg-surface border border-border rounded-lg px-3 py-2 text-sm text-gray-200 placeholder:text-muted outline-none focus:border-accent"
          />
          <button
            onClick={handleSearch}
            disabled={searchLoading || !query.trim()}
            className="px-4 py-2 text-sm rounded-lg bg-accent/20 text-accent hover:bg-accent/30 disabled:opacity-40 transition-colors"
          >
            {searchLoading ? "..." : "Search"}
          </button>
        </div>
      </div>

      {/* Search results */}
      {searched && (
        <div className="bg-surface-alt rounded-xl border border-border overflow-hidden">
          <div className="px-4 py-3 border-b border-border text-xs text-muted uppercase tracking-wider font-medium">
            Search Results
          </div>
          {searchLoading && (
            <div className="p-6 text-sm text-muted animate-pulse">Searching...</div>
          )}
          {searchError && (
            <div className="p-6 text-sm text-red-400">{searchError}</div>
          )}
          {!searchLoading && !searchError && searchResults.length === 0 && (
            <div className="p-6 text-sm text-muted">No markets found</div>
          )}
          {!searchLoading && searchResults.length > 0 && (
            <div className="divide-y divide-border">
              {searchResults.map(m => (
                <button
                  key={m.id}
                  onClick={() => handleSelectMarket(m)}
                  className={`w-full text-left px-4 py-3 hover:bg-surface/50 transition-colors ${selectedMarketId === m.id ? "bg-accent/5" : ""}`}
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="min-w-0 flex-1">
                      <div className="text-sm text-gray-200 truncate">{m.title}</div>
                      {m.description && (
                        <div className="text-xs text-muted mt-0.5 line-clamp-2">{m.description}</div>
                      )}
                    </div>
                    <span className={`inline-block text-[10px] px-1.5 py-0.5 rounded-full mt-0.5 shrink-0 ${venueBadge(m.venue)}`}>
                      {m.venue}
                    </span>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Market pairs detail */}
      {selectedMarketId && (
        <div className="bg-surface-alt rounded-xl border border-border overflow-hidden">
          <div className="px-4 py-3 border-b border-border text-xs text-muted uppercase tracking-wider font-medium">
            Pairs for: {selectedMarketTitle}
          </div>
          {marketPairsLoading && (
            <div className="p-6 text-sm text-muted animate-pulse">Loading pairs...</div>
          )}
          {marketPairsError && (
            <div className="p-6 text-sm text-red-400">{marketPairsError}</div>
          )}
          {!marketPairsLoading && !marketPairsError && marketPairs.length === 0 && (
            <div className="p-6 text-sm text-muted">No pairs found for this market</div>
          )}
          {!marketPairsLoading && marketPairs.length > 0 && (
            <div className="overflow-x-auto">
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
                  {marketPairs.map(p => (
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
                      <td className="px-4 py-3">
                        <span className={`text-xs px-2 py-0.5 rounded-full ${statusBadge(p.status)}`}>{p.status}</span>
                      </td>
                      <td className="px-4 py-3 text-xs text-muted max-w-[200px] truncate" title={p.reasoning}>{p.reasoning}</td>
                      <td className="px-4 py-3">
                        {p.status === "PENDING_APPROVAL" ? (
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
                              className="px-3 py-1 text-xs rounded-lg bg-red-500/20 text-red-400 hover:bg-red-500/30 disabled:opacity-40 transition-colors"
                            >
                              {actionLoading === p.id ? "..." : "Reject"}
                            </button>
                          </div>
                        ) : (
                          <span className="text-xs text-muted">—</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Pending pairs */}
      <div className="bg-surface-alt rounded-xl border border-border overflow-hidden">
        <div className="px-4 py-3 border-b border-border text-xs text-muted uppercase tracking-wider font-medium">
          Pending Approval
        </div>
        {pendingLoading && (
          <div className="p-6 text-sm text-muted animate-pulse">Loading pairs...</div>
        )}
        {pendingError && (
          <div className="p-6 text-sm text-red-400">{pendingError}</div>
        )}
        {!pendingLoading && !pendingError && pendingPairs.length === 0 && (
          <div className="flex flex-col items-center justify-center py-16 text-muted">
            <svg className="w-12 h-12 mb-4 opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p className="text-sm">No pending pairs to review</p>
            <p className="text-xs mt-1">New pairs will appear here once discovered and reviewed by the AI.</p>
          </div>
        )}
        {!pendingLoading && pendingPairs.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left border-collapse">
              <thead>
                <tr className="border-b border-border text-xs text-muted uppercase tracking-wider">
                  <th className="px-4 py-3 font-medium">Market A</th>
                  <th className="px-4 py-3 font-medium">Market B</th>
                  <th className="px-4 py-3 font-medium">Category</th>
                  <th className="px-4 py-3 font-medium">Conf.</th>
                  <th className="px-4 py-3 font-medium">Rel.</th>
                  <th className="px-4 py-3 font-medium">Models</th>
                  <th className="px-4 py-3 font-medium">Reasoning</th>
                  <th className="px-4 py-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {pendingPairs.map(p => (
                  <tr key={p.id} className="border-b border-border last:border-0 hover:bg-surface/50 transition-colors">
                    <td className="px-4 py-3">
                      <div className="text-gray-200 max-w-[200px] truncate" title={p.market_a_title}>{p.market_a_title}</div>
                      <span className={`inline-block text-[10px] px-1.5 py-0.5 rounded-full mt-1 ${venueBadge(p.venue_a)}`}>{p.venue_a}</span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="text-gray-200 max-w-[200px] truncate" title={p.market_b_title}>{p.market_b_title}</div>
                      <span className={`inline-block text-[10px] px-1.5 py-0.5 rounded-full mt-1 ${venueBadge(p.venue_b)}`}>{p.venue_b}</span>
                    </td>
                    <td className="px-4 py-3 text-muted text-xs">{p.category}</td>
                    <td className={`px-4 py-3 text-xs font-semibold ${confidenceColor(p.confidence)}`}>
                      {(p.confidence * 100).toFixed(0)}%
                    </td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-0.5 rounded-full ${relationshipBadge(p.relationship)}`}>{p.relationship}</span>
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
                          className="px-3 py-1 text-xs rounded-lg bg-red-500/20 text-red-400 hover:bg-red-500/30 disabled:opacity-40 transition-colors"
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
