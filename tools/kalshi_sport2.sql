SELECT
  SUM(CASE WHEN title ~* '\y(NBA|basketball)\y' THEN 1 ELSE 0 END) as nba,
  SUM(CASE WHEN title ~* '\y(NFL|football|super bowl)\y' THEN 1 ELSE 0 END) as nfl,
  SUM(CASE WHEN title ~* '\y(MLB|baseball|world series)\y' THEN 1 ELSE 0 END) as mlb,
  SUM(CASE WHEN title ~* '\y(NHL|hockey|stanley cup)\y' THEN 1 ELSE 0 END) as nhl,
  SUM(CASE WHEN title ~* '\y(soccer|premier league|champions league|la liga|bundesliga|mls|liga mx)\y' THEN 1 ELSE 0 END) as soccer,
  SUM(CASE WHEN title ~* '\y(tennis|atp|wta)\y' THEN 1 ELSE 0 END) as tennis,
  SUM(CASE WHEN title ~* '\y(golf|pga|masters)\y' THEN 1 ELSE 0 END) as golf,
  SUM(CASE WHEN title ~* '\y(ufc|mma)\y' THEN 1 ELSE 0 END) as ufc,
  SUM(CASE WHEN title ~* '\y(runs scored|runs|inning|strikeout|home run|hits)\y' THEN 1 ELSE 0 END) as baseball_terms,
  SUM(CASE WHEN title ~* '\y(goal|goals|both teams to score|btst|corner|yellow card|red card|free kick|penalty|offside)\y' THEN 1 ELSE 0 END) as soccer_terms,
  SUM(CASE WHEN title ~* '\y(point|points|quarter|free throw|three pointer|rebound|assist|turnover)\y' THEN 1 ELSE 0 END) as basketball_terms,
  SUM(CASE WHEN title ~* '\y(puck|period|faceoff|shot|save)\y' THEN 1 ELSE 0 END) as hockey_terms
FROM markets WHERE venue = 'KALSHI' AND venue_market_id LIKE 'KXMVESPORT%';
