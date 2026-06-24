interface Strategy {
  name: string
  totalPnL: number
  totalReturn: number
  winRate: number
  totalBundles: number
  successfulBundles: number
}

export function StrategyRanking({ best, worst, period }: { best: Strategy; worst: Strategy; period: string }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div className="bg-surface-alt rounded-xl border border-border p-4 border-l-4 border-l-green">
        <div className="flex items-center gap-2 mb-3">
          <span className="text-xs px-1.5 py-0.5 rounded text-green bg-green/10 font-medium">Best</span>
          <span className="text-sm font-semibold text-gray-200">{best.name}</span>
          <span className="text-xs text-muted">({period})</span>
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
          <div><span className="text-muted">P&L</span><p className="text-sm font-medium text-green">+${best.totalPnL.toLocaleString()}</p></div>
          <div><span className="text-muted">Return</span><p className="text-sm font-medium text-green">{best.totalReturn}%</p></div>
          <div><span className="text-muted">Win Rate</span><p className="text-sm font-medium text-gray-200">{best.winRate}%</p></div>
          <div><span className="text-muted">Bundles</span><p className="text-sm font-medium text-gray-200">{best.successfulBundles}/{best.totalBundles}</p></div>
        </div>
      </div>

      <div className="bg-surface-alt rounded-xl border border-border p-4 border-l-4 border-l-red">
        <div className="flex items-center gap-2 mb-3">
          <span className="text-xs px-1.5 py-0.5 rounded text-red bg-red/10 font-medium">Worst</span>
          <span className="text-sm font-semibold text-gray-200">{worst.name}</span>
          <span className="text-xs text-muted">({period})</span>
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
          <div><span className="text-muted">P&L</span><p className="text-sm font-medium text-gray-200">+${worst.totalPnL.toLocaleString()}</p></div>
          <div><span className="text-muted">Return</span><p className="text-sm font-medium text-amber">{worst.totalReturn}%</p></div>
          <div><span className="text-muted">Win Rate</span><p className="text-sm font-medium text-gray-200">{worst.winRate}%</p></div>
          <div><span className="text-muted">Bundles</span><p className="text-sm font-medium text-gray-200">{worst.successfulBundles}/{worst.totalBundles}</p></div>
        </div>
      </div>
    </div>
  )
}
