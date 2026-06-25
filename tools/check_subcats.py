import subprocess
sql = "SELECT venue, COALESCE(subcategory, 'none') as subcat, COUNT(*) as cnt FROM markets WHERE subcategory IS NOT NULL AND subcategory != '' GROUP BY venue, subcategory ORDER BY venue, cnt DESC;"
r = subprocess.run(["docker","exec","arby-postgres","psql","-U","arby","-d","arby","-c",sql], capture_output=True, text=True, cwd="/home/ec2-user/Arby")
print(r.stdout)
# Also check what doesn't have a subcategory
sql2 = "SELECT venue, COUNT(*) as cnt FROM markets WHERE subcategory IS NULL OR subcategory = '' GROUP BY venue;"
r2 = subprocess.run(["docker","exec","arby-postgres","psql","-U","arby","-d","arby","-c",sql2], capture_output=True, text=True, cwd="/home/ec2-user/Arby")
print("No subcategory:", r2.stdout)
