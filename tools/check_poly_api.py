import subprocess, json

# Check API response structure
r = subprocess.run(
    ["curl", "-s", "https://clob.polymarket.com/markets?limit=3"],
    capture_output=True, text=True
)
data = json.loads(r.stdout)
print("type:", type(data).__name__)
if isinstance(data, dict):
    print("keys:", list(data.keys()))
    if "data" in data and isinstance(data["data"], list):
        print("data len:", len(data["data"]))
        if data["data"]:
            print("sample keys:", list(data["data"][0].keys()))
elif isinstance(data, list):
    print("len:", len(data))
    if data:
        print("sample keys:", list(data[0].keys()))
