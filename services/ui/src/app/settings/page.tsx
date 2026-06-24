export default function SettingsPage() {
  return (
    <div className="p-6 max-w-3xl mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-100">Settings</h1>
        <p className="text-sm text-muted mt-1">Configure API keys and preferences</p>
      </div>

      <Section title="Venue API Keys">
        <SettingRow label="Kalshi Email" type="text" placeholder="user@example.com" />
        <SettingRow label="Kalshi Key ID" type="password" placeholder="••••••••" />
        <SettingRow label="Kalshi Private Key" type="password" placeholder="PEM encoded..." />
        <SettingRow label="Polymarket API Key" type="password" placeholder="••••••••" />
        <SettingRow label="Polymarket Wallet Key" type="password" placeholder="0x..." />
      </Section>

      <Section title="Trading Preferences">
        <SettingRow label="Max Position Size" type="text" placeholder="$1,000" />
        <SettingRow label="Min Spread Threshold" type="text" placeholder="3.0¢" />
        <SettingRow label="Max Active Bundles" type="text" placeholder="10" />
        <ToggleRow label="Auto-match approved pairs" enabled />
        <ToggleRow label="Email notifications" enabled={false} />
      </Section>

      <Section title="System">
        <SettingRow label="Log Level" type="select" value="info" />
        <SettingRow label="Metrics Port" type="text" placeholder="8086" />
      </Section>
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="mb-8">
      <h2 className="text-sm font-semibold text-gray-200 mb-3">{title}</h2>
      <div className="bg-surface-alt rounded-xl border border-border divide-y divide-border/50">
        {children}
      </div>
    </div>
  )
}

function SettingRow({ label, type, placeholder, value: val }: {
  label: string; type: string; placeholder?: string; value?: string
}) {
  return (
    <div className="flex items-center justify-between px-4 py-3">
      <span className="text-sm text-muted">{label}</span>
      {type === "select" ? (
        <select className="bg-surface border border-border rounded-lg px-3 py-1.5 text-sm text-gray-200 outline-none w-48">
          <option>{val}</option>
        </select>
      ) : (
        <input
          type={type}
          className="bg-surface border border-border rounded-lg px-3 py-1.5 text-sm text-gray-200 placeholder-muted outline-none w-48 focus:border-accent"
          placeholder={placeholder}
          readOnly
        />
      )}
    </div>
  )
}

function ToggleRow({ label, enabled }: { label: string; enabled: boolean }) {
  return (
    <div className="flex items-center justify-between px-4 py-3">
      <span className="text-sm text-muted">{label}</span>
      <div className={`w-10 h-5 rounded-full relative transition-colors ${enabled ? "bg-accent" : "bg-border"}`}>
        <div className={`absolute top-0.5 w-4 h-4 rounded-full bg-white transition-transform ${enabled ? "translate-x-5" : "translate-x-0.5"}`} />
      </div>
    </div>
  )
}
