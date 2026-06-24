import { NextRequest, NextResponse } from "next/server"

const GO_API_URL = process.env.GO_API_URL || "http://polybot:8086"

export async function GET(req: NextRequest) {
  const status = req.nextUrl.searchParams.get("status") || ""
  const url = `${GO_API_URL}/api/v1/matching/pairs${status ? `?status=${status}` : ""}`
  const res = await fetch(url)
  const data = await res.json()
  return NextResponse.json(data)
}
