import subprocess
# Check all venues and their category values
sql = "SELECT venue, category, COUNT(*) as cnt FROM markets WHERE embedding IS NOT NULL GROUP BY venue, category ORDER BY cnt DESC LIMIT 20;"
r = subprocess.run(["docker","exec","arby-postgres","psql","-U","arby","-d","arby","-c",sql], capture_output=True, text=True, cwd="/home/ec2-user/Arby")
print("GROUP BY:", r.stdout)

# Check raw category strings
sql2 = "SELECT venue, category, char_length(category) as len FROM markets WHERE embedding IS NOT NULL AND venue='KALSHI' LIMIT 5;"
r2 = subprocess.run(["docker","exec","arby-postgres","psql","-U","arby","-d","arby","-c",sql2], capture_output=True, text=True, cwd="/home/ec2-user/Arby")
print("RAW:", r2.stdout)
