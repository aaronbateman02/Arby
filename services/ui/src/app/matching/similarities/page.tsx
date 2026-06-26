"use client"

import { useEffect, useState } from "react"

interface Market {
  id: string
  venue: string
  venue_market_id: string
  title: string
  description: string
  category: string
}

interface SimilarityPair {
  market_a: Market
  market_b: Market
  similarity: number
}

export default function SimilaritiesPage() {
  const [pairs, setPairs] = useState<SimilarityPair[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  useEffect(() => {
    const fetchData = async () => {
      try {
        setError("")
        const res = await fetch("/api/matching/top-similarities")
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        setPairs(await res.json())
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load")
      } finally {
        setLoading(false)
      }
    }
    fetchData()
    const interval = setInterval(fetchData, 30_000)
    return () => clearInterval(interval)
  }, [])

  if (loading) {
    return <div className="p-6 max-w-7xl mx-auto"><h1 className="text-2xl font-bold text-gray-100">Top Similarities</h1><p className="text-sm text-muted mt-1">Loading…</p></div>
  }
  if (error) {
    return <div className="p-6 max-w-7xl mx-auto"><h1 className="text-2xl font-bold text-gray-100">Top Similarities</h1><p className="text-sm text-red mt-1">{error}</p></div>
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-100">Top 100 Cross-Venue Similarities</h1>
        <p className="text-sm text-muted mt-1">Highest embedding cosine similarity between Kalshi and Polymarket</p>
      </div>

      {pairs.length === 0 ? (
        <div className="bg-surface-alt rounded-xl border border-border p-8 text-center">
          <p className="text-muted">No pairs found. Need embedded markets on both venues.</p>
        </div>
      ) : (
        <div className="bg-surface-alt rounded-xl border border-border overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left px-4 py-2 text-muted font-medium w-12">#</th>
                  <th className="text-right px-3 py-2 text-muted font-medium w-16">Sim</th>
                  <th className="text-left px-4 py-2 text-muted font-medium">Kalshi</th>
                  <th className="text-left px-4 py-2 text-muted font-medium">Polymarket</th>
                </tr>
              </thead>
              <tbody>
                {pairs.map((p, i) => (
                  <tr key={i} className="border-b border-border/50 last:border-0 hover:bg-white/5">
                    <td className="px-4 py-2 text-muted">{i + 1}</td>
                    <td className="px-3 py-2 text-right">
                      <span className={p.similarity >= 0.80 ? "text-green font-semibold" : p.similarity >= 0.70 ? "text-yellow" : "text-muted"}>
                        {(p.similarity * 100).toFixed(1)}%
                      </span>
                    </td>
                    <td className="px-4 py-2">
                      <div className="text-gray-200 truncate max-w-md" title={p.market_a.title}>{p.market_a.title}</div>
                      <div className="text-muted text-xs truncate max-w-md">{p.market_a.description?.slice(0, 120)}</div>
                    </td>
                    <td className="px-4 py-2">
                      <div className="text-gray-200 truncate max-w-md" title={p.market_b.title}>{p.market_b.title}</div>
                      <div className="text-muted text-xs truncate max-w-md">{p.market_b.description?.slice(0, 120)}</div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
