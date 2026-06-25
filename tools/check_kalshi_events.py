import requests
# Check Kalshi events endpoint
resp = requests.get("https://api.elections.kalshi.com/trade-api/v2/events?limit=3&status=open")
print("/events status:", resp.status_code)
if resp.status_code == 200:
    data = resp.json()
    for e in data.get("events", [])[:3]:
        print("Keys:", list(e.keys()))
        print("Event ticker:", e.get("event_ticker"))
        print("Title:", e.get("title"))
        print("Description:", e.get("description", "(missing)"))

# Also check if there's a series endpoint
resp2 = requests.get("https://api.elections.kalshi.com/trade-api/v2/series?limit=3")
print("\n/series status:", resp2.status_code)
if resp2.status_code == 200:
    data2 = resp2.json()
    for s in data2.get("series", [])[:3]:
        print("Keys:", list(s.keys()))
        print("Series ticker:", s.get("series_ticker"))
        print("Title:", s.get("title"))
        print("Description:", s.get("description", "(missing)"))
