"use client"

import { useState } from "react"
import { appSettings, availableModels } from "@/lib/data"

export default function SettingsPage() {
  const [settings, setSettings] = useState(structuredClone(appSettings))
  const [saved, setSaved] = useState(false)

  const updateKalshi = <K extends keyof typeof settings.venueKeys.kalshi>(key: K, value: (typeof settings.venueKeys.kalshi)[K]) =>
    setSettings((s) => ({ ...s, venueKeys: { ...s.venueKeys, kalshi: { ...s.venueKeys.kalshi, [key]: value } } }))

  const updatePoly = <K extends keyof typeof settings.venueKeys.polymarket>(key: K, value: (typeof settings.venueKeys.polymarket)[K]) =>
    setSettings((s) => ({ ...s, venueKeys: { ...s.venueKeys, polymarket: { ...s.venueKeys.polymarket, [key]: value } } }))

  const updateTrading = <K extends keyof typeof settings.trading>(key: K, value: (typeof settings.trading)[K]) =>
    setSettings((s) => ({ ...s, trading: { ...s.trading, [key]: value } }))

  const updateReview = <K extends keyof typeof settings.pairReview>(key: K, value: (typeof settings.pairReview)[K]) =>
    setSettings((s) => ({ ...s, pairReview: { ...s.pairReview, [key]: value } }))

  const handleSave = () => {
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Settings</h1>
          <p className="text-sm text-muted mt-1">Configure venues, trading defaults, and AI pair review</p>
        </div>
        <button
          onClick={handleSave}
          className={`px-4 py-2 text-sm rounded-lg transition-colors ${saved ? "bg-green text-white" : "bg-accent text-white hover:bg-accent-hover"}`}
        >
          {saved ? "Saved!" : "Save Changes"}
        </button>
      </div>

      <Section title="Venue API Keys" subtitle="Credentials for exchange connectivity">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-3">
            <div className="flex items-center gap-2 mb-2">
              <span className="w-2 h-2 rounded-full bg-accent" />
              <span className="text-sm font-semibold text-gray-200">Kalshi</span>
            </div>
            <TextInput label="Email" value={settings.venueKeys.kalshi.email} onChange={(v) => updateKalshi("email", v)} />
            <TextInput label="Key ID" value={settings.venueKeys.kalshi.keyId} onChange={(v) => updateKalshi("keyId", v)} />
            <TextInput label="Private Key Path" value={settings.venueKeys.kalshi.privateKey} onChange={(v) => updateKalshi("privateKey", v)} />
          </div>
          <div className="space-y-3">
            <div className="flex items-center gap-2 mb-2">
              <span className="w-2 h-2 rounded-full bg-amber" />
              <span className="text-sm font-semibold text-gray-200">Polymarket</span>
            </div>
            <TextInput label="API Key" value={settings.venueKeys.polymarket.apiKey} onChange={(v) => updatePoly("apiKey", v)} />
            <TextInput label="Wallet Private Key Path" value={settings.venueKeys.polymarket.walletPrivateKey} onChange={(v) => updatePoly("walletPrivateKey", v)} />
            <TextInput label="Wallet Address" value={settings.venueKeys.polymarket.walletAddress} onChange={(v) => updatePoly("walletAddress", v)} />
          </div>
        </div>
      </Section>

      <Section title="Trading Defaults" subtitle="Global trading parameters applied to all strategies">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <NumberInput label="Max Position Size ($)" value={settings.trading.maxPositionSize} onChange={(v) => updateTrading("maxPositionSize", v)} min={0} />
          <NumberInput label="Min Spread Threshold (¢)" value={settings.trading.minSpreadThreshold} onChange={(v) => updateTrading("minSpreadThreshold", v)} min={0} step={0.5} />
          <NumberInput label="Max Active Bundles" value={settings.trading.maxActiveBundles} onChange={(v) => updateTrading("maxActiveBundles", v)} min={1} />
          <NumberInput label="Default Min ROI (%)" value={settings.trading.defaultMinRoi} onChange={(v) => updateTrading("defaultMinRoi", v)} min={0} step={0.5} />
          <SelectInput label="Log Level" value={settings.trading.logLevel} options={["debug", "info", "warn", "error"]} onChange={(v) => updateTrading("logLevel", v)} />
          <NumberInput label="Metrics Port" value={settings.trading.metricsPort} onChange={(v) => updateTrading("metricsPort", v)} min={1024} max={65535} />
        </div>
        <div className="flex flex-wrap gap-6 mt-4 pt-4 border-t border-border">
          <Toggle label="Auto-match approved pairs" value={settings.trading.autoMatchApproved} onChange={(v) => updateTrading("autoMatchApproved", v)} />
          <Toggle label="Email notifications" value={settings.trading.emailNotifications} onChange={(v) => updateTrading("emailNotifications", v)} />
          <Toggle label="Paper mode (simulate all trades)" value={settings.trading.paperMode} onChange={(v) => updateTrading("paperMode", v)} />
        </div>
      </Section>

      <Section title="AI Pair Review" subtitle="Configure which models review each leg and how pairs are batched">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-3">
            <div className="flex items-center gap-2 mb-2">
              <span className="w-2 h-2 rounded-full bg-accent" />
              <span className="text-sm font-semibold text-gray-200">Leg A Review Model</span>
              <span className="text-[10px] text-muted bg-surface-hover px-1.5 py-0.5 rounded">Kalshi</span>
            </div>
            <ModelSelect value={settings.pairReview.legAModel} onChange={(v) => updateReview("legAModel", v)} />
            <p className="text-xs text-muted">Reviews the Kalshi-side of each pair — evaluates spread feasibility, liquidity depth, and execution risk.</p>
          </div>
          <div className="space-y-3">
            <div className="flex items-center gap-2 mb-2">
              <span className="w-2 h-2 rounded-full bg-amber" />
              <span className="text-sm font-semibold text-gray-200">Leg B Review Model</span>
              <span className="text-[10px] text-muted bg-surface-hover px-1.5 py-0.5 rounded">Polymarket</span>
            </div>
            <ModelSelect value={settings.pairReview.legBModel} onChange={(v) => updateReview("legBModel", v)} />
            <p className="text-xs text-muted">Reviews the Polymarket-side — assesses CLOB liquidity, fill probability, and adverse selection risk.</p>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mt-4 pt-4 border-t border-border">
          <NumberInput label="Batch Size (first leg)" value={settings.pairReview.batchSize} onChange={(v) => updateReview("batchSize", v)} min={1} max={200} help="Number of candidate pairs sent to the Leg A model in each review batch." />
          <NumberInput label="Confidence Threshold" value={Math.round(settings.pairReview.confidenceThreshold * 100)} onChange={(v) => updateReview("confidenceThreshold", v / 100)} min={0} max={100} suffix="%" help="Pairs scoring above this threshold are auto-approved. Below goes to manual review." />
          <Toggle label="Auto-approve above threshold" value={settings.pairReview.autoApproveAboveThreshold} onChange={(v) => updateReview("autoApproveAboveThreshold", v)} />
        </div>

        <div className="mt-4">
          <label className="text-xs text-muted block mb-1">
            Review Prompt Template
          </label>
          <textarea
            className="w-full bg-surface border border-border rounded-lg px-3 py-2 text-sm text-gray-200 outline-none focus:border-accent resize-y min-h-[72px]"
            value={settings.pairReview.reviewPromptTemplate}
            onChange={(e) => updateReview("reviewPromptTemplate", e.target.value)}
          />
          <p className="text-xs text-muted mt-1">Custom prompt sent to the model alongside each pair. Leave default for recommended behavior.</p>
        </div>
      </Section>
    </div>
  )
}

