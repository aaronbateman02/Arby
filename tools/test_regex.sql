SELECT 'yes Kevin Gausman: 3+,yes Freddy Peralta: 3+' ~* '\y([A-Z][a-z]+\s+[A-Z][a-z]+:\s*\d+\+)\y' as matches_go;
SELECT 'yes Fabian Marozsan,yes Jack Draper,yes Hugo Gaston' ~* '\y([A-Z][a-z]+\s+[A-Z][a-z]+)\y.*\y([A-Z][a-z]+\s+[A-Z][a-z]+)\y' as two_names;
