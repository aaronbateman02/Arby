"use client"

import { useState } from "react"
import { strategyDetails } from "@/lib/data"
import type { StrategyDetail, StrategyConfig } from "@/lib/data"
import { SettingsModal } from "@/components/SettingsModal"

const typeColors: Record<string, string> = {
  Spread: "bg-accent/10 text-accent",
  Event: "bg-green/10 text-green",
  Macro: "bg-amber/10 text-amber",
  Commodities: "bg-red/10 text-red",
  Sports: "bg-sky-500/10 text-sky-500",
}

const statusBadge: Record<string, string> = {
  active: "text-green bg-green/10",
  paused: "text-amber bg-amber/10",
  disabled: "text-muted bg-surface-hover",
}

export default function StrategiesPage() {
  const [expanded, setExpanded] = useState<string | null>(null)
  const [editing, setEditing] = useState<StrategyDetail | null>(null)

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Strategies</h1>
          <p className="text-sm text-muted mt-1">Configure and monitor every trading strategy. Expand to inspect monitored pairs.</p>
        </div>
      </div>

      {strategyDetails.map((s) => (
        <div key={s.id} className="bg-surface-alt rounded-xl border border-border overflow-hidden">
          <div className="p-5">
            <div className="flex items-start gap-4">
              <button className="flex-1 text-left" onClick={() => setExpanded(expanded === s.id ? null : s.id)}>
                <div className="flex items-start justify-between gap-3">
                  <div className="flex items-center gap-2">
                    <h2 className="text-base font-semibold text-gray-100">{s.name}</h2>
                    <span className={`text-xs px-1.5 py-0.5 rounded ${typeColors[s.type] ?? "text-muted bg-surface-hover"}`}>{s.type}</span>
                    <span className={`text-xs px-1.5 py-0.5 rounded ${statusBadge[s.status]}`}>{s.status}</span>
                  </div>
                  <span className="text-xs text-muted shrink-0">{expanded === s.id ? "▲ Hide pairs" : "▼ View pairs"}</span>
                </div>
                <p className="text-xs text-muted mt-1.5">{s.description}</p>
                {s.pausedReason && <p className="text-xs text-amber mt-2 italic">{s.pausedReason}</p>}
                <div className="grid grid-cols-5 gap-3 text-center mt-4">
                  <StatBox label="Total" value={String(s.stats.totalBundles)} />
                  <StatBox label="Completed" value={String(s.stats.completedBundles)} positive />
                  <StatBox label="Aborted" value={String(s.stats.abortedBundles)} negative />
                  <StatBox label="P&L" value={`+$${s.stats.totalPnL.toLocaleString()}`} positive />
                  <StatBox label="Win Rate" value={`${s.stats.winRate}%`} positive={s.stats.winRate >= 70} />
                </div>
              </button>
              <button
                onClick={() => setEditing(s)}
                className="shrink-0 text-xs px-2.5 py-1.5 rounded-lg border border-border text-muted hover:text-gray-200 hover:border-muted transition-colors"
              >
                Edit
              </button>
            </div>
          </div>

          {expanded === s.id && s.monitoredPairs.length > 0 && (
            <div className="border-t border-border bg-surface/50 px-5 py-4">
              <PairsPanel pairs={s.monitoredPairs} minRoi={s.config.minRoi} />
            </div>
          )}

          {expanded === s.id && s.monitoredPairs.length === 0 && (
            <div className="border-t border-border px-5 py-6 text-center text-sm text-muted">
              No pairs currently monitored by this strategy.
            </div>
          )}
        </div>
      ))}

      {editing && (
        <SettingsModal
          strategy={editing}
          onClose={() => setEditing(null)}
          onSave={(updated) => {
            Object.assign(editing, updated)
            setEditing(null)
          }}
        />
      )}
    </div>
  )
}

function StatBox({ label, value, positive, negative }: { label: string; value: string; positive?: boolean; negative?: boolean }) {
  return (
    <div className="bg-surface rounded-lg border border-border/50 px-3 py-2">
      <p className="text-xs text-muted">{label}</p>
      <p className={`text-sm font-bold mt-0.5 ${positive ? "text-green" : negative ? "text-red" : "text-gray-200"}`}>{value}</p>
    </div>
  )
}

