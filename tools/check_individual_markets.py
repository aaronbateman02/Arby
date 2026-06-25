import requests
# Fetch individual markets from the bundle legs
tickers = [
    "KXWCGAME-26JUN25PARAUS-TIE",
    "KXWCGAME-26JUN26TUNNED-NED",
    "KXWCGAME-26JUN26NZLBEL-BEL",
    "KXWCGAME-26JUN27PANENG-ENG"
]
for t in tickers:
    resp = requests.get(f"https://api.elections.kalshi.com/trade-api/v2/markets?ticker={t}")
    data = resp.json()
    m = data.get("markets", [{}])[0] if data.get("markets") else {}
    print(f"\nTicker: {t}")
    print(f"  Title: {m.get('title', '(missing)')}")
    print(f"  Market type: {m.get('market_type', '(missing)')}")
    print(f"  Event ticker: {m.get('event_ticker', '(missing)')}")
    print(f"  Rules: {str(m.get('rules_primary', ''))[:200]}")
