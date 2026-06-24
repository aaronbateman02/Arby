"use client"

import { useState, useEffect } from "react"

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

export default function ReviewPage() {
  const [pairs, setPairs] = useState<MatchPair[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  useEffect(() => {
    fetch("/api/matching/pairs?status=PENDING_APPROVAL")
      .then(r => r.json())
      .then(data => setPairs(data.pairs ?? []))
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
      setPairs(prev => prev.filter(p => p.id !== id))
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
      setPairs(prev => prev.filter(p => p.id !== id))
    } catch {
      setError("Failed to reject")
    } finally {
      setActionLoading(null)
    }
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
    if (v === "kalshi") return "bg-accent/20 text-accent"
    if (v === "polymarket") return "bg-amber/20 text-amber"
    return "bg-gray-500/20 text-gray-400"
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-100">Match Review</h1>
        <p className="text-sm text-muted mt-1">Review and approve/reject candidate match pairs</p>
      </div>

      {loading && <div className="text-sm text-muted animate-pulse">Loading pairs...</div>}
      {error && <div className="text-sm text-red-400">{error}</div>}

      {!loading && !error && pairs.length === 0 && (
        <div className="flex flex-col items-center justify-center py-20 text-muted">
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
                <th className="px-4 py-3 font-medium">Category</th>
                <th className="px-4 py-3 font-medium">Conf.</th>
                <th className="px-4 py-3 font-medium">Rel.</th>
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
  )
}
