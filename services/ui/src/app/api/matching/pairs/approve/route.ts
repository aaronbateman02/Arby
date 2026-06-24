import { NextRequest, NextResponse } from "next/server"

const GO_API_URL = process.env.GO_API_URL || "http://polybot:8086"

export async function POST(req: NextRequest) {
  const { id } = await req.json()
  const res = await fetch(`${GO_API_URL}/api/v1/matching/pairs/${encodeURIComponent(id)}/approve`, { method: "POST" })
  const data = await res.json()
  return NextResponse.json(data)
}
