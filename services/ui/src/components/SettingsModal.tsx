"use client"

import { useState } from "react"
import type { StrategyDetail, StrategyConfig } from "@/lib/data"

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
            <Field label="Min ROI (%)" value={config.minRoi} onChange={(v) => set("minRoi", v)} step={0.1} min={0} />
            <Field label="Max Positions / Bundle" value={config.maxPositionsPerBundle} onChange={(v) => set("maxPositionsPerBundle", v)} min={1} />
            <Field label="Max Position ($)" value={config.maxPositionDollars} onChange={(v) => set("maxPositionDollars", v)} min={0} />
            <Field label="Min Spread (¢)" value={config.minSpread} onChange={(v) => set("minSpread", v)} step={0.5} min={0} />
            <Field label="Max Daily Exposure ($)" value={config.maxDailyExposure} onChange={(v) => set("maxDailyExposure", v)} min={0} />
            <Field label="Cooldown (min)" value={config.cooldownMinutes} onChange={(v) => set("cooldownMinutes", v)} min={0} />
            <Field label="Max Days to Resolution" value={config.maxDaysToResolution} onChange={(v) => set("maxDaysToResolution", v)} min={1} />
          </div>

          <div className="space-y-3 pt-3 border-t border-border">
            <Toggle label="Auto-Execute" value={config.autoExecute} onChange={(v) => set("autoExecute", v)} />
            <Toggle label="Notify on Match" value={config.notifyOnMatch} onChange={(v) => set("notifyOnMatch", v)} />
            <Toggle label="Enabled" value={config.enabled} onChange={(v) => set("enabled", v)} />
            <Toggle label="Paper Mode" value={config.paperMode} onChange={(v) => set("paperMode", v)} />
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

function Field({ label, value, onChange, step, min }: {
  label: string; value: number; onChange: (v: number) => void; step?: number; min?: number
}) {
  return (
    <label className="text-xs text-muted">
      {label}
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

function Toggle({ label, value, onChange }: { label: string; value: boolean; onChange: (v: boolean) => void }) {
  return (
    <label className="flex items-center justify-between">
      <span className="text-sm text-muted">{label}</span>
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
