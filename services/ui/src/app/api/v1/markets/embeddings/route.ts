import { NextRequest, NextResponse } from "next/server"

export const dynamic = "force-dynamic"

const GO_API_URL = process.env.GO_API_URL || "http://polybot:8086"

export async function POST(req: NextRequest) {
  console.log("EMBEDDINGS PROXY CALLED", {
    method: req.method,
    url: req.url,
    headers: Object.fromEntries(req.headers),
    bodySize: parseInt(req.headers.get("content-length") || "0"),
  })
  const body = await req.json()
  const res = await fetch(`${GO_API_URL}/api/v1/markets/embeddings`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  })
  const data = await res.json()
  return NextResponse.json(data)
}
