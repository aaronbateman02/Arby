package matching

import (
	"regexp"
	"strings"
)

type categoryRule struct {
	pattern  *regexp.Regexp
	category string
}

type subcategoryRule struct {
	pattern     *regexp.Regexp
	subcategory string
}

var kalshiRules []categoryRule
var polyRules []categoryRule
type marketTypeRule struct {
	pattern    *regexp.Regexp
	marketType string
}

var kalshiSubRules []subcategoryRule
var kalshiTitleSubRules []subcategoryRule
var polySubRules []subcategoryRule
var marketTypeRules []marketTypeRule

func init() {
	kalshiRules = []categoryRule{
		{regexp.MustCompile(`^KX(BTC|ETH|SOL|XRP|DOGE)[^D]*(15M|1H|2H|4H)`), "crypto_intraday"},
		{regexp.MustCompile(`^KX(BTCD|ETHD|ETHHD|SOLD|DOGE|XRP|BNB|BTC|ETH|SOL)`), "crypto_daily"},
		{regexp.MustCompile(`^KX(MLB|NBA|NHL|NFLGAME|NFLTOTAL|MLS|NWSL|AFL|EKSTRAKLASA|LIGAMX|KBO|ATPMATCH|WTA|UCLGAME|EPLGAME)`), "sports_spread"},
		{regexp.MustCompile(`^KX(NFLWINS|NFLDIV|NFLCHAMP|PGATOUR|PGATOP|PGAR|PGACHAMP|UFC|NASCAR|MVESPORTS|MVECROSS|MLBWINS|NBACHAMP|NHLCHAMP)`), "sports_team"},
		{regexp.MustCompile(`^KX(GOV|ELEC|PRES|SEC|SEN|HOUSE|PARTY|VOTE|POLL|CONG|DEMS|REP|MAGA|POTUS|SPOTIFYD)`), "politics"},
		{regexp.MustCompile(`^KX(WAR|NATO|UN|COUNTRY|GEOPO)`), "world_events"},
		{regexp.MustCompile(`^KX(GOLDD|GOLD|OIL|GAS|CPI|GDP|FED|RATE|JOB|UNEMPLOY|INFL)`), "economics"},
		{regexp.MustCompile(`^KX(SPACE|SPACEX|FDA|AI|TECH)`), "science_tech"},
		{regexp.MustCompile(`^KX(OSCAR|GRAMMY|EMMY|MOVIE|SHOW|TV)`), "entertainment"},
	}

	polyRules = []categoryRule{
		{regexp.MustCompile(`(?i)\b(bitcoin|btc|ethereum|eth|solana|sol|dogecoin|doge|ripple|xrp|binance|bnb|crypto|defi|nft|web3|altcoin|stablecoin|usdc|usdt)\b`), "crypto_daily"},
		{regexp.MustCompile(`(?i)\b(trump|biden|harris|democrat|republican|gop|congress|senate|house of representatives|president|presidency|election|ballot|vote|governor|mayor|cabinet|inauguration|impeach|potus|vp|vice president|midterm|primary|poll|approval rating|white house|supreme court|roe|abortion|gun control|immigration|border|tariff|nato|g7|g20|un security|sanctions|executive order)\b`), "politics"},
		{regexp.MustCompile(`(?i)\b(nfl|nba|mlb|nhl|mls|ncaa|college football|college basketball|epl|premier league|champions league|la liga|bundesliga|serie a|ligue 1|ufc|mma|tennis|atp|wta|pga tour)\b`), "sports_team"},
		{regexp.MustCompile(`(?i)\b(super bowl|world series|stanley cup|nba finals|nba champion|mls cup|pga|masters|us open|wimbledon|french open|australian open|ufc champion|formula 1|f1 champion|world cup|copa america|euro 2024|euro 2025|euro 2026|olympics|nascar)\b`), "sports_team"},
		{regexp.MustCompile(`(?i)\b(war|ceasefire|invasion|conflict|troops|military|nato|un resolution|sanctions|nuclear|missile|drone|attack|coup|protest|revolution|treaty|peace deal|ukraine|russia|israel|gaza|taiwan|china|north korea|iran|middle east|africa|europe|asia|latin america|refugee|humanitarian)\b`), "world_events"},
		{regexp.MustCompile(`(?i)\b(spacex|nasa|rocket|launch|satellite|mars|moon|asteroid|fda|drug approval|vaccine|clinical trial|cancer|alzheimer|ai|artificial intelligence|gpt|llm|chatgpt|openai|anthropic|google deepmind|autonomous|self-driving|quantum|fusion|climate change|carbon|emissions|renewable|solar|wind energy)\b`), "science_tech"},
		{regexp.MustCompile(`(?i)\b(gdp|cpi|inflation|interest rate|federal reserve|fed rate|unemployment|jobs report|recession|stock market|s&p 500|nasdaq|dow jones|oil price|gold price|commodities|trade deficit|budget|debt ceiling|ipo)\b`), "economics"},
		{regexp.MustCompile(`(?i)\b(oscar|grammy|emmy|golden globe|box office|movie|film|tv show|series|celebrity|award)\b`), "entertainment"},
		{regexp.MustCompile(`(?i)\b(climate|weather|hurricane|tornado|flood|drought|temperature|record heat|cold spell|blizzard)\b`), "climate"},
	}

	kalshiSubRules = []subcategoryRule{
		{regexp.MustCompile(`^KXNFL`), "nfl"},
		{regexp.MustCompile(`^KXNBA`), "nba"},
		{regexp.MustCompile(`^KXMLB`), "mlb"},
		{regexp.MustCompile(`^KXNHL`), "nhl"},
		{regexp.MustCompile(`^KXMLS|^KXNWSL`), "mls"},
		{regexp.MustCompile(`^KXUCLGAME|^KXEPLGAME`), "soccer"},
		{regexp.MustCompile(`^KXAFL`), "afl"},
		{regexp.MustCompile(`^KXPGATOUR|^KXPGATOP|^KXPGAR|^KXPGACHAMP`), "golf"},
		{regexp.MustCompile(`^KXUFC`), "ufc"},
		{regexp.MustCompile(`^KXNASCAR`), "nascar"},
		{regexp.MustCompile(`^KXATP|^KXWTA`), "tennis"},
		{regexp.MustCompile(`^KX(GOV|ELEC|PRES|SEC|SEN|HOUSE|PARTY|VOTE|POLL|CONG|DEMS|REP|MAGA|POTUS)`), "us_politics"},
		{regexp.MustCompile(`^KX(WAR|NATO|UN|COUNTRY|GEOPO)`), "geopolitics"},
		{regexp.MustCompile(`^KX(BTC|ETH|SOL|XRP|DOGE|BNB)`), "crypto"},
		{regexp.MustCompile(`^KX(GOLDD|GOLD|OIL|GAS|CPI|GDP|FED|RATE|JOB|UNEMPLOY|INFL)`), "macro"},
		{regexp.MustCompile(`^KX(SPACE|SPACEX|FDA|AI|TECH)`), "tech"},
		{regexp.MustCompile(`^KX(OSCAR|GRAMMY|EMMY|MOVIE|SHOW|TV)`), "culture"},
	}

	kalshiTitleSubRules = []subcategoryRule{
		// Title-based fallback for KXMVESPORT/KXMVECROSS and any other Kalshi tickers
		{regexp.MustCompile(`(?i)\b(goal|goals|both teams to score|btst|shot on target|corner|yellow card|red card|offside|free kick|penalty|world cup|champions league|premier league|la liga|bundesliga|serie a|liga mx|epl|uefa|fifa)\b`), "soccer"},
		{regexp.MustCompile(`(?i)\b(runs? scored|inning|strikeout|home run|hr|hits|rbi|era|batting|pitch|pitcher)\b`), "mlb"},
		{regexp.MustCompile(`(?i)\b(point|quarter|free throw|three pointer|three point|rebound|assist|turnover|foul|nba|basketball)\b`), "nba"},
		{regexp.MustCompile(`(?i)\b(nfl|touchdown|field goal|passing yards|rushing yards|reception|interception|sack|punt|kickoff)\b`), "nfl"},
		{regexp.MustCompile(`(?i)\b(tennis|atp|wta|grand slam|wimbledon|us open|french open|australian open|set|ace|break point)\b`), "tennis"},
		{regexp.MustCompile(`(?i)\b(ufc|mma|knockout|submission|octagon|round)\b`), "ufc"},
		{regexp.MustCompile(`(?i)\b(golf|pga|masters|pga tour|birdie|eagle|par|putt|fairway|green)\b`), "golf"},
	}

	polySubRules = []subcategoryRule{
		{regexp.MustCompile(`(?i)\b(nfl|super bowl)\b`), "nfl"},
		{regexp.MustCompile(`(?i)\b(nba|nba finals|nba champion)\b`), "nba"},
		{regexp.MustCompile(`(?i)\b(mlb|world series)\b`), "mlb"},
		{regexp.MustCompile(`(?i)\b(nhl|stanley cup)\b`), "nhl"},
		{regexp.MustCompile(`(?i)\b(mls|mls cup|premier league|champions league|la liga|bundesliga|serie a|ligue 1)\b`), "soccer"},
		{regexp.MustCompile(`(?i)\b(ufc|mma)\b`), "ufc"},
		{regexp.MustCompile(`(?i)\b(tennis|atp|wta|wimbledon|us open|french open|australian open)\b`), "tennis"},
		{regexp.MustCompile(`(?i)\b(golf|pga|masters|pga tour)\b`), "golf"},
		{regexp.MustCompile(`(?i)\b(nascar|formula 1|f1)\b`), "motorsports"},
		{regexp.MustCompile(`(?i)\b(trump|biden|harris|election|president|presidency|congress|senate|governor|vote|ballot|democrat|republican)\b`), "us_politics"},
		{regexp.MustCompile(`(?i)\b(ukraine|russia|israel|gaza|taiwan|china|north korea|iran|war|ceasefire|conflict|sanctions|nato)\b`), "geopolitics"},
		{regexp.MustCompile(`(?i)\b(bitcoin|btc|ethereum|eth|solana|crypto|defi)\b`), "crypto"},
		{regexp.MustCompile(`(?i)\b(inflation|interest rate|federal reserve|gdp|cpi|recession|stock market)\b`), "macro"},
		{regexp.MustCompile(`(?i)\b(spacex|nasa|ai|artificial intelligence|fda|vaccine|climate change)\b`), "tech"},
		{regexp.MustCompile(`(?i)\b(oscar|grammy|emmy|golden globe|box office|movie)\b`), "culture"},
	}

	marketTypeRules = []marketTypeRule{
		{regexp.MustCompile(`(?i)\b(over\s+\d+\.?\d*|under\s+\d+\.?\d*|o\s*\d+\.?\d*|u\s*\d+\.?\d*)\b`), "over_under"},
		{regexp.MustCompile(`(?i)\b(spread|cover|handicap|-\d+\.?\d*\s*(points?|goals?|runs?)?)\b`), "spread"},
		{regexp.MustCompile(`(?i)\b(champion|championship|win (the |their )?(div|conf|league|title|series|cup|super bowl|world series|stanley cup|nba finals)|to win it all|will win (the )?(nba|nfl|mlb|nhl|world cup|super bowl|championship|title|trophy)|future|season (win|total|proposition)|finish (first|top|in the)|award|mvp|rookie of the|most valuable|bettor of the|nfl draft)\b`), "future"},
		{regexp.MustCompile(`(?i)\b(player|points?|assists?|rebounds?|strikeouts?|touchdowns?|yards?|goals?\s+scored|ace|hits?|rbi|walks?|sacks?|interceptions?|three pointers?|field goals?\s+made|free throws?\s+made|turnovers?|blocks?|steals?|passing|rushing|receiving)\b`), "player_prop"},
		{regexp.MustCompile(`\b[A-Z][a-z]+ [A-Z][a-z]+:\s*\d+\+`), "player_prop"},
		{regexp.MustCompile(`(?i)\b(both teams to score|btst|team to score first|first (team|to score|goal|touchdown|run|point)|last team|race to|will there be a|total (goals?|runs?|points?) (in|by|over)|team total|highest (scoring|paid)|will (either|both)|ends in|draw|tie)\b`), "team_prop"},
		{regexp.MustCompile(`(?i)\b(to win|wins?|beat|defeat|outright|winner|ml|moneyline|1x2|double chance)\b`), "moneyline"},
		{regexp.MustCompile(`\b[A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)*\s+vs\.?\s+[A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)*\b`), "moneyline"},
		{regexp.MustCompile(`(?i)\b(over\/under|o\/u|over_under|total (goals?|runs?|points?))\b`), "over_under"},
	}
}

