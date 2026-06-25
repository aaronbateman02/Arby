import requests
# Fetch a known sports event
resp = requests.get("https://api.elections.kalshi.com/trade-api/v2/events?event_ticker=KXWCGAME-26JUN25PARAUS")
data = resp.json()
events = data.get("events", [])
print(f"Events: {len(events)}")
for e in events:
    print(f"  Ticker: {e.get('event_ticker')}")
    print(f"  Title: {e.get('title')}")
    print(f"  Subtitle: {e.get('sub_title')}")
    print(f"  Category: {e.get('category')}")
    print(f"  Series: {e.get('series_ticker')}")
    # Fetch markets for this event
    mresp = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?event_ticker=KXWCGAME-26JUN25PARAUS&limit=5")
    mdata = mresp.json()
    print(f"  Markets: {len(mdata.get('markets', []))}")
    for m in mdata.get("markets", [])[:3]:
        print(f"    Ticker: {m.get('ticker')}, Title: {m.get('title')[:60]}")
