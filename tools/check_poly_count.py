import subprocess, json

r = subprocess.run(
    ["curl", "-s", "https://clob.polymarket.com/markets?limit=1"],
    capture_output=True, text=True
)
data = json.loads(r.stdout)
print("count:", data.get("count"))
print("next_cursor:", data.get("next_cursor"))

# Get second page
cursor = data.get("next_cursor")
if cursor:
    r2 = subprocess.run(
        ["curl", "-s", f"https://clob.polymarket.com/markets?limit=1&next_cursor={cursor}"],
        capture_output=True, text=True
    )
    d2 = json.loads(r2.stdout)
    print("page2 id:", d2["data"][0]["condition_id"] if d2.get("data") else "none")
