import subprocess

sql = "SELECT venue, category, COUNT(*) as cnt FROM markets WHERE embedding IS NOT NULL AND category IS NOT NULL AND category != '' GROUP BY venue, category ORDER BY venue, cnt DESC;"
r = subprocess.run(
    ["docker", "exec", "arby-postgres", "psql", "-U", "arby", "-d", "arby", "-c", sql],
    capture_output=True, text=True, cwd="/home/ec2-user/Arby"
)
print(r.stdout)
