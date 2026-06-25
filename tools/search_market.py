import subprocess
ticker = 'KXMLBGAME-26JUN251945AZSTL-STL'
cmd = "docker exec -i arby-postgres psql -U arby -d arby -c \"SELECT id, venue, venue_market_id, title, description, category, subcategory, market_type, yes_bid, yes_ask, no_bid, no_ask, close_time FROM markets WHERE venue_market_id ILIKE '%{ticker}%' OR venue_market_id ILIKE '%KXMLBGAME%';\"".format(ticker=ticker)
result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
print("STDOUT:", result.stdout)
if result.stderr:
    print("STDERR:", result.stderr[:500])