func CategorizeMarket(venue, ticker, title string) string {
	switch strings.ToUpper(venue) {
	case "KALSHI":
		for _, rule := range kalshiRules {
			if rule.pattern.MatchString(ticker) {
				return rule.category
			}
		}
	case "POLYMARKET":
		for _, rule := range polyRules {
			if rule.pattern.MatchString(title) {
				return rule.category
			}
		}
	}
	return ""
}

func SubcategorizeMarket(venue, ticker, title string) string {
	switch strings.ToUpper(venue) {
	case "KALSHI":
		// First try ticker-based subcategory rules
		for _, rule := range kalshiSubRules {
			if rule.pattern.MatchString(ticker) {
				return rule.subcategory
			}
		}
		// Fallback to title-based rules for KXMVESPORT/KXMVECROSS markets
		for _, rule := range kalshiTitleSubRules {
			if rule.pattern.MatchString(title) {
				return rule.subcategory
			}
		}
	case "POLYMARKET":
		for _, rule := range polySubRules {
			if rule.pattern.MatchString(title) {
				return rule.subcategory
			}
		}
	}
	return ""
}

func CategorizeMarketType(title string) string {
	for _, rule := range marketTypeRules {
		if rule.pattern.MatchString(title) {
			return rule.marketType
		}
	}
	return ""
}
