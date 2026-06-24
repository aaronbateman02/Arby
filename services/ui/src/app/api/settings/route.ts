import { NextRequest, NextResponse } from "next/server"
import { promises as fs } from "fs"
import path from "path"
import { appSettings } from "@/lib/data"

const SETTINGS_FILE = process.env.SETTINGS_FILE_PATH || "/data/settings/config.json"

type Settings = typeof appSettings

async function readFile(): Promise<Settings | null> {
  try {
    const raw = await fs.readFile(SETTINGS_FILE, "utf-8")
    return JSON.parse(raw)
  } catch {
    return null
  }
}

async function ensureDir() {
  await fs.mkdir(path.dirname(SETTINGS_FILE), { recursive: true })
}

export async function GET() {
  const saved = await readFile()
  const merged: Settings = saved
    ? { ...appSettings, ...saved, venueKeys: { ...appSettings.venueKeys, ...saved.venueKeys }, trading: { ...appSettings.trading, ...saved.trading }, pairReview: { ...appSettings.pairReview, ...saved.pairReview, typeSpecificPrompts: { ...appSettings.pairReview.typeSpecificPrompts, ...saved.pairReview?.typeSpecificPrompts } } }
    : appSettings
  return NextResponse.json(merged)
}

export async function POST(req: NextRequest) {
  try {
    const body: Settings = await req.json()
    await ensureDir()
    await fs.writeFile(SETTINGS_FILE, JSON.stringify(body, null, 2), "utf-8")
    return NextResponse.json({ ok: true })
  } catch (err) {
    return NextResponse.json({ ok: false, error: (err as Error).message }, { status: 500 })
  }
}
