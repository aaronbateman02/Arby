select id, venue, venue_market_id, title, description, category, subcategory, market_type, structure_type, status, resolution_date 
from markets 
where venue = 'KALSHI' 
  and (venue_market_id like '%KXMLBGAME%' or title ilike '%st.%20louis%' or title ilike '%cardinal%' or title ilike '%diamondback%')
limit 20;
