"use client"

import { useCallback, useEffect, useState } from "react"

interface Settings {
  similarity_threshold: number
}

export default function MatchingSettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null)
  const [threshold, setThreshold] = useState(0.8)
  const [initialThreshold, setInitialThreshold] = useState(0.8)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")
  const [saved, setSaved] = useState(false)
  const [loadError, setLoadError] = useState("")

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        setLoadError("")
        const res = await fetch("/api/matching/settings")
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data: Settings = await res.json()
        setSettings(data)
        setThreshold(data.similarity_threshold)
        setInitialThreshold(data.similarity_threshold)
      } catch (e) {
        setLoadError(e instanceof Error ? e.message : "Failed to load settings")
      } finally {
        setLoading(false)
      }
    }
    fetchSettings()
  }, [])

  const handleSave = useCallback(async () => {
    setSaving(true)
    setError("")
    setSaved(false)
    try {
      const res = await fetch("/api/matching/settings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ similarity_threshold: threshold }),
      })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      setInitialThreshold(threshold)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save settings")
    } finally {
      setSaving(false)
    }
  }, [threshold])

  const changed = threshold !== initialThreshold

  if (loading) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <h1 className="text-2xl font-bold text-gray-100">Matching Settings</h1>
        <p className="text-sm text-muted mt-1">Loading settings…</p>
      </div>
    )
  }

  if (loadError) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <h1 className="text-2xl font-bold text-gray-100">Matching Settings</h1>
        <p className="text-sm text-red mt-1">{loadError}</p>
      </div>
    )
  }

  const handleDownload = async () => {
    const res = await fetch("/api/matching/embed-script")
    if (!res.ok) return
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = "embed_worker.py"
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-100">Matching Settings</h1>
        <p className="text-sm text-muted mt-1">Configure matching pipeline behavior</p>
      </div>

      <div className="bg-surface-alt rounded-xl border border-border p-6 space-y-4">
        <div>
          <h2 className="text-lg font-semibold text-gray-100">Similarity Threshold</h2>
          <p className="text-sm text-muted mt-1">
            Markets with similarity below this threshold will not be sent for LLM comparison. Changes take effect on the next discovery cycle.
          </p>
        </div>

        <div className="flex items-center gap-4">
          <input
            type="range"
            min={0.5}
            max={1.0}
            step={0.01}
            value={threshold}
            onChange={(e) => setThreshold(Number(e.target.value))}
            className="flex-1 accent-blue-500"
          />
          <span className="text-sm font-mono text-gray-100 w-12 text-right">{Math.round(threshold * 100)}%</span>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={handleSave}
            disabled={!changed || saving}
            className="px-4 py-2 text-sm font-medium rounded-lg bg-blue-600 text-white hover:bg-blue-500 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            {saving ? "Saving…" : "Save"}
          </button>
          {saved && <span className="text-sm text-green">Saved</span>}
          {error && <span className="text-sm text-red">{error}</span>}
        </div>
      </div>

      <div className="bg-surface-alt rounded-xl border border-border p-6 space-y-4">
        <div>
          <h2 className="text-lg font-semibold text-gray-100">Embed Worker</h2>
          <p className="text-sm text-muted mt-1">
            Download the standalone embedding script to run on your local machine. It uses BAAI/bge-large-en-v1.5 via sentence-transformers.
          </p>
        </div>

        <div className="bg-black/20 rounded-lg px-4 py-3">
          <code className="text-sm text-gray-100">python embed_worker.py --host https://arby.nostrabotus.com</code>
        </div>

        <button
          onClick={handleDownload}
          className="px-4 py-2 text-sm font-medium rounded-lg bg-surface-alt border border-border text-gray-100 hover:bg-border transition-colors"
        >
          Download embed_worker.py
        </button>
      </div>
    </div>
  )
}