function Section({ title, subtitle, children }: { title: string; subtitle: string; children: React.ReactNode }) {
  return (
    <div className="bg-surface-alt rounded-xl border border-border p-5">
      <div className="mb-4">
        <h2 className="text-sm font-semibold text-gray-200">{title}</h2>
        <p className="text-xs text-muted mt-0.5">{subtitle}</p>
      </div>
      {children}
    </div>
  )
}

function TextInput({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <label className="text-xs text-muted block">
      {label}
      <input
        type="text"
        className="mt-1 w-full bg-surface border border-border rounded-lg px-3 py-1.5 text-sm text-gray-200 outline-none focus:border-accent font-mono"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
    </label>
  )
}

function NumberInput({ label, value, onChange, min, max, step, suffix, help }: {
  label: string; value: number; onChange: (v: number) => void; min?: number; max?: number; step?: number; suffix?: string; help?: string
}) {
  return (
    <label className="text-xs text-muted block">
      <span className="flex items-center gap-1">
        {label}
        {help && <HelpBubble text={help} />}
      </span>
      <div className="relative mt-1">
        <input
          type="number"
          min={min}
          max={max}
          step={step ?? 1}
          className="w-full bg-surface border border-border rounded-lg px-3 py-1.5 text-sm text-gray-200 outline-none focus:border-accent"
          value={value}
          onChange={(e) => onChange(Number(e.target.value))}
        />
        {suffix && <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-muted">{suffix}</span>}
      </div>
    </label>
  )
}

function SelectInput({ label, value, options, onChange }: { label: string; value: string; options: string[]; onChange: (v: string) => void }) {
  return (
    <label className="text-xs text-muted block">
      {label}
      <select
        className="mt-1 w-full bg-surface border border-border rounded-lg px-3 py-1.5 text-sm text-gray-200 outline-none focus:border-accent"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      >
        {options.map((o) => <option key={o} value={o}>{o}</option>)}
      </select>
    </label>
  )
}

function Toggle({ label, value, onChange }: { label: string; value: boolean; onChange: (v: boolean) => void }) {
  return (
    <label className="flex items-center gap-2 text-sm text-muted cursor-pointer">
      <button
        type="button"
        onClick={() => onChange(!value)}
        className={`w-10 h-5 rounded-full relative transition-colors shrink-0 ${value ? "bg-accent" : "bg-border"}`}
      >
        <div className={`absolute top-0.5 w-4 h-4 rounded-full bg-white transition-transform ${value ? "translate-x-5" : "translate-x-0.5"}`} />
      </button>
      {label}
    </label>
  )
}

function ModelSelect({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  return (
    <select
      className="w-full bg-surface border border-border rounded-lg px-3 py-2 text-sm text-gray-200 outline-none focus:border-accent"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    >
      {availableModels.map((m) => <option key={m} value={m}>{m}</option>)}
    </select>
  )
}

function HelpBubble({ text }: { text: string }) {
  const [open, setOpen] = useState(false)
  return (
    <span className="relative inline-flex">
      <button
        type="button"
        onMouseEnter={() => setOpen(true)}
        onMouseLeave={() => setOpen(false)}
        onClick={() => setOpen(!open)}
        className="inline-flex items-center justify-center w-3.5 h-3.5 rounded-full bg-border text-muted hover:text-gray-200 hover:bg-muted transition-colors text-[10px] font-bold leading-none"
      >
        ?
      </button>
      {open && (
        <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 w-56 z-50">
          <div className="bg-gray-900 text-gray-100 text-xs rounded-lg px-3 py-2 shadow-lg border border-gray-700 leading-relaxed">
            {text}
          </div>
          <div className="absolute top-full left-1/2 -translate-x-1/2 w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-gray-900" />
        </div>
      )}
    </span>
  )
}
