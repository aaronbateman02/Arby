export default function BundlesPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Bundles</h1>
          <p className="text-sm text-muted mt-1">Multi-leg arbitrage positions</p>
        </div>
        <div className="flex gap-2">
          <select className="bg-surface-alt border border-border rounded-lg px-3 py-2 text-sm text-gray-200 outline-none">
            <option>All bundles</option>
            <option>Active</option>
            <option>Completed</option>
            <option>Cancelled</option>
          </select>
        </div>
      </div>

      <div className="space-y-4">
        {bundles.map((b, i) => (
          <div key={i} className="bg-surface-alt rounded-xl border border-border p-5">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <h2 className="text-sm font-semibold text-gray-200">{b.name}</h2>
                <span className={`text-xs px-1.5 py-0.5 rounded ${
                  b.status === "active" ? "text-green bg-green/10" :
                  b.status === "completed" ? "text-muted bg-surface-hover" : "text-red bg-red/10"
                }`}>{b.status}</span>
              </div>
              <div className="flex items-center gap-4 text-sm">
                <span className="text-muted">Exposure: <span className="text-gray-200">{b.exposure}</span></span>
                <span className="text-green">P&L: +{b.pnl}</span>
              </div>
            </div>
            <div className="space-y-2">
              {b.legs.map((leg, j) => (
                <div key={j} className="flex items-center justify-between py-2 px-3 rounded-lg bg-surface/50 text-sm">
                  <div className="flex items-center gap-2">
                    <span className={`w-1.5 h-1.5 rounded-full ${leg.venue === "Kalshi" ? "bg-accent" : "bg-amber"}`} />
                    <span className="text-gray-200">{leg.name}</span>
                  </div>
                  <div className="flex items-center gap-4 text-muted">
                    <span>{leg.venue}</span>
                    <span>Side: {leg.side}</span>
                    <span>Size: {leg.size}</span>
                    <span className={leg.status === "filled" ? "text-green" : "text-amber"}>{leg.status}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

const bundles = [
  {
    name: "BTC June Cross", status: "active", exposure: "$1,200", pnl: "$48.00",
    legs: [
      { name: "BTC > $120K in June?", venue: "Kalshi", side: "Buy YES", size: "$600", status: "filled" },
      { name: "BTC > $120K in June?", venue: "Polymarket", side: "Buy NO", size: "$600", status: "filled" },
    ],
  },
  {
    name: "DOW July Spread", status: "active", exposure: "$850", pnl: "$34.00",
    legs: [
      { name: "Will DOW hit 45K by July?", venue: "Kalshi", side: "Buy YES", size: "$425", status: "filled" },
      { name: "Will DOW hit 45K by July?", venue: "Polymarket", side: "Buy NO", size: "$425", status: "filled" },
    ],
  },
  {
    name: "ETH ETF Basket", status: "active", exposure: "$2,100", pnl: "$105.00",
    legs: [
      { name: "ETH ETF approved by Aug?", venue: "Kalshi", side: "Buy YES", size: "$600", status: "filled" },
      { name: "ETH ETF approved by Aug?", venue: "Polymarket", side: "Buy NO", size: "$500", status: "working" },
      { name: "SEC approves ETH ETF 2026?", venue: "Kalshi", side: "Buy YES", size: "$500", status: "filled" },
      { name: "SEC approves ETH ETF 2026?", venue: "Polymarket", side: "Buy NO", size: "$500", status: "filled" },
    ],
  },
]
