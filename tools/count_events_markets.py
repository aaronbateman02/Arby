import requests

url = "https://api.elections.kalshi.com/trade-api/v2/events?limit=1000&status=open"
total_events = 0
total_markets = 0
while url:
    resp = requests.get(url)
    data = resp.json()
    events = data.get("events", [])
    total_events += len(events)
    for e in events:
        et = e.get("event_ticker", "")
        mresp = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?event_ticker={et}&limit=100")
        mdata = mresp.json()
        n_markets = len(mdata.get("markets", []))
        total_markets += n_markets
    print(f"Processed {total_events} events, {total_markets} markets so far...")
    cursor = data.get("cursor", "")
    url = f"https://api.elections.kalshi.com/trade-api/v2/events?limit=1000&cursor={cursor}&status=open" if cursor else ""

print(f"\nTotal events: {total_events}")
print(f"Total markets: {total_markets}")
