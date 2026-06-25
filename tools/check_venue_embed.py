import subprocess, json
sql = "SELECT venue, COUNT(*) as total, COUNT(*) FILTER (WHERE embedding IS NOT NULL) as embedded FROM markets GROUP BY venue;"
r = subprocess.run(["docker","exec","arby-postgres","psql","-U","arby","-d","arby","-c",sql], capture_output=True, text=True, cwd="/home/ec2-user/Arby")
print(r.stdout)
