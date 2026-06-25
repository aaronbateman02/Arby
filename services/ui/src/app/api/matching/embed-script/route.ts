import { NextResponse } from "next/server"

const GO_API_URL = process.env.GO_API_URL || "http://polybot:8086"

export async function GET() {
  const res = await fetch(`${GO_API_URL}/api/v1/matching/embed-script`)
  const headers = new Headers()
  headers.set("Content-Type", res.headers.get("Content-Type") || "text/x-python")
  headers.set("Content-Disposition", res.headers.get("Content-Disposition") || "attachment; filename=embed_worker.py")
  return new NextResponse(res.body, { status: res.status, headers })
}
