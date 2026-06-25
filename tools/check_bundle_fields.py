import requests
resp = requests.get("https://api.elections.kalshi.com/trade-api/v2/markets?limit=1&status=open")
m = resp.json()["markets"][0]
print("mve_selected_legs:", m.get("mve_selected_legs"))
print("mve_collection_ticker:", m.get("mve_collection_ticker"))
print("price_ranges:", str(m.get("price_ranges", ""))[:300])
