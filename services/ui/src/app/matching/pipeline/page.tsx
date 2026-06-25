"use client"

import { useEffect, useState } from "react"

interface Stats {
  unembedded: number
  embedded: number
  pending_candidates: number
  reviewed_candidates: number
  pairs_pending_approval: number
  pairs_approved: number
  pairs_rejected: number
}

const stages = [
  { key: "unembedded" as const, label: "Unembedded", desc: "Markets awaiting embedding" },
  { key: "embedded" as const, label: "Embedded", desc: "Markets with vector embeddings" },
  { key: "pending_candidates" as const, label: "Candidates", desc: "Similar pairs found by ANN" },
  { key: "reviewed_candidates" as const, label: "LLM Reviewed", desc: "Pairs analyzed by LLM" },
  { key: "pairs_pending_approval" as const, label: "Pending Approval", desc: "Awaiting human review" },
]

export default function PipelinePage() {
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  useEffect(() => {
    const fetchStats = async () => {
      try {
        setError("")
        const res = await fetch("/api/matching/stats")
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = await res.json()
        setStats(data)
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load stats")
      } finally {
        setLoading(false)
      }
    }
    fetchStats()
    const interval = setInterval(fetchStats, 30_000)
    return () => clearInterval(interval)
  }, [])

  if (loading) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <h1 className="text-2xl font-bold text-gray-100">Pipeline</h1>
        <p className="text-sm text-muted mt-1">Loading pipeline data…</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <h1 className="text-2xl font-bold text-gray-100">Pipeline</h1>
        <p className="text-sm text-red mt-1">{error}</p>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-100">Pipeline</h1>
        <p className="text-sm text-muted mt-1">Matching pipeline stages overview</p>
      </div>

      <div className="flex flex-col lg:flex-row items-stretch lg:items-center gap-1.5">
        {stages.map((stage, i) => (
          <div key={stage.key} className="flex-1 min-w-0 flex items-stretch lg:items-center gap-1.5">
            <div className="bg-surface-alt rounded-xl border border-border p-4 flex-1">
              <p className="text-xs text-muted mb-1">{stage.label}</p>
              <p className="text-3xl font-bold text-gray-100">{stats?.[stage.key] ?? 0}</p>
              <p className="text-xs text-muted mt-1">{stage.desc}</p>
            </div>
            {i < stages.length - 1 && (
              <span className="hidden lg:flex items-center text-muted text-xl shrink-0 px-1 select-none">→</span>
            )}
          </div>
        ))}
      </div>

      <div className="flex items-center gap-6">
        <div className="bg-surface-alt rounded-xl border border-border px-5 py-4">
          <p className="text-xs text-muted mb-0.5">Approved</p>
          <p className="text-2xl font-bold text-green">{stats?.pairs_approved ?? 0}</p>
        </div>
        <div className="bg-surface-alt rounded-xl border border-border px-5 py-4">
          <p className="text-xs text-muted mb-0.5">Rejected</p>
          <p className="text-2xl font-bold text-red">{stats?.pairs_rejected ?? 0}</p>
        </div>
      </div>
    </div>
  )
}
