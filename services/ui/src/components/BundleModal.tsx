import type { Bundle } from "@/lib/data"

export function BundleModal({ bundle, onClose }: { bundle: Bundle; onClose: () => void }) {
  const daysToResolve = Math.ceil(
    (new Date(bundle.resolvesAt).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24)
  )

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <div className="bg-surface rounded-xl border border-border shadow-xl max-w-2xl w-full mx-4 max-h-[85vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between p-5 border-b border-border">
          <div>
            <h2 className="text-lg font-bold text-gray-100">{bundle.name}</h2>
            <p className="text-xs text-muted mt-0.5">ID: {bundle.id}</p>
          </div>
          <button onClick={onClose} className="text-muted hover:text-gray-200 p-1">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="p-5 space-y-5">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <Detail label="Strategy" value={bundle.strategy} />
            <Detail label="Opened" value={bundle.openedAt} />
            <Detail label="Resolves" value={bundle.resolvesAt} />
            <Detail label="Time to Resolve" value={`${daysToResolve} days`} />
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <Detail label="Exposure" value={`$${bundle.exposure.toLocaleString()}`} />
            <Detail label="Projected ROI" value={`${bundle.projectedRoi}%`} positive />
            <Detail label="Actual ROI" value={`${bundle.actualRoi}%`} positive />
            <Detail label="Total Fees" value={`$${bundle.totalFees.toFixed(2)}`} negative />
          </div>

          <div>
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-semibold text-gray-200">Legs</h3>
              <span className="text-xs text-muted">{bundle.legs.length} legs</span>
            </div>
            <div className="space-y-2">
              <div className="grid grid-cols-5 gap-2 text-xs text-muted px-3">
                <span className="col-span-2">Market</span>
                <span className="text-right">Venue</span>
                <span className="text-right">Est. Cost</span>
                <span className="text-right">Act. Cost</span>
              </div>
              {bundle.legs.map((leg, i) => (
                <div key={i} className="grid grid-cols-5 gap-2 items-center py-2 px-3 rounded-lg bg-surface-alt/50 text-sm">
                  <div className="col-span-2 flex items-center gap-2">
                    <span className={`w-1.5 h-1.5 rounded-full ${leg.venue === "Kalshi" ? "bg-accent" : "bg-amber"}`} />
                    <span className="text-gray-200 truncate">{leg.market}</span>
                  </div>
                  <span className="text-right text-muted">{leg.venue}</span>
                  <span className="text-right text-muted">${leg.estimatedCost.toFixed(2)}</span>
                  <div className="flex items-center justify-end gap-2">
                    <span className="text-gray-200">${leg.actualCost.toFixed(2)}</span>
                    <span className={`text-xs px-1 py-0.5 rounded ${
                      leg.status === "filled" ? "text-green bg-green/10" :
                      leg.status === "working" ? "text-amber bg-amber/10" : "text-muted bg-surface-hover"
                    }`}>{leg.status}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 pt-3 border-t border-border">
            <Detail label="Side" value={bundle.legs[0]?.side ?? "—"} />
            <Detail label="P&L" value={`+$${bundle.pnl.toFixed(2)}`} positive />
            <Detail label="Status" value={bundle.status} />
            <Detail label="Total Fees" value={`$${bundle.totalFees.toFixed(2)}`} negative />
          </div>
        </div>
      </div>
    </div>
  )
}

function Detail({ label, value, positive }: { label: string; value: string; positive?: boolean; negative?: boolean }) {
  return (
    <div>
      <p className="text-xs text-muted">{label}</p>
      <p className={`text-sm font-medium mt-0.5 ${positive ? "text-green" : "text-gray-200"}`}>{value}</p>
    </div>
  )
}
