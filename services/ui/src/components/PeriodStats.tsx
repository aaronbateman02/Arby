import { StatCard } from "./StatCard"

interface Stats {
  openPositions: number
  projectedProfit: number
  roi: number
  successfulFills: number
  aborted: number
}

export function PeriodStats({ stats }: { stats: Stats }) {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
      <StatCard label="Open Positions" value={String(stats.openPositions)} />
      <StatCard label="Projected Profit" value={`$${stats.projectedProfit.toLocaleString()}`} positive />
      <StatCard label="ROI" value={`${stats.roi}%`} positive />
      <StatCard label="Filled" value={String(stats.successfulFills)} positive />
      <StatCard label="Aborted" value={String(stats.aborted)} negative />
    </div>
  )
}
