import requests
# Check if events have individual sub-markets
resp = requests.get("https://api.elections.kalshi.com/trade-api/v2/events?limit=50&status=open")
data = resp.json()
print(f"Total events: {len(data.get('events',[]))}")
for e in data.get("events", [])[:5]:
    et = e.get("event_ticker", "")
    title = e.get("title", "")
    sub = e.get("sub_title", "")
    # Check if event has sub-markets
    mresp = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?event_ticker={et}&limit=5")
    mdata = mresp.json()
    markets = mdata.get("markets", [])
    print(f"\nEvent: {et}")
    print(f"  Title: {title}")
    print(f"  Subtitle: {sub}")
    print(f"  Markets: {len(markets)}")
    for m in markets[:3]:
        print(f"    Ticker: {m.get('ticker')}, Title: {m.get('title')[:60]}, Type: {m.get('market_type')}")
