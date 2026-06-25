import requests

# Try to fetch this specific market
ticker = "KXMLBGAME-26JUN251945AZSTL-STL"
resp = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?ticker={ticker}")
print(f"Status: {resp.status_code}")
data = resp.json()
markets = data.get("markets", [])
print(f"Markets: {len(markets)}")
for m in markets:
    print(f"  Ticker: {m['ticker']}")
    print(f"  Title: {m['title'][:80]}")
    print(f"  Event ticker: {m.get('event_ticker')}")
    print(f"  Market type: {m.get('market_type')}")
    print(f"  mve_collection: {m.get('mve_collection_ticker', '(none)')}")
    print(f"  Rules: {m.get('rules_primary', '')[:200]}")
    print(f"  Status: {m.get('status')}")

# Also try fetching the event
event_ticker = "KXMLBGAME-26JUN251945AZSTL"
resp2 = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/events?event_ticker={event_ticker}")
print(f"\nEvent query status: {resp2.status_code}")
e2 = resp2.json()
for e in e2.get("events", []):
    print(f"  Event ticker: {e.get('event_ticker')}")
    print(f"  Title: {e.get('title')}")
    print(f"  Category: {e.get('category')}")

# Also try markets by event ticker
resp3 = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?event_ticker={event_ticker}&limit=10")
print(f"\nMarkets for event: {resp3.status_code}")
m3 = resp3.json()
for m in m3.get("markets", []):
    print(f"  {m['ticker']} | {m['title'][:60]} | mve={m.get('mve_collection_ticker','(none)')}")