function PairsPanel({ pairs, minRoi }: { pairs: import("@/lib/data").MonitoredPair[]; minRoi: number }) {
  const [filter, setFilter] = useState<"all" | "criteria" | "executed">("criteria")
  const [search, setSearch] = useState("")

  const filtered = pairs
    .filter((p) => {
      if (filter === "criteria") return p.meetsCriteria && !p.executed
      if (filter === "executed") return p.executed
      return true
    })
    .filter((p) => {
      if (!search) return true
      const q = search.toLowerCase()
      return p.marketA.title.toLowerCase().includes(q) || p.marketB.title.toLowerCase().includes(q)
    })

  return (
    <div>
      <div className="flex flex-wrap items-center gap-3 mb-4">
        <span className="text-sm font-semibold text-gray-200">Monitored Pairs</span>
        {(["criteria", "executed", "all"] as const).map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`text-xs px-2 py-1 rounded border transition-colors ${
              filter === f
                ? "bg-accent text-white border-accent"
                : "border-border text-muted hover:text-gray-200 hover:border-muted"
            }`}
          >
            {f === "criteria" ? "Meets Criteria" : f === "executed" ? "Executed" : "All"}
          </button>
        ))}
        <input
          className="md:ml-auto w-full md:w-56 bg-surface border border-border rounded px-3 py-1.5 text-sm text-gray-200 placeholder-muted outline-none"
          placeholder="Search pairs..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <span className="text-xs text-muted">{filtered.length} shown</span>
      </div>

      {filtered.length === 0 ? (
        <p className="text-sm text-muted text-center py-8">No pairs match this filter.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-muted text-xs">
                <th className="text-left py-2 pr-3 font-medium">Market A</th>
                <th className="text-left py-2 px-3 font-medium">Market B</th>
                <th className="text-center py-2 px-3 font-medium">ROI</th>
                <th className="text-center py-2 px-3 font-medium">Spread</th>
                <th className="text-center py-2 px-3 font-medium">Cost/Sh</th>
                <th className="text-center py-2 px-3 font-medium">Expires</th>
                <th className="text-center py-2 pl-3 font-medium">Est. Profit</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((p) => (
                <tr key={p.id} className={`border-b border-border/50 ${p.meetsCriteria && !p.executed ? "bg-accent/5" : ""}`}>
                  <td className="py-2.5 pr-3">
                    <LegCell venue={p.marketA.venue} dir={p.marketA.dir} title={p.marketA.title} price={p.marketA.price} />
                  </td>
                  <td className="py-2.5 px-3">
                    <LegCell venue={p.marketB.venue} dir={p.marketB.dir} title={p.marketB.title} price={p.marketB.price} />
                  </td>
                  <td className={`py-2.5 px-3 text-center font-medium ${p.currentRoi >= minRoi ? "text-green" : "text-amber"}`}>{p.currentRoi}%</td>
                  <td className="py-2.5 px-3 text-center text-muted">{p.currentSpread}¢</td>
                  <td className="py-2.5 px-3 text-center text-muted font-mono">${p.costPerShare.toFixed(2)}</td>
                  <td className="py-2.5 px-3 text-center text-muted">{p.expiresIn}</td>
                  <td className="py-2.5 px-3 text-center">
                    <div className="flex items-center justify-center gap-1.5">
                      <span className="text-green font-medium">${p.estimatedProfit.toFixed(2)}</span>
                      {p.executed && <span className="text-xs px-1 py-0.5 rounded text-green bg-green/10">executed</span>}
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

function LegCell({ venue, dir, title, price }: { venue: string; dir: string; title: string; price: number }) {
  return (
    <div className="flex items-start gap-1.5">
      <span className={`mt-0.5 inline-flex h-4 w-4 items-center justify-center rounded text-[9px] font-bold text-white ${venue === "Kalshi" ? "bg-accent" : "bg-amber"}`}>
        {venue === "Kalshi" ? "K" : "P"}
      </span>
      <span className={`inline-flex rounded px-1 py-0.5 text-[10px] font-bold ${dir === "BUY YES" ? "text-green bg-green/10" : "text-red bg-red/10"}`}>
        {dir}
      </span>
      <span className="text-gray-200 truncate max-w-[200px]" title={title}>{title}</span>
      <span className="text-muted font-mono text-xs ml-auto">${price.toFixed(2)}</span>
    </div>
  )
}
