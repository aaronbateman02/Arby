export default function ReviewPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Match Review</h1>
          <p className="text-sm text-muted mt-1">157 pairs awaiting review</p>
        </div>
        <div className="flex gap-2">
          <button className="bg-accent hover:bg-accent-hover text-white text-sm px-4 py-2 rounded-lg transition-colors">
            Review All
          </button>
        </div>
      </div>

      <div className="space-y-3">
        {reviews.map((r, i) => (
          <div key={i} className="bg-surface-alt rounded-xl border border-border p-4 flex items-center gap-4">
            <div className="flex-1 min-w-0">
              <h3 className="text-sm font-medium text-gray-200 truncate">{r.market}</h3>
              <div className="flex items-center gap-3 text-xs text-muted mt-1">
                <span>Kalshi: <span className="text-gray-200">{r.kalshi}</span></span>
                <span>Poly: <span className="text-gray-200">{r.poly}</span></span>
                <span>Spread: <span className="text-green">{r.spread}</span></span>
                <span>Score: <span className={r.score >= 85 ? "text-green" : "text-amber"}>{r.score}%</span></span>
              </div>
            </div>
            <div className="flex gap-2">
              <button className="bg-green/10 text-green text-xs px-3 py-1.5 rounded-lg hover:bg-green/20 transition-colors">Match</button>
              <button className="bg-red/10 text-red text-xs px-3 py-1.5 rounded-lg hover:bg-red/20 transition-colors">Dismiss</button>
              <button className="bg-surface-hover text-muted text-xs px-3 py-1.5 rounded-lg hover:text-gray-200 transition-colors">Details</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

const reviews = [
  { market: "Will DOW hit 45K by July?", kalshi: "$0.42", poly: "$0.38", spread: "4.0¢", score: 94 },
  { market: "BTC > $120K in June?", kalshi: "$0.31", poly: "$0.27", spread: "4.0¢", score: 91 },
  { market: "Fed cuts rates in Q3?", kalshi: "$0.65", poly: "$0.61", spread: "4.0¢", score: 87 },
  { market: "S&P 500 > 5600 EOM?", kalshi: "$0.48", poly: "$0.44", spread: "4.0¢", score: 85 },
  { market: "ETH ETF approved by Aug?", kalshi: "$0.55", poly: "$0.50", spread: "5.0¢", score: 82 },
  { market: "Apple > $250 by Sept?", kalshi: "$0.37", poly: "$0.33", spread: "4.0¢", score: 78 },
  { market: "US GDP growth > 2%?", kalshi: "$0.58", poly: "$0.54", spread: "4.0¢", score: 73 },
  { market: "Oil > $90 by Dec?", kalshi: "$0.35", poly: "$0.31", spread: "4.0¢", score: 71 },
  { market: "NFLX subscriber beat?", kalshi: "$0.62", poly: null, spread: "—", score: 45 },
]
