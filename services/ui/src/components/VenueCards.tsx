import { venues } from "@/lib/data"

export function VenueCards() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {venues.map((v) => (
        <div key={v.name} className="bg-surface-alt rounded-xl border border-border p-4">
          <div className="flex items-center gap-2 mb-3">
            <span className={`w-2 h-2 rounded-full ${v.name === "Kalshi" ? "bg-accent" : "bg-amber"}`} />
            <span className="text-sm font-semibold text-gray-200">{v.name}</span>
          </div>
          <div className="grid grid-cols-3 gap-3">
            <div>
              <p className="text-xs text-muted">Cash</p>
              <p className="text-lg font-bold text-gray-100">${v.cash.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-xs text-muted">Positions</p>
              <p className="text-lg font-bold text-gray-100">${v.positions.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-xs text-muted">Portfolio</p>
              <p className="text-lg font-bold text-gray-100">${v.portfolio.toLocaleString()}</p>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}
