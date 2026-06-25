import { NextResponse } from "next/server"

export const dynamic = "force-dynamic"

const GO_API_URL = process.env.GO_API_URL || "http://polybot:8086"

export async function GET() {
  const res = await fetch(`${GO_API_URL}/api/v1/matching/embed-script`, { cache: "no-store" })
  const text = await res.text()
  return new NextResponse(text, {
    status: res.status,
    headers: {
      "Content-Type": "text/x-python",
      "Content-Disposition": "attachment; filename=embed_worker.py",
    },
  })
}
