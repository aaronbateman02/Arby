export default function DashboardPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Dashboard</h1>
          <p className="text-sm text-muted mt-1">Cross-venue arbitrage overview</p>
        </div>
        <div className="flex items-center gap-3 text-sm">
          <span className="text-green flex items-center gap-1">
            <span className="w-2 h-2 rounded-full bg-green" />
            All systems nominal
          </span>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <MetricCard label="Active Opportunities" value="23" change="+4" positive />
        <MetricCard label="Matched Today" value="$142.3K" change="+12.3%" positive />
        <MetricCard label="Pending Review" value="157" change="+" context />
        <MetricCard label="Total P&L" value="$18,422" change="+$2,143" positive />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        <div className="lg:col-span-2 bg-surface-alt rounded-xl border border-border p-5">
          <h2 className="text-sm font-semibold text-gray-200 mb-4">Cross-Venue Spreads</h2>
          <div className="space-y-3">
            <SpreadRow market="Will DOW hit 45K by July?" kalshi="$0.42" poly="$0.38" spread="4.0¢" dir="+4" />
            <SpreadRow market="BTC > $120K in June?" kalshi="$0.31" poly="$0.27" spread="4.0¢" dir="+4" />
            <SpreadRow market="Fed cuts rates in Q3?" kalshi="$0.65" poly="$0.61" spread="4.0¢" dir="+4" />
            <SpreadRow market="S&P 500 > 5600 EOM?" kalshi="$0.48" poly="$0.44" spread="4.0¢" dir="+4" />
            <SpreadRow market="ETH ETF approved by Aug?" kalshi="$0.55" poly="$0.50" spread="5.0¢" dir="+5" />
          </div>
        </div>
        <div className="bg-surface-alt rounded-xl border border-border p-5">
          <h2 className="text-sm font-semibold text-gray-200 mb-4">Recent Matches</h2>
          <div className="space-y-4">
            <ActivityItem title="BTC > $120K" status="matched" time="2m ago" />
            <ActivityItem title="DOW > 45K July" status="matched" time="5m ago" />
            <ActivityItem title="Fed cuts Q3" status="pending" time="12m ago" />
            <ActivityItem title="S&P > 5600" status="matched" time="18m ago" />
            <ActivityItem title="ETH ETF" status="pending" time="25m ago" />
          </div>
        </div>
      </div>

      <div className="bg-surface-alt rounded-xl border border-border p-5">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-sm font-semibold text-gray-200">Active Bundles</h2>
          <span className="text-xs text-muted">3 running</span>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <BundleCard
            name="BTC June Cross"
            markets={2}
            spread="4.0¢"
            exposure="$1,200"
            status="active"
          />
          <BundleCard
            name="DOW July Spread"
            markets={2}
            spread="4.0¢"
            exposure="$850"
            status="active"
          />
          <BundleCard
            name="ETH ETF Basket"
            markets={4}
            spread="5.0¢"
            exposure="$2,100"
            status="pending"
          />
        </div>
      </div>
    </div>
  )
}

function MetricCard({ label, value, change, positive, context: ctx }: {
  label: string; value: string; change: string; positive?: boolean; context?: boolean
}) {
  return (
    <div className="bg-surface-alt rounded-xl border border-border p-4">
      <p className="text-xs text-muted mb-1">{label}</p>
      <p className="text-2xl font-bold text-gray-100">{value}</p>
      <p className={`text-sm mt-1 ${ctx ? "text-accent" : positive ? "text-green" : "text-red"}`}>
        {change}
      </p>
    </div>
  )
}

function SpreadRow({ market, kalshi, poly, spread, dir }: {
  market: string; kalshi: string; poly: string; spread: string; dir: string
}) {
  return (
    <div className="flex items-center justify-between py-2 px-3 rounded-lg hover:bg-surface-hover transition-colors">
      <span className="text-sm text-gray-200 flex-1 truncate">{market}</span>
      <div className="flex items-center gap-4 text-sm">
        <span className="text-muted w-14 text-right">{kalshi}</span>
        <span className="text-muted">→</span>
        <span className="text-muted w-14 text-right">{poly}</span>
        <span className="w-12 text-right text-green font-medium">{spread}</span>
      </div>
    </div>
  )
}

function ActivityItem({ title, status, time }: {
  title: string; status: string; time: string
}) {
  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-2">
        <span className={`w-2 h-2 rounded-full ${status === "matched" ? "bg-green" : "bg-amber"}`} />
        <span className="text-sm text-gray-200">{title}</span>
      </div>
      <div className="flex items-center gap-2">
        <span className={`text-xs px-1.5 py-0.5 rounded ${status === "matched" ? "text-green bg-green/10" : "text-amber bg-amber/10"}`}>
          {status}
        </span>
        <span className="text-xs text-muted">{time}</span>
      </div>
    </div>
  )
}

function BundleCard({ name, markets, spread, exposure, status }: {
  name: string; markets: number; spread: string; exposure: string; status: string
}) {
  return (
    <div className="bg-surface rounded-lg border border-border p-4">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-sm font-semibold text-gray-200">{name}</h3>
        <span className={`text-xs px-1.5 py-0.5 rounded ${status === "active" ? "text-green bg-green/10" : "text-amber bg-amber/10"}`}>
          {status}
        </span>
      </div>
      <div className="grid grid-cols-3 gap-2 text-xs text-muted">
        <div><span className="text-gray-300">{markets}</span> markets</div>
        <div><span className="text-green">{spread}</span> spread</div>
        <div><span className="text-gray-300">{exposure}</span> exposure</div>
      </div>
    </div>
  )
}
