-- V34: Title-based subcategory fallback for Kalshi MVESPORT/MVECROSS markets.
-- These tickers don't encode sport type, so we infer from title keywords.

UPDATE markets
SET subcategory = CASE
    WHEN title ~* '\y(goal|goals|both teams to score|btst|shot on target|corner|yellow card|red card|offside|free kick|penalty|world cup|champions league|premier league|la liga|bundesliga|serie a|liga mx|epl|uefa|fifa)\y' THEN 'soccer'
    WHEN title ~* '\y(runs? scored|inning|strikeout|home run|hr|hits|rbi|era|batting|pitch|pitcher)\y' THEN 'mlb'
    WHEN title ~* '\y(point|quarter|free throw|three pointer|three point|rebound|assist|turnover|foul|nba|basketball)\y' THEN 'nba'
    WHEN title ~* '\y(nfl|touchdown|field goal|passing yards|rushing yards|reception|interception|sack|punt|kickoff)\y' THEN 'nfl'
    WHEN title ~* '\y(tennis|atp|wta|grand slam|wimbledon|us open|french open|australian open|set|ace|break point)\y' THEN 'tennis'
    WHEN title ~* '\y(ufc|mma|knockout|submission|octagon|round)\y' THEN 'ufc'
    WHEN title ~* '\y(golf|pga|masters|pga tour|birdie|eagle|par|putt|fairway|green)\y' THEN 'golf'
END
WHERE venue = 'KALSHI' AND (subcategory IS NULL OR subcategory = '');
