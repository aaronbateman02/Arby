-- V35: Add market_type column and backfill with title-based classification.
-- Mirrors pkg/matching/categorize.go CategorizeMarketType function.

ALTER TABLE markets ADD COLUMN IF NOT EXISTS market_type VARCHAR(50);

UPDATE markets
SET market_type = CASE
    WHEN title ~* '\y(over\s+\d+\.?\d*|under\s+\d+\.?\d*|o\s*\d+\.?\d*|u\s*\d+\.?\d*|more than\s*\d+|less than\s*\d+|at least\s*\d+|at most\s*\d+)\y' THEN 'over_under'
    WHEN title ~* '\y(spread|cover|handicap|-\d+\.?\d*\s*(points?|goals?|runs?))\y' THEN 'spread'
    WHEN title ~* '\y(champion|championship|win (the |their )?(div|conf|league|title|series|cup|super bowl|world series|stanley cup|nba finals)|to win it all|will win (the )?(nba|nfl|mlb|nhl|world cup|super bowl|championship|title|trophy)|future|season (win|total|proposition)|finish (first|top|in the)|award|mvp|rookie of the|most valuable|bettor of the|nfl draft)\y' THEN 'future'
    WHEN title ~* '\y(player|points?|assists?|rebounds?|strikeouts?|touchdowns?|yards?|goals?\s+scored|ace|hits?|rbi|walks?|sacks?|interceptions?|three pointers?|field goals?\s+made|free throws?\s+made|turnovers?|blocks?|steals?|passing|rushing|receiving)\y' THEN 'player_prop'
    WHEN title ~* '\y([A-Z][a-z]+\s+[A-Z][a-z]+:\s*\d+\+)\y' THEN 'player_prop'
    WHEN title ~* '\y(both teams to score|btst|team to score first|first (team|to score|goal|touchdown|run|point)|last team|race to|will there be a|total (goals?|runs?|points?) (in|by|over)|team total|highest (scoring|paid)|will (either|both)|ends in|draw|tie)\y' THEN 'team_prop'
    WHEN title ~* '\y(to win|wins?|beat|defeat|outright|winner|ml|moneyline|1x2|double chance)\y' THEN 'moneyline'
    WHEN title ~* '\y([A-Z][a-zA-Z]+(\s+[A-Z][a-zA-Z]+)*\s+vs\.?\s+[A-Z][a-zA-Z]+(\s+[A-Z][a-zA-Z]+)*)\y' THEN 'moneyline'
    WHEN title ~* '\y(over\/under|o\/u|over_under|total (goals?|runs?|points?))\y' THEN 'over_under'
END
WHERE market_type IS NULL OR market_type = '';
