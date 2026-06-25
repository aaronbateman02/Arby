import subprocess, json
result = subprocess.run(
    ["docker", "exec", "arby-postgres", "psql", "-U", "arby", "-d", "arby", "-c",
     "SELECT venue, COUNT(*) as cnt FROM markets WHERE embedding IS NOT NULL GROUP BY venue ORDER BY cnt DESC"],
    capture_output=True, text=True, cwd="/home/ec2-user/Arby"
)
print(result.stdout)
print(result.stderr)
