SELECT 'moneyline' as type, COUNT(*) as cnt FROM markets WHERE title ~* '\y(win|wins|to win|to beat|winner|champion|championship)\y'
UNION ALL
SELECT 'over_under', COUNT(*) FROM markets WHERE title ~* '\y(over\s+\d+\.?\d*|under\s+\d+\.?\d*)\y'
UNION ALL
SELECT 'spread', COUNT(*) FROM markets WHERE title ~* '\y(spread|cover|handicap|-\d+\.?\d*|plus\s+\d+\.?\d*)\y'
UNION ALL
SELECT 'player_prop', COUNT(*) FROM markets WHERE title ~* '\y(player|points|assists|rebounds|touchdown|goals? scored|strikeout|hit|yard|ace)\y'
UNION ALL
SELECT 'team_prop', COUNT(*) FROM markets WHERE title ~* '\y(both teams to score|first (team|to|goal)|last (team|to)|race to|will there be|total (goals?|runs?|points?))\y'
ORDER BY type
LIMIT 20;
SELECT COUNT(*) as total_markets FROM markets;
