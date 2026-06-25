SELECT
  SUM(CASE WHEN title ILIKE '%nba%' OR title ILIKE '%basketball%' THEN 1 ELSE 0 END) as nba,
  SUM(CASE WHEN title ILIKE '%nfl%' OR title ILIKE '%football%' OR title ILIKE '%super bowl%' THEN 1 ELSE 0 END) as nfl,
  SUM(CASE WHEN title ILIKE '%mlb%' OR title ILIKE '%baseball%' OR title ILIKE '%world series%' THEN 1 ELSE 0 END) as mlb,
  SUM(CASE WHEN title ILIKE '%nhl%' OR title ILIKE '%hockey%' OR title ILIKE '%stanley cup%' THEN 1 ELSE 0 END) as nhl,
  SUM(CASE WHEN title ILIKE '%soccer%' OR title ILIKE '%premier league%' OR title ILIKE '%champions league%' OR title ILIKE '%la liga%' OR title ILIKE '%bundesliga%' OR title ILIKE '%liga mx%' OR title ILIKE '%mls%' THEN 1 ELSE 0 END) as soccer,
  SUM(CASE WHEN title ILIKE '%tennis%' OR title ILIKE '%atp%' OR title ILIKE '%wta%' THEN 1 ELSE 0 END) as tennis,
  SUM(CASE WHEN title ILIKE '%golf%' OR title ILIKE '%pga%' OR title ILIKE '%masters%' THEN 1 ELSE 0 END) as golf,
  SUM(CASE WHEN title ILIKE '%ufc%' OR title ILIKE '%mma%' THEN 1 ELSE 0 END) as ufc
FROM markets WHERE venue = 'KALSHI' AND venue_market_id LIKE 'KXMVESPORT%';
