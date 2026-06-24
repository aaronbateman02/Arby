"use client"

import { useState } from "react"
import { timePeriods, periodStats, strategiesByPeriod } from "@/lib/data"
import { VenueCards } from "@/components/VenueCards"
import { PeriodSelector } from "@/components/PeriodSelector"
import { PeriodStats } from "@/components/PeriodStats"
import { BundleTable } from "@/components/BundleTable"
import { StrategyRanking } from "@/components/StrategyRanking"

export default function DashboardPage() {
  const [period, setPeriod] = useState(timePeriods[2])

  const stats = periodStats[period]
  const ranking = strategiesByPeriod[period]

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-100">Dashboard</h1>
        <p className="text-sm text-muted mt-1">Cross-venue arbitrage overview</p>
      </div>

      <section>
        <h2 className="text-sm font-semibold text-gray-200 mb-3">Account Portfolio</h2>
        <VenueCards />
      </section>

      <section>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-sm font-semibold text-gray-200">Performance</h2>
          <PeriodSelector periods={timePeriods} selected={period} onSelect={setPeriod} />
        </div>
        <PeriodStats stats={stats} />
      </section>

      <section>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-sm font-semibold text-gray-200">Strategy Rankings</h2>
        </div>
        <StrategyRanking best={ranking.best} worst={ranking.worst} period={period} />
      </section>

      <section>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-sm font-semibold text-gray-200">Soonest Resolving</h2>
        </div>
        <BundleTable />
      </section>
    </div>
  )
}
