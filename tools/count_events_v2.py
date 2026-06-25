import requests

url = "https://api.elections.kalshi.com/trade-api/v2/events?limit=50&status=open"
total_events = 0
total_markets = 0
cats = {}

while url:
    resp = requests.get(url)
    data = resp.json()
    events = data.get("events", [])
    if not events:
        break
    total_events += len(events)
    for e in events:
        c = e.get("category", "unknown")
        cats[c] = cats.get(c, 0) + 1
        et = e.get("event_ticker", "")
        mresp = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?event_ticker={et}&limit=100")
        mdata = mresp.json()
        total_markets += len(mdata.get("markets", []))
    print(f"  {total_events} events, {total_markets} markets")
    cursor = data.get("cursor", "")
    url = f"https://api.elections.kalshi.com/trade-api/v2/events?limit=50&cursor={cursor}&status=open" if cursor else ""

print(f"\nTotal events: {total_events}")
print(f"Total markets: {total_markets}")
print(f"\nCategories:")
for c, n in sorted(cats.items(), key=lambda x: -x[1]):
    print(f"  {c}: {n}")
