import subprocess

sql = "SELECT COUNT(*) FROM match_candidates;"
result = subprocess.run(
    ["docker", "exec", "arby-postgres", "psql", "-U", "arby", "-d", "arby", "-t", "-A", "-c", sql],
    capture_output=True, text=True, cwd="/home/ec2-user/Arby"
)
print("candidates:", result.stdout.strip())
print("stderr:", result.stderr.strip())
