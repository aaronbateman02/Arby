import { NextRequest, NextResponse } from "next/server"

const GO_API_URL = process.env.GO_API_URL || "http://polybot:8086"

export async function GET() {
  const res = await fetch(`${GO_API_URL}/api/v1/matching/settings`)
  const data = await res.json()
  return NextResponse.json(data)
}

export async function POST(req: NextRequest) {
  const body = await req.json()
  const res = await fetch(`${GO_API_URL}/api/v1/matching/settings`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  const data = await res.json()
  return NextResponse.json(data)
}
