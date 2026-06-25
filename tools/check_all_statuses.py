import requests
# Check closed/settled markets too
for status in ["open", "closed", "settled", "expired"]:
    resp = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?limit=5&status={status}")
    data = resp.json()
    markets = data.get("markets", [])
    print(f"{status}: {len(markets)} markets returned")
    if markets:
        m = markets[0]
        print(f"  Sample: {m.get('ticker')} | {m.get('title')[:60]} | type={m.get('market_type')}")
