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

interface CategoryCount {
  venue: string
  category: string
  count: number
}

interface PipelineCounts {
  events: CategoryCount[]
  markets: CategoryCount[]
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
  const [counts, setCounts] = useState<PipelineCounts | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  useEffect(() => {
    const fetchAll = async () => {
      try {
        setError("")
        const [statsRes, countsRes] = await Promise.all([
          fetch("/api/matching/stats"),
          fetch("/api/matching/pipeline-counts"),
        ])
        if (!statsRes.ok) throw new Error(`Stats HTTP ${statsRes.status}`)
        if (!countsRes.ok) throw new Error(`Counts HTTP ${countsRes.status}`)
        setStats(await statsRes.json())
        setCounts(await countsRes.json())
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load data")
      } finally {
        setLoading(false)
      }
    }
    fetchAll()
    const interval = setInterval(fetchAll, 30_000)
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

      {counts && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <CountTable title="Events" data={counts.events} />
          <CountTable title="Markets" data={counts.markets} />
        </div>
      )}
    </div>
  )
}

function CountTable({ title, data }: { title: string; data: CategoryCount[] }) {
  const venues = Array.from(new Set(data.map((d) => d.venue)))
  const categories = Array.from(new Set(data.map((d) => d.category)))
  const getCount = (venue: string, cat: string) => data.find((d) => d.venue === venue && d.category === cat)?.count ?? 0
  const venueTotals = venues.map((v) => data.filter((d) => d.venue === v).reduce((sum, d) => sum + d.count, 0))

  return (
    <div className="bg-surface-alt rounded-xl border border-border overflow-hidden">
      <div className="px-5 py-3 border-b border-border">
        <h2 className="text-sm font-semibold text-gray-200">{title}</h2>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border">
              <th className="text-left px-5 py-2 text-muted font-medium">Category</th>
              {venues.map((v) => (
                <th key={v} className="text-right px-4 py-2 text-muted font-medium">{v}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {categories.map((cat) => (
              <tr key={cat} className="border-b border-border/50 last:border-0">
                <td className="px-5 py-1.5 text-gray-300">{cat}</td>
                {venues.map((v) => (
                  <td key={v} className="text-right px-4 py-1.5 text-gray-200 tabular-nums">
                    {getCount(v, cat).toLocaleString()}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr className="border-t border-border bg-white/5">
              <td className="px-5 py-2 font-medium text-gray-200">Total</td>
              {venueTotals.map((t, i) => (
                <td key={i} className="text-right px-4 py-2 font-semibold text-gray-100 tabular-nums">
                  {t.toLocaleString()}
                </td>
              ))}
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  )
}
