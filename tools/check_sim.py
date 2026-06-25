import subprocess

# Get one Kalshi market with embedding
sql1 = "SELECT id FROM markets WHERE venue = 'KALSHI' AND embedding IS NOT NULL LIMIT 1;"
r1 = subprocess.run(
    ["docker", "exec", "arby-postgres", "psql", "-U", "arby", "-d", "arby", "-t", "-A", "-c", sql1],
    capture_output=True, text=True, cwd="/home/ec2-user/Arby"
)
mid = r1.stdout.strip()
print(f"Market: {mid}")

# Get top 10 cross-venue similar
sql2 = f"SELECT m2.id, m2.venue, ROUND((1 - (m1.embedding <=> m2.embedding))::numeric, 4) AS sim FROM markets m1, markets m2 WHERE m1.id = '{mid}' AND m2.venue = 'POLYMARKET' AND m2.embedding IS NOT NULL AND m2.id != m1.id ORDER BY sim DESC LIMIT 10;"
r2 = subprocess.run(
    ["docker", "exec", "arby-postgres", "psql", "-U", "arby", "-d", "arby", "-t", "-A", "-c", sql2],
    capture_output=True, text=True, cwd="/home/ec2-user/Arby"
)
print(r2.stdout)
if r2.stderr:
    print("ERR:", r2.stderr)
