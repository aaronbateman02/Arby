import requests

url = "https://api.elections.kalshi.com/trade-api/v2/markets?limit=1000&status=open"
total = 0
while url:
    resp = requests.get(url)
    data = resp.json()
    markets = data.get("markets", [])
    total += len(markets)
    cursor = data.get("cursor", "")
    url = f"https://api.elections.kalshi.com/trade-api/v2/markets?limit=1000&status=open&cursor={cursor}" if cursor else ""
    if total % 10000 == 0:
        print(f"Fetched {total}...")
print(f"Total Kalshi markets: {total}")
