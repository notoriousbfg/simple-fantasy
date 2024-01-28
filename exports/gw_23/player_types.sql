PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE player_types (
		id INTEGER PRIMARY KEY,
		name TEXT,
		plural_name TEXT,
		short_name TEXT,
		team_player_count INTEGER,
		team_min_play_count INTEGER,
		team_max_play_count INTEGER
	);
INSERT INTO player_types VALUES(1,'Goalkeeper','Goalkeepers','GKP',2,1,1);
INSERT INTO player_types VALUES(2,'Defender','Defenders','DEF',5,3,5);
INSERT INTO player_types VALUES(3,'Midfielder','Midfielders','MID',5,2,5);
INSERT INTO player_types VALUES(4,'Forward','Forwards','FWD',3,1,3);
COMMIT;
