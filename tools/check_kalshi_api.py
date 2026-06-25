import requests
resp = requests.get("https://api.elections.kalshi.com/trade-api/v2/markets?limit=3&status=open")
data = resp.json()
for m in data.get("markets", []):
    print("Keys:", list(m.keys()))
    print("Ticker:", m.get("ticker"))
    print("Title:", m.get("title"))
    print("Sector:", m.get("sector"))
    print("Series:", m.get("series"))
    print("Description:", m.get("description", "(missing)"))
    print("Subtitle:", m.get("subtitle", "(missing)"))
    print("---")
