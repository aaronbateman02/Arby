import { NextResponse } from "next/server"

const GO_API_URL = process.env.GO_API_URL || "http://polybot:8086"

export async function GET() {
  const res = await fetch(`${GO_API_URL}/api/v1/matching/stats`)
  const data = await res.json()
  return NextResponse.json(data)
}
