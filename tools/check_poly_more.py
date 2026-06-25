import subprocess, json

endpoints = [
    "https://clob.polymarket.com/markets?limit=1",
    "https://clob.polymarket.com/markets?limit=1&closed=true",
    "https://clob.polymarket.com/markets?limit=1&archived=true",
]
for url in endpoints:
    r = subprocess.run(["curl", "-s", url], capture_output=True, text=True)
    data = json.loads(r.stdout)
    print(f"{url.split('?')[1]:35s} count={data.get('count','?')} has_data={bool(data.get('data'))}")

# Check tags from first market
r = subprocess.run(
    ["curl", "-s", "https://clob.polymarket.com/markets?limit=1"],
    capture_output=True, text=True
)
data = json.loads(r.stdout)
m = data.get("data", [{}])[0]
print(f"\ntags sample: {m.get('tags', [])[:5]}")
print(f"tag types: {[type(t).__name__ for t in m.get('tags', [])]}")
