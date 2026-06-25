import subprocess
sql = "SELECT DISTINCT category FROM markets WHERE venue = 'KALSHI' AND category IS NOT NULL AND category != '' ORDER BY category;"
r = subprocess.run(["docker","exec","arby-postgres","psql","-U","arby","-d","arby","-c",sql], capture_output=True, text=True, cwd="/home/ec2-user/Arby")
print(r.stdout)
