import requests

# Check events endpoint pagination
resp = requests.get("https://api.elections.kalshi.com/trade-api/v2/events?limit=50&status=open")
data = resp.json()
events = data.get("events", [])
print(f"Events returned: {len(events)}")
print(f"Cursor: {data.get('cursor', '(none)')}")

if events:
    # Check categories
    cats = {}
    for e in events:
        c = e.get("category", "unknown")
        cats[c] = cats.get(c, 0) + 1
    print(f"\nCategories:")
    for c, n in sorted(cats.items(), key=lambda x: -x[1]):
        print(f"  {c}: {n}")
