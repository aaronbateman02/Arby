import { useState } from "react"
import { bundles } from "@/lib/data"
import type { Bundle } from "@/lib/data"
import { BundleModal } from "./BundleModal"

export function BundleTable() {
  const [search, setSearch] = useState("")
  const [selected, setSelected] = useState<Bundle | null>(null)

  const filtered = bundles.filter((b) =>
    b.name.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <>
      <div className="bg-surface-alt rounded-xl border border-border overflow-hidden">
        <div className="p-3 border-b border-border flex items-center gap-3">
          <svg className="w-4 h-4 text-muted shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            className="bg-transparent text-sm text-gray-200 placeholder-muted outline-none flex-1"
            placeholder="Search bundles..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <span className="text-xs text-muted">{filtered.length} bundles</span>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-muted text-xs">
                <th className="text-left py-2.5 px-4 font-medium">Name</th>
                <th className="text-left py-2.5 px-4 font-medium">Strategy</th>
                <th className="text-right py-2.5 px-4 font-medium">Exposure</th>
                <th className="text-right py-2.5 px-4 font-medium">Proj. ROI</th>
                <th className="text-right py-2.5 px-4 font-medium">Act. ROI</th>
                <th className="text-right py-2.5 px-4 font-medium">P&L</th>
                <th className="text-right py-2.5 px-4 font-medium">Resolves</th>
                <th className="text-right py-2.5 px-4 font-medium">Status</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((b) => (
                <tr
                  key={b.id}
                  onClick={() => setSelected(b)}
                  className="border-b border-border/50 hover:bg-surface-hover transition-colors cursor-pointer"
                >
                  <td className="py-2.5 px-4 text-gray-200 font-medium">{b.name}</td>
                  <td className="py-2.5 px-4 text-muted">{b.strategy}</td>
                  <td className="py-2.5 px-4 text-right text-gray-200">${b.exposure.toLocaleString()}</td>
                  <td className="py-2.5 px-4 text-right text-green">{b.projectedRoi}%</td>
                  <td className={`py-2.5 px-4 text-right ${b.actualRoi >= b.projectedRoi ? "text-green" : "text-amber"}`}>{b.actualRoi}%</td>
                  <td className="py-2.5 px-4 text-right text-green">+${b.pnl.toFixed(2)}</td>
                  <td className="py-2.5 px-4 text-right text-muted">{b.resolvesAt.split(" ")[0]}</td>
                  <td className="py-2.5 px-4 text-right">
                    <span className={`text-xs px-1.5 py-0.5 rounded ${
                      b.status === "active" ? "text-green bg-green/10" :
                      b.status === "completed" ? "text-muted bg-surface-hover" : "text-red bg-red/10"
                    }`}>{b.status}</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
      {selected && <BundleModal bundle={selected} onClose={() => setSelected(null)} />}
    </>
  )
}
