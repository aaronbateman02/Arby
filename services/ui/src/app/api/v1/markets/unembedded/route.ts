import { NextRequest, NextResponse } from "next/server"

export const dynamic = "force-dynamic"

const GO_API_URL = process.env.GO_API_URL || "http://polybot:8086"

export async function GET(req: NextRequest) {
  const limit = req.nextUrl.searchParams.get("limit") || "64"
  const res = await fetch(`${GO_API_URL}/api/v1/markets/unembedded?limit=${limit}`, { cache: "no-store" })
  const data = await res.json()
  return NextResponse.json(data)
}
