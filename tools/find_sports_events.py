import requests

# Fetch more events looking for sports
for offset in range(0, 500, 50):
    resp = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/events?limit=50&offset={offset}&status=open")
    data = resp.json()
    events = data.get("events", [])
    if not events:
        break
    for e in events:
        title = e.get("title", "")
        cat = e.get("category", "")
        et = e.get("event_ticker", "")
        # sports-related
        if any(kw in title.lower() for kw in ["game", "match", "win", "beat", "cup", "goal", "score", "nfl", "nba", "mlb", "nhl", "super bowl", "soccer", "football", "tennis", "basketball", "baseball", "hockey", "world cup", "champion", "final", "playoff", "stadium", "team", "player"]):
            mresp = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?event_ticker={et}&limit=5")
            mdata = mresp.json()
            markets = mdata.get("markets", [])
            for m in markets[:3]:
                print(f"Event: {et} | {title[:60]} | Market: {m.get('ticker')} | Type: {m.get('market_type')} | Title: {m.get('title')[:60]}")
