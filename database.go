package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

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
