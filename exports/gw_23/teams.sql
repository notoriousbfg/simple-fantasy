PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE teams (
		id INTEGER PRIMARY KEY,
		name TEXT,
		short_name TEXT
	);
INSERT INTO teams VALUES(2,'Aston Villa','AVL');
INSERT INTO teams VALUES(4,'Brentford','BRE');
INSERT INTO teams VALUES(7,'Chelsea','CHE');
INSERT INTO teams VALUES(10,'Fulham','FUL');
INSERT INTO teams VALUES(11,'Liverpool','LIV');
INSERT INTO teams VALUES(12,'Luton','LUT');
INSERT INTO teams VALUES(13,'Man City','MCI');
INSERT INTO teams VALUES(20,'Wolves','WOL');
COMMIT;
