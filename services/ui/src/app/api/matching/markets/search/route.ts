import { NextRequest, NextResponse } from "next/server"

const GO_API_URL = process.env.GO_API_URL || "http://polybot:8086"

export async function GET(req: NextRequest) {
  const venue = req.nextUrl.searchParams.get("venue")
  const q = req.nextUrl.searchParams.get("q")
  if (!venue || !q) {
    return NextResponse.json({ markets: [] })
  }
  const res = await fetch(
    `${GO_API_URL}/api/v1/matching/markets/search?venue=${encodeURIComponent(venue)}&q=${encodeURIComponent(q)}&limit=10`,
  )
  const data = await res.json()
  return NextResponse.json(data)
}
