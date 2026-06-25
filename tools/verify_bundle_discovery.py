import requests

# This event is NOT in the /events listing but has sub-markets
ticker = "KXMLBGAME-26JUN251945AZSTL"
resp = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?event_ticker={ticker}&limit=10")
markets = resp.json().get("markets", [])
print(f"Markets for {ticker}: {len(markets)}")
for m in markets:
    print(f"  {m['ticker']} | {m['title']} | mve={m.get('mve_collection_ticker','(none)')} | event={m.get('event_ticker')}")

# Check one bundle's mve_selected_legs
bundle_ticker = "KXMVESPORTSMULTIGAMEEXTENDED-S202670CB0972D3C-4614B8AEA4B"  # has St. Louis
resp2 = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?ticker={bundle_ticker}")
bundle = resp2.json().get("markets", [{}])[0]
legs = bundle.get("mve_selected_legs", [])
print(f"\nBundle {bundle_ticker[:40]}... has {len(legs)} legs:")
for leg in legs[:5]:
    print(f"  event={leg['event_ticker']}, market={leg['market_ticker']}, side={leg['side']}")
    # Fetch this individual market
    m3 = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?ticker={leg['market_ticker']}")
    m3m = m3.json().get("markets", [{}])[0]
    print(f"    -> Title: {m3m.get('title','?')[:60]}")
