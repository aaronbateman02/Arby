import requests
# Try the old system's endpoint
resp = requests.get("https://trading-api.kalshi.com/trade-api/v2/markets?limit=3&status=open")
print("Status:", resp.status_code)
if resp.status_code == 200:
    data = resp.json()
    print("Count:", len(data.get("markets", [])))
    for m in data.get("markets", [])[:3]:
        print(f"\nTicker: {m.get('ticker')}")
        print(f"  Title: {m.get('title')}")
        print(f"  Type: {m.get('market_type')}")
        print(f"  Description: {m.get('description', '(missing)')}")
        print(f"  Subtitle: {m.get('subtitle', '(missing)')}")
        print(f"  Event ticker: {m.get('event_ticker')}")
        print(f"  Rules: {m.get('rules_primary', '')[:100]}")
else:
    print("Response:", resp.text[:500])
