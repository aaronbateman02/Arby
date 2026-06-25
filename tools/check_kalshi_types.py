import requests
# Check what market_types Kalshi has
resp = requests.get("https://api.elections.kalshi.com/trade-api/v2/markets?limit=100&status=open")
data = resp.json()
types = {}
for m in data.get("markets", []):
    mt = m.get("market_type", "unknown")
    ticker = m.get("ticker", "")
    if "MVE@" in ticker or "MVE" in ticker:
        key = f"MVE_prefix"
    else:
        key = ticker[:15] if ticker else "unknown"
    types[key] = types.get(key, 0) + 1

print("Market type distribution (sample of 100):")
for k, v in sorted(types.items(), key=lambda x: -x[1]):
    print(f"  {k}: {v}")

# Also check event_ticker presence
events_with = sum(1 for m in data.get("markets", []) if m.get("event_ticker"))
print(f"\nMarkets with event_ticker: {events_with}/{len(data['markets'])}")
