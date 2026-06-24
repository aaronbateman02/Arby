export default function MarketsPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Markets</h1>
          <p className="text-sm text-muted mt-1">Browse cross-venue pricing</p>
        </div>
        <div className="flex gap-2">
          <input
            className="bg-surface-alt border border-border rounded-lg px-3 py-2 text-sm text-gray-200 placeholder-muted w-64 outline-none focus:border-accent"
            placeholder="Search markets..."
            readOnly
          />
          <select className="bg-surface-alt border border-border rounded-lg px-3 py-2 text-sm text-gray-200 outline-none">
            <option>All venues</option>
            <option>Kalshi</option>
            <option>Polymarket</option>
          </select>
        </div>
      </div>

      <div className="bg-surface-alt rounded-xl border border-border overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-muted">
              <th className="text-left py-3 px-4 font-medium">Market</th>
              <th className="text-right py-3 px-4 font-medium">Kalshi</th>
              <th className="text-right py-3 px-4 font-medium">Polymarket</th>
              <th className="text-right py-3 px-4 font-medium">Spread</th>
              <th className="text-right py-3 px-4 font-medium">Volume (24h)</th>
              <th className="text-right py-3 px-4 font-medium">Liquidity</th>
              <th className="text-right py-3 px-4 font-medium">Status</th>
            </tr>
          </thead>
          <tbody>
            {markets.map((m, i) => (
              <tr key={i} className="border-b border-border/50 hover:bg-surface-hover transition-colors">
                <td className="py-3 px-4 text-gray-200 font-medium">{m.name}</td>
                <td className={`py-3 px-4 text-right ${m.kalshi ? "text-gray-200" : "text-muted"}`}>{m.kalshi ?? "—"}</td>
                <td className={`py-3 px-4 text-right ${m.poly ? "text-gray-200" : "text-muted"}`}>{m.poly ?? "—"}</td>
                <td className="py-3 px-4 text-right text-green font-medium">{m.spread}</td>
                <td className="py-3 px-4 text-right text-muted">{m.volume}</td>
                <td className="py-3 px-4 text-right">
                  <span className={`text-xs px-1.5 py-0.5 rounded ${m.liquidity === "high" ? "text-green bg-green/10" : m.liquidity === "medium" ? "text-amber bg-amber/10" : "text-red bg-red/10"}`}>
                    {m.liquidity}
                  </span>
                </td>
                <td className="py-3 px-4 text-right">
                  <span className={`text-xs px-1.5 py-0.5 rounded ${m.status === "active" ? "text-green bg-green/10" : "text-muted bg-surface-hover"}`}>
                    {m.status}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

const markets = [
  { name: "Will DOW hit 45K by July?", kalshi: "$0.42", poly: "$0.38", spread: "4.0¢", volume: "$12.4K", liquidity: "high", status: "active" },
  { name: "BTC > $120K in June?", kalshi: "$0.31", poly: "$0.27", spread: "4.0¢", volume: "$8.9K", liquidity: "high", status: "active" },
  { name: "Fed cuts rates in Q3?", kalshi: "$0.65", poly: "$0.61", spread: "4.0¢", volume: "$15.2K", liquidity: "medium", status: "active" },
  { name: "S&P 500 > 5600 EOM?", kalshi: "$0.48", poly: "$0.44", spread: "4.0¢", volume: "$6.1K", liquidity: "medium", status: "active" },
  { name: "ETH ETF approved by Aug?", kalshi: "$0.55", poly: "$0.50", spread: "5.0¢", volume: "$22.3K", liquidity: "high", status: "active" },
  { name: "Apple > $250 by Sept?", kalshi: "$0.37", poly: "$0.33", spread: "4.0¢", volume: "$4.2K", liquidity: "low", status: "active" },
  { name: "US inflation < 3% in 2026?", kalshi: "$0.72", poly: null, spread: "—", volume: "$1.1K", liquidity: "low", status: "inactive" },
]
