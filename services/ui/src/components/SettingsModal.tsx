"use client"

import { useState } from "react"
import type { StrategyDetail, StrategyConfig } from "@/lib/data"

const helpText: Record<string, string> = {
  "Min ROI (%)": "The minimum return on investment required before a pair is eligible for execution. Pairs below this threshold will be monitored but not traded until they cross this bar.",
  "Max Positions / Bundle": "How many times the strategy can execute on a single bundle/opportunity. Prevents over-concentration on one market pair.",
  "Max Position ($)": "The maximum dollar amount allocated to a single position. Limits per-trade exposure to a manageable level.",
  "Min Spread (¢)": "The minimum price gap in cents between the two venues' markets. Narrower spreads may not cover execution costs.",
  "Max Daily Exposure ($)": "The total dollar limit across all positions opened in a single day. Hard cap on daily risk.",
  "Cooldown (min)": "Minutes to wait after executing a bundle before re-evaluating the same pair. Prevents rapid-fire re-entries on the same opportunity.",
  "Max Days to Resolution": "Filters out pairs whose underlying market resolves too far in the future. Keeps capital locked up for a manageable duration.",
  "Auto-Execute": "When enabled, the strategy will automatically submit orders when a pair meets all criteria. Disable to review opportunities manually.",
  "Notify on Match": "Sends a notification when a qualifying pair is found. Useful when auto-execute is off so you can review and approve manually.",
  "Enabled": "Turns the strategy on or off. Disabled strategies will not scan for pairs or execute any trades.",
  "Paper Mode": "When enabled, all trades are simulated (paper traded) rather than submitted to real venues. Use for testing and validation.",
}

export function SettingsModal({ strategy, onClose, onSave }: {
  strategy: StrategyDetail
  onClose: () => void
  onSave: (updated: StrategyDetail) => void
}) {
  const [config, setConfig] = useState<StrategyConfig>({ ...strategy.config })
  const [saving, setSaving] = useState(false)

  const set = <K extends keyof StrategyConfig>(key: K, value: StrategyConfig[K]) =>
    setConfig((c) => ({ ...c, [key]: value }))

  const handleSave = () => {
    setSaving(true)
    setTimeout(() => {
      onSave({ ...strategy, config })
      setSaving(false)
    }, 400)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <div className="bg-surface rounded-xl border border-border shadow-xl max-w-lg w-full mx-4 max-h-[85vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between p-5 border-b border-border">
          <div>
            <h2 className="text-lg font-bold text-gray-100">{strategy.name}</h2>
            <p className="text-xs text-muted mt-0.5 font-mono">{strategy.id}</p>
          </div>
          <button onClick={onClose} className="text-muted hover:text-gray-200 p-1">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="p-5 space-y-5">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field label="Min ROI (%)" help={helpText["Min ROI (%)"]} value={config.minRoi} onChange={(v) => set("minRoi", v)} step={0.1} min={0} />
            <Field label="Max Positions / Bundle" help={helpText["Max Positions / Bundle"]} value={config.maxPositionsPerBundle} onChange={(v) => set("maxPositionsPerBundle", v)} min={1} />
            <Field label="Max Position ($)" help={helpText["Max Position ($)"]} value={config.maxPositionDollars} onChange={(v) => set("maxPositionDollars", v)} min={0} />
            <Field label="Min Spread (¢)" help={helpText["Min Spread (¢)"]} value={config.minSpread} onChange={(v) => set("minSpread", v)} step={0.5} min={0} />
            <Field label="Max Daily Exposure ($)" help={helpText["Max Daily Exposure ($)"]} value={config.maxDailyExposure} onChange={(v) => set("maxDailyExposure", v)} min={0} />
            <Field label="Cooldown (min)" help={helpText["Cooldown (min)"]} value={config.cooldownMinutes} onChange={(v) => set("cooldownMinutes", v)} min={0} />
            <Field label="Max Days to Resolution" help={helpText["Max Days to Resolution"]} value={config.maxDaysToResolution} onChange={(v) => set("maxDaysToResolution", v)} min={1} />
          </div>

          <div className="space-y-3 pt-3 border-t border-border">
            <Toggle label="Auto-Execute" help={helpText["Auto-Execute"]} value={config.autoExecute} onChange={(v) => set("autoExecute", v)} />
            <Toggle label="Notify on Match" help={helpText["Notify on Match"]} value={config.notifyOnMatch} onChange={(v) => set("notifyOnMatch", v)} />
            <Toggle label="Enabled" help={helpText["Enabled"]} value={config.enabled} onChange={(v) => set("enabled", v)} />
            <Toggle label="Paper Mode" help={helpText["Paper Mode"]} value={config.paperMode} onChange={(v) => set("paperMode", v)} />
          </div>
        </div>

        <div className="flex items-center justify-end gap-2 p-5 border-t border-border">
          <button onClick={onClose} className="px-3 py-1.5 text-sm rounded-lg border border-border text-muted hover:text-gray-200 transition-colors">Cancel</button>
          <button onClick={handleSave} disabled={saving} className="px-4 py-1.5 text-sm rounded-lg bg-accent text-white hover:bg-accent-hover disabled:opacity-70 transition-colors">
            {saving ? "Saving..." : "Save Changes"}
          </button>
        </div>
      </div>
    </div>
  )
}

function Field({ label, help, value, onChange, step, min }: {
  label: string; help: string; value: number; onChange: (v: number) => void; step?: number; min?: number
}) {
  return (
    <label className="text-xs text-muted">
      <span className="flex items-center gap-1">
        {label}
        <HelpBubble text={help} />
      </span>
      <input
        type="number"
        step={step ?? 1}
        min={min ?? 0}
        className="mt-1 w-full bg-surface border border-border rounded-lg px-3 py-1.5 text-sm text-gray-200 outline-none focus:border-accent"
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
      />
    </label>
  )
}

function Toggle({ label, help, value, onChange }: { label: string; help: string; value: boolean; onChange: (v: boolean) => void }) {
  return (
    <label className="flex items-center justify-between">
      <span className="flex items-center gap-1 text-sm text-muted">
        {label}
        <HelpBubble text={help} />
      </span>
      <button
        type="button"
        onClick={() => onChange(!value)}
        className={`w-10 h-5 rounded-full relative transition-colors ${value ? "bg-accent" : "bg-border"}`}
      >
        <div className={`absolute top-0.5 w-4 h-4 rounded-full bg-white transition-transform ${value ? "translate-x-5" : "translate-x-0.5"}`} />
      </button>
    </label>
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
        <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 w-56 z-60">
          <div className="bg-gray-900 text-gray-100 text-xs rounded-lg px-3 py-2 shadow-lg border border-gray-700 leading-relaxed">
            {text}
          </div>
          <div className="absolute top-full left-1/2 -translate-x-1/2 w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-gray-900" />
        </div>
      )}
    </span>
  )
}
