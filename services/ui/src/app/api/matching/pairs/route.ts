import { NextRequest, NextResponse } from "next/server"

export const dynamic = "force-dynamic"

const GO_API_URL = process.env.GO_API_URL || "http://polybot:8086"

export async function GET(req: NextRequest) {
  const params = req.nextUrl.searchParams.toString()
  const url = `${GO_API_URL}/api/v1/matching/pairs${params ? `?${params}` : ""}`
  const res = await fetch(url, { cache: "no-store" })
  const data = await res.json()
  return NextResponse.json(data)
}
