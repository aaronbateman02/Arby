import requests

# Fetch markets for the event directly
resp = requests.get("https://api.elections.kalshi.com/trade-api/v2/events?event_ticker=KXWCGAME-26JUN25PARAUS")
data = resp.json()
print("Event:", data.get("events", [{}])[0].get("title"))

# Try fetching the event's markets with different approach
resp2 = requests.get("https://api.elections.kalshi.com/trade-api/v2/markets?limit=100&status=open")
markets = resp2.json().get("markets", [])

# Find all non-KXMV tickers
non_bundle = [m for m in markets if not m["ticker"].startswith("KXMV")]
print(f"\nNon-bundle markets in sample of 100: {len(non_bundle)}")
for m in non_bundle:
    print(f"  {m['ticker']} | {m['title'][:60]} | type={m['market_type']}")

# Also search by event_ticker
resp3 = requests.get("https://api.elections.kalshi.com/trade-api/v2/markets?event_ticker=KXWCGAME-26JUN25PARAUS&limit=10")
print(f"\nMarkets for event KXWCGAME-26JUN25PARAUS: {len(resp3.json().get('markets', []))}")
for m in resp3.json().get("markets", [])[:5]:
    print(f"  {m['ticker']} | {m['title'][:80]} | mve={m.get('mve_collection_ticker','(none)')}")
