package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func StoreData(data *Data) error {
	store := PlayerStore{}
	store.Setup()

	for _, player := range data.Players {
		err := store.StorePlayer(player, data.CurrentGameweek().ID)
		if err != nil {
			return err
		}
	}

	return nil
}

type PlayerStore struct {
	Connection *sql.DB
}

func (p *PlayerStore) Connect() (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", "./players.sqlite")
	if err != nil {
		return nil, err
	}
	p.Connection = conn
	return conn, nil
}

func (p *PlayerStore) Close() error {
	return p.Connection.Close()
}

func (p *PlayerStore) Setup() error {
	db, _ := p.Connect()
	defer p.Close()

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS players (
		id INTEGER PRIMARY KEY,
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

func (p *PlayerStore) StorePlayer(player Player, gameweekID GameweekID) error {
	db, _ := p.Connect()
	defer p.Close()

	query := `
		INSERT INTO players (id, gameweek_id, name, form, points_per_game, total_points, cost, raw_cost, team_id, type_id, minutes, goals, assists, conceded, clean_sheets, yellow_cards, red_cards, bonus, starts, average_starts, matches_played, ict_index, ict_index_rank, most_captained, picked_percentage)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(query, player.ID, gameweekID, player.Name, player.Form, player.PointsPerGame, player.TotalPoints, player.Cost, player.RawCost, player.Team.ID, player.Type.ID, player.Stats.Minutes, player.Stats.Goals, player.Stats.Assists, player.Stats.Conceded, player.Stats.CleanSheets, player.Stats.YellowCards, player.Stats.RedCards, player.Stats.Bonus, player.Stats.Starts, player.Stats.AverageStarts, player.Stats.MatchesPlayed, player.Stats.ICTIndex, player.Stats.ICTIndexRank, player.MostCaptained, player.PickedPercentage)

	fmt.Print(result, err)

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
