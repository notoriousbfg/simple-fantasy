package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbName = "./players.sqlite"
)

func StoreData(data *Data, gameweekInt int) error {
	store := PlayerStore{
		GameweekID: gameweekInt,
	}
	store.Setup()

	for _, playerType := range data.PlayerTypes {
		if err := store.StorePlayerType(playerType); err != nil {
			return err
		}
	}

	// for _, team := range data.Teams {
	// 	err := store.StoreTeam(team)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	for _, player := range data.Players {
		if err := store.StorePlayer(player); err != nil {
			return err
		}
	}

	if err := store.Dump(); err != nil {
		return err
	}

	return nil
}

type PlayerStore struct {
	GameweekID int
	Connection *sql.DB
}

func (p *PlayerStore) Connect() (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	p.Connection = conn
	return conn, nil
}

func (p *PlayerStore) Close() error {
	return p.Connection.Close()
}

func (p *PlayerStore) Dump() error {
	// this pattern isn't working
	p.Connect()
	defer p.Close()

	exportDir := fmt.Sprintf("./exports/gw_%d", p.GameweekID)
	err := os.Mkdir(exportDir, os.ModePerm)
	if err != nil {
		// end silently if dir already exists
		return nil
	}

	// Get a list of tables in the database
	tables, err := p.getTableNames()
	if err != nil {
		return err
	}

	// Dump each table to a separate SQL file
	for _, table := range tables {
		if err = p.dumpTableToFile(table, exportDir); err != nil {
			return err
		}
	}

	fmt.Println()

	return nil
}

func (p *PlayerStore) Setup() error {
	db, _ := p.Connect()
	defer p.Close()

	_, err := db.Exec(`DROP TABLE players; DROP TABLE player_types;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE player_types (
		id INTEGER PRIMARY KEY,
		name TEXT,
		plural_name TEXT,
		short_name TEXT,
		team_player_count INTEGER,
		team_min_play_count INTEGER,
		team_max_play_count INTEGER
	)`)

	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE players (
		gameweek_player_id VARCHAR PRIMARY KEY,
		id INTEGER,
		gameweek_id INTEGER,
		name TEXT,
		form REAL,
		points_per_game REAL,
		total_points INTEGER,
		cost TEXT,
		raw_cost REAL,
		team_id INTEGER,
		type_id INTEGER,
		minutes INTEGER,
		goals INTEGER,
		assists INTEGER,
		conceded INTEGER,
		clean_sheets INTEGER,
		yellow_cards INTEGER,
		red_cards INTEGER,
		bonus INTEGER,
		starts INTEGER,
		average_starts REAL,
		matches_played REAL,
		ict_index REAL,
		ict_index_rank INTEGER,
		most_captained BOOLEAN,
		picked_percentage REAL
	)`)

	if err != nil {
		return err
	}

	return nil
}

func (p *PlayerStore) StorePlayer(player Player) error {
	db, _ := p.Connect()
	defer p.Close()

	query := `
		INSERT OR IGNORE INTO players (gameweek_player_id, id, gameweek_id, name, form, points_per_game, total_points, cost, raw_cost, team_id, type_id, minutes, goals, assists, conceded, clean_sheets, yellow_cards, red_cards, bonus, starts, average_starts, matches_played, ict_index, ict_index_rank, most_captained, picked_percentage)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, fmt.Sprintf("%d_%d", p.GameweekID, player.ID), player.ID, p.GameweekID, player.Name, player.Form, player.PointsPerGame, player.TotalPoints, player.Cost, player.RawCost, player.Team.ID, player.Type.ID, player.Stats.Minutes, player.Stats.Goals, player.Stats.Assists, player.Stats.Conceded, player.Stats.CleanSheets, player.Stats.YellowCards, player.Stats.RedCards, player.Stats.Bonus, player.Stats.Starts, player.Stats.AverageStarts, player.Stats.MatchesPlayed, player.Stats.ICTIndex, player.Stats.ICTIndexRank, player.MostCaptained, player.PickedPercentage)

	if err != nil {
		return err
	}

	return nil
}

func (p *PlayerStore) StorePlayerType(playerType PlayerType) error {
	db, _ := p.Connect()
	defer p.Close()

	query := `
		INSERT OR IGNORE INTO player_types (id, name, plural_name, short_name, team_player_count, team_min_play_count, team_max_play_count)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, playerType.ID, playerType.Name, playerType.PluralName, playerType.ShortName, playerType.TeamPlayerCount, playerType.TeamMinPlayCount, playerType.TeamMaxPlayCount)

	if err != nil {
		return err
	}

	return nil
}

func (p *PlayerStore) GetPlayer(playerID PlayerID) (Player, error) {
	db, _ := p.Connect()
	defer p.Close()

	row := db.QueryRow(fmt.Sprintf("SELECT * FROM `players` WHERE `id` = %d", playerID))

	var player Player
	err := row.Scan(
		&player.ID,
		&player.Name,
	)

	if err != nil {
		return Player{}, err
	}

	fmt.Printf("player: %+v", player)
	return Player{}, nil
}

// getTableNames retrieves a list of table names from the SQLite database.
func (p *PlayerStore) getTableNames() ([]string, error) {
	rows, err := p.Connection.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// dumpTableToFile dumps the contents of the specified table to a SQL file.
func (p *PlayerStore) dumpTableToFile(tableName, tempDir string) error {
	outputFile := filepath.Join(tempDir, tableName+".sql")

	// use the sqlite3 command-line tool to dump the table to an SQL file
	cmd := exec.Command("sqlite3", dbName, fmt.Sprintf(".dump %s", tableName))
	dumpOutput, err := cmd.Output()
	if err != nil {
		return err
	}

	// write the dump output to the SQL file
	err = os.WriteFile(outputFile, dumpOutput, os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Printf("table '%s' dumped to '%s'\n", tableName, outputFile)

	return nil
}
