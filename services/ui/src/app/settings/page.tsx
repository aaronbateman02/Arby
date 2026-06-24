"use client"

import { useState, useEffect, useCallback } from "react"

type OpenRouterModel = { id: string; name: string; description?: string }
type Settings = Record<string, unknown>

export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null)
  const [loading, setLoading] = useState(true)
  const [saved, setSaved] = useState(false)
  const [saveError, setSaveError] = useState("")
  const [models, setModels] = useState<OpenRouterModel[]>([])
  const [loadingModels, setLoadingModels] = useState(false)
  const [modelError, setModelError] = useState("")

  useEffect(() => {
    fetch("/api/settings")
      .then((r) => r.json())
      .then((data) => setSettings(data))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const update = (path: string[], value: unknown) => {
    setSettings((prev) => {
      if (!prev) return prev
      const next = structuredClone(prev)
      let obj = next as Record<string, unknown>
      for (let i = 0; i < path.length - 1; i++) {
        const key = path[i]
        if (!(key in obj)) obj[key] = {}
        obj = obj[key] as Record<string, unknown>
      }
      obj[path[path.length - 1]] = value
      return next
    })
  }

  const fetchModels = useCallback(async () => {
    const key = settings?.openrouterApiKey as string | undefined
    if (!key) return
    setLoadingModels(true)
    setModelError("")
    try {
      const res = await fetch("https://openrouter.ai/api/v1/models", {
        headers: { Authorization: `Bearer ${key}` },
      })
      if (!res.ok) throw new Error(`OpenRouter returned ${res.status}`)
      const json = await res.json()
      setModels(json.data ?? [])
      if (!json.data?.length) setModelError("No models returned")
    } catch (e) {
      setModelError((e as Error).message)
      setModels([])
    } finally {
      setLoadingModels(false)
    }
  }, [settings?.openrouterApiKey])

  useEffect(() => {
    if ((settings?.openrouterApiKey as string | undefined)?.length) fetchModels()
  }, [])

  const handleSave = async () => {
    if (!settings) return
    setSaveError("")
    try {
      const res = await fetch("/api/settings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(settings),
      })
      const json = await res.json()
      if (!json.ok) throw new Error(json.error || "save failed")
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch (e) {
      setSaveError((e as Error).message)
    }
  }

  if (loading) {
    return (
      <div className="p-6 max-w-4xl mx-auto">
        <div className="text-sm text-muted animate-pulse">Loading settings...</div>
      </div>
    )
  }

  if (!settings) {
    return (
      <div className="p-6 max-w-4xl mx-auto">
        <div className="text-sm text-red-400">Failed to load settings</div>
      </div>
    )
  }

  const vk = settings.venueKeys as Record<string, Record<string, string>>
  const tr = settings.trading as Record<string, unknown>
  const pr = settings.pairReview as Record<string, unknown>
  const modelOptions = models.map((m) => ({ value: m.id, label: m.name || m.id }))

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Settings</h1>
          <p className="text-sm text-muted mt-1">Configure venues, trading defaults, and AI pair review</p>
        </div>
        <div className="flex items-center gap-3">
          {saveError && <span className="text-xs text-red-400">{saveError}</span>}
          <button
            onClick={handleSave}
            className={`px-4 py-2 text-sm rounded-lg transition-colors ${saved ? "bg-green text-white" : "bg-accent text-white hover:bg-accent-hover"}`}
          >
            {saved ? "Saved!" : "Save Changes"}
          </button>
        </div>
      </div>

      <Section title="Venue API Keys" subtitle="Credentials for exchange connectivity">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-3">
            <div className="flex items-center gap-2 mb-2">
              <span className="w-2 h-2 rounded-full bg-accent" />
              <span className="text-sm font-semibold text-gray-200">Kalshi</span>
            </div>
            <TextInput label="Email" value={(vk.kalshi?.email as string) ?? ""} onChange={(v) => update(["venueKeys", "kalshi", "email"], v)} />
            <TextInput label="Key ID" value={(vk.kalshi?.keyId as string) ?? ""} onChange={(v) => update(["venueKeys", "kalshi", "keyId"], v)} />
            <TextInput label="Private Key Path" value={(vk.kalshi?.privateKey as string) ?? ""} onChange={(v) => update(["venueKeys", "kalshi", "privateKey"], v)} />
          </div>
          <div className="space-y-3">
            <div className="flex items-center gap-2 mb-2">
              <span className="w-2 h-2 rounded-full bg-amber" />
              <span className="text-sm font-semibold text-gray-200">Polymarket</span>
            </div>
            <TextInput label="API Key" value={(vk.polymarket?.apiKey as string) ?? ""} onChange={(v) => update(["venueKeys", "polymarket", "apiKey"], v)} />
            <TextInput label="Wallet Private Key Path" value={(vk.polymarket?.walletPrivateKey as string) ?? ""} onChange={(v) => update(["venueKeys", "polymarket", "walletPrivateKey"], v)} />
            <TextInput label="Wallet Address" value={(vk.polymarket?.walletAddress as string) ?? ""} onChange={(v) => update(["venueKeys", "polymarket", "walletAddress"], v)} />
          </div>
        </div>
      </Section>

      <Section title="Trading Defaults" subtitle="Global trading parameters applied to all strategies">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <NumberInput label="Max Position Size ($)" value={tr.maxPositionSize as number} onChange={(v) => update(["trading", "maxPositionSize"], v)} min={0} />
          <NumberInput label="Min Spread Threshold (¢)" value={tr.minSpreadThreshold as number} onChange={(v) => update(["trading", "minSpreadThreshold"], v)} min={0} step={0.5} />
          <NumberInput label="Max Active Bundles" value={tr.maxActiveBundles as number} onChange={(v) => update(["trading", "maxActiveBundles"], v)} min={1} />
          <NumberInput label="Default Min ROI (%)" value={tr.defaultMinRoi as number} onChange={(v) => update(["trading", "defaultMinRoi"], v)} min={0} step={0.5} />
          <SelectInput label="Log Level" value={tr.logLevel as string} options={["debug", "info", "warn", "error"]} onChange={(v) => update(["trading", "logLevel"], v)} />
          <NumberInput label="Metrics Port" value={tr.metricsPort as number} onChange={(v) => update(["trading", "metricsPort"], v)} min={1024} max={65535} />
        </div>
        <div className="flex flex-wrap gap-6 mt-4 pt-4 border-t border-border">
          <Toggle label="Auto-match approved pairs" value={tr.autoMatchApproved as boolean} onChange={(v) => update(["trading", "autoMatchApproved"], v)} />
          <Toggle label="Email notifications" value={tr.emailNotifications as boolean} onChange={(v) => update(["trading", "emailNotifications"], v)} />
          <Toggle label="Paper mode (simulate all trades)" value={tr.paperMode as boolean} onChange={(v) => update(["trading", "paperMode"], v)} />
        </div>
      </Section>

      <Section title="AI Pair Review" subtitle="Configure models, batching, and review prompts">
        <div className="mb-5 pb-5 border-b border-border">
          <div className="flex items-end gap-3">
            <div className="flex-1">
              <TextInput
                label="OpenRouter API Key"
                value={(settings.openrouterApiKey as string) ?? ""}
                onChange={(v) => update(["openrouterApiKey"], v)}
              />
            </div>
            <button
              type="button"
              onClick={fetchModels}
              disabled={!settings.openrouterApiKey || loadingModels}
              className="shrink-0 px-3 py-1.5 text-xs rounded-lg bg-surface border border-border text-muted hover:text-gray-200 hover:border-accent transition-colors disabled:opacity-40"
            >
              {loadingModels ? "Loading..." : "Refresh"}
            </button>
          </div>
          <p className="text-xs text-muted mt-1">Key is stored server-side. Enter and click Refresh to load available models.</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-5 pb-5 border-b border-border">
          <div className="space-y-3">
            <div className="flex items-center gap-2 mb-2">
              <span className="w-2 h-2 rounded-full bg-accent" />
              <span className="text-sm font-semibold text-gray-200">Leg A — Batch Screener</span>
              <span className="text-[10px] text-muted bg-surface-hover px-1.5 py-0.5 rounded">First Pass</span>
            </div>
            <ModelSelect
              value={pr.legAModel as string}
              onChange={(v) => update(["pairReview", "legAModel"], v)}
              options={modelOptions}
              loading={loadingModels}
              error={modelError}
              hasKey={!!settings.openrouterApiKey}
            />
            <p className="text-xs text-muted">Receives a batch of candidate pairs and determines which describe the same real-world event.</p>
          </div>
          <div className="space-y-3">
            <div className="flex items-center gap-2 mb-2">
              <span className="w-2 h-2 rounded-full bg-amber" />
              <span className="text-sm font-semibold text-gray-200">Leg B — Second Opinion</span>
              <span className="text-[10px] text-muted bg-surface-hover px-1.5 py-0.5 rounded">Confidence Check</span>
            </div>
            <ModelSelect
              value={pr.legBModel as string}
              onChange={(v) => update(["pairReview", "legBModel"], v)}
              options={modelOptions}
              loading={loadingModels}
              error={modelError}
              hasKey={!!settings.openrouterApiKey}
            />
            <p className="text-xs text-muted">Cross-checks confirmed matches from Leg A with a different model for verification.</p>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <NumberInput label="Batch Size (Leg A)" value={pr.batchSize as number} onChange={(v) => update(["pairReview", "batchSize"], v)} min={1} max={200} help="Number of candidate pairs sent to the Leg A model per batch." />
          <NumberInput label="Confidence Threshold" value={Math.round((pr.confidenceThreshold as number) * 100)} onChange={(v) => update(["pairReview", "confidenceThreshold"], v / 100)} min={0} max={100} suffix="%" help="Pairs scoring above this after Leg B are auto-approved." />
          <Toggle label="Auto-approve above threshold" value={pr.autoApproveAboveThreshold as boolean} onChange={(v) => update(["pairReview", "autoApproveAboveThreshold"], v)} />
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

function ModelSelect({ value, onChange, options, loading, error, hasKey }: {
  value: string; onChange: (v: string) => void; options: { value: string; label: string }[]; loading: boolean; error: string; hasKey: boolean
}) {
  if (!hasKey) {
    return (
      <div className="w-full bg-surface border border-border rounded-lg px-3 py-2 text-sm text-muted italic">
        Enter OpenRouter API key above and click Refresh
      </div>
    )
  }
  if (loading) {
    return (
      <div className="w-full bg-surface border border-border rounded-lg px-3 py-2 text-sm text-muted">
        Loading models...
      </div>
    )
  }
  if (error) {
    return (
      <div className="w-full bg-surface border border-red-500/40 rounded-lg px-3 py-2 text-sm text-red-400">
        {error}
      </div>
    )
  }
  if (!options.length) {
    return (
      <div className="w-full bg-surface border border-border rounded-lg px-3 py-2 text-sm text-muted italic">
        No models loaded
      </div>
    )
  }

  return (
    <select
      className="w-full bg-surface border border-border rounded-lg px-3 py-2 text-sm text-gray-200 outline-none focus:border-accent"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    >
      <option value="">Select a model...</option>
      {options.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
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
