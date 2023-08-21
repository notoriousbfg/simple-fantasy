package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	fixturesApi = "https://fantasy.premierleague.com/api/fixtures/"
	statsApi    = "https://fantasy.premierleague.com/api/bootstrap-static/"
)

type apiTeam struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

type apiEvent struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Deadline        time.Time `json:"deadline_time"`
	MostCaptainedID int       `json:"most_captained"`
}

type apiElement struct {
	ID              int     `json:"id"`
	Name            string  `json:"web_name"`
	Form            string  `json:"form"`
	Cost            int     `json:"now_cost"`
	TypeID          int     `json:"element_type"`
	TeamID          int     `json:"team"`
	Minutes         int     `json:"minutes"`
	Goals           int     `json:"goals_scored"`
	Assists         int     `json:"assists"`
	Conceded        int     `json:"goals_conceded"`
	CleanSheets     int     `json:"clean_sheets"`
	YellowCards     int     `json:"yellow_cards"`
	RedCards        int     `json:"red_cards"`
	Bonus           int     `json:"bonus"`
	StartsPerNinety float32 `json:"starts_per_90"`
	ICTIndex        string  `json:"ict_index"`
	ICTIndexRank    int     `json:"ict_index_rank"`
}

type apiElementType struct {
	ID           int    `json:"id"`
	Name         string `json:"singular_name"`
	PluralName   string `json:"plural_name"`
	ShortName    string `json:"singular_name_short"`
	PlayerCount  int    `json:"squad_select"`
	SquadMinPlay int    `json:"squad_min_play"`
	SquadMaxPlay int    `json:"squad_max_play"`
}

type apiStats struct {
	Teams        []apiTeam        `json:"teams"`
	Events       []apiEvent       `json:"events"`
	Elements     []apiElement     `json:"elements"`
	ElementTypes []apiElementType `json:"element_types"`
}

type apiFixture struct {
	AwayTeamID         int `json:"team_a"`
	HomeTeamID         int `json:"team_h"`
	EventID            int `json:"event"`
	AwayTeamDifficulty int `json:"team_a_difficulty"`
	HomeTeamDifficulty int `json:"team_h_difficulty"`
}

type apiFixtures []apiFixture

type Data struct {
	PlayerTypes []PlayerType
	Gameweeks   []Gameweek
	Fixtures    []*Fixture
	Teams       []*Team
}

func (d *Data) FixturesByGameWeek(gameweek int) []Fixture {
	fixtures := make([]Fixture, 0)
	for _, fixture := range d.Fixtures {
		if GameweekID(gameweek) == fixture.Gameweek.ID {
			fixtures = append(fixtures, *fixture)
		}
	}
	return fixtures
}

func (d *Data) Gameweek(gw int) *Gameweek {
	for _, gameweek := range d.Gameweeks {
		if gameweek.ID == GameweekID(gw) {
			return &gameweek
		}
	}
	return nil
}

func (d *Data) PlayerType(pt string) *PlayerType {
	for _, playerType := range d.PlayerTypes {
		if playerType.Name == pt {
			return &playerType
		}
	}
	return nil
}

type PlayerTypeID int

type PlayerType struct {
	ID               PlayerTypeID
	Name             string
	PluralName       string
	ShortName        string
	TeamPlayerCount  int
	TeamMinPlayCount int
	TeamMaxPlayCount int
}

type PlayerStats struct {
	Minutes       int
	Goals         int
	Assists       int
	Conceded      int
	CleanSheets   int
	YellowCards   int
	RedCards      int
	Bonus         int
	AverageStarts float32
	ICTIndex      float32
	ICTIndexRank  int
}

type PlayerID int

type Player struct {
	ID            PlayerID
	Name          string
	Form          float32
	Cost          string
	Team          *Team
	Type          PlayerType
	Stats         PlayerStats
	MostCaptained bool
}

type TeamID int

type Team struct {
	ID        TeamID
	Name      string
	ShortName string
	Players   []Player
}

type GameweekID int

type Gameweek struct {
	ID              GameweekID
	Name            string
	Deadline        string
	MostCaptainedID PlayerID
}

type Fixture struct {
	Gameweek           *Gameweek
	HomeTeam           *Team
	AwayTeam           *Team
	HomeTeamDifficulty int
	AwayTeamDifficulty int
	DifficultyMajority int
}

func BuildData() (*Data, error) {
	data := &Data{}

	statsApiBody, err := getJsonBody(statsApi)
	if err != nil {
		panic(err)
	}
	var statsResp apiStats
	if err := json.Unmarshal(statsApiBody, &statsResp); err != nil {
		panic(err)
	}

	gameweeksByID := make(map[GameweekID]*Gameweek, 0)
	for _, apiEvent := range statsResp.Events {
		gameweekID := GameweekID(apiEvent.ID)
		gameweek := &Gameweek{
			ID:              gameweekID,
			Name:            apiEvent.Name,
			Deadline:        apiEvent.Deadline.Format("02 Jan 15:04"),
			MostCaptainedID: PlayerID(apiEvent.MostCaptainedID),
		}
		gameweeksByID[gameweekID] = gameweek
		data.Gameweeks = append(data.Gameweeks, *gameweek)
	}

	var teams []*Team
	teamsByID := make(map[TeamID]*Team, 0)
	for _, apiTeam := range statsResp.Teams {
		newTeam := Team{
			ID:        TeamID(apiTeam.ID),
			Name:      apiTeam.Name,
			ShortName: apiTeam.ShortName,
		}
		teams = append(teams, &newTeam)
		teamsByID[newTeam.ID] = &newTeam
	}
	data.Teams = teams

	playerTypesByID := make(map[PlayerTypeID]PlayerType, 0)
	for _, apiElementType := range statsResp.ElementTypes {
		newType := PlayerType{
			ID:               PlayerTypeID(apiElementType.ID),
			Name:             apiElementType.Name,
			PluralName:       apiElementType.PluralName,
			ShortName:        apiElementType.ShortName,
			TeamPlayerCount:  apiElementType.PlayerCount,
			TeamMinPlayCount: apiElementType.SquadMinPlay,
			TeamMaxPlayCount: apiElementType.SquadMaxPlay,
		}
		playerTypeID := PlayerTypeID(apiElementType.ID)
		playerTypesByID[playerTypeID] = newType
		data.PlayerTypes = append(data.PlayerTypes, newType)
	}

	teamPlayersByID := make(map[TeamID][]Player, 0)
	for _, apiPlayer := range statsResp.Elements {
		playerForm, err := strconv.ParseFloat(apiPlayer.Form, 32)
		if err != nil {
			return &Data{}, err
		}

		playerTeam, ok := teamsByID[TeamID(apiPlayer.TeamID)]
		if !ok {
			return &Data{}, fmt.Errorf("missing team ID '%d'", apiPlayer.TeamID)
		}

		playerType, ok := playerTypesByID[PlayerTypeID(apiPlayer.TypeID)]
		if !ok {
			return &Data{}, fmt.Errorf("missing player type ID '%d'", apiPlayer.TypeID)
		}

		ictIndex, err := strconv.ParseFloat(apiPlayer.ICTIndex, 32)
		if err != nil {
			return &Data{}, err
		}

		newPlayer := Player{
			ID:   PlayerID(apiPlayer.ID),
			Name: apiPlayer.Name,
			Form: float32(playerForm),
			Cost: fmt.Sprintf("£%.1fm", float32(apiPlayer.Cost)/float32(10)),
			Team: playerTeam,
			Type: playerType,
			Stats: PlayerStats{
				Minutes:       apiPlayer.Minutes,
				Goals:         apiPlayer.Goals,
				Assists:       apiPlayer.Assists,
				Conceded:      apiPlayer.Conceded,
				CleanSheets:   apiPlayer.CleanSheets,
				YellowCards:   apiPlayer.YellowCards,
				RedCards:      apiPlayer.RedCards,
				Bonus:         apiPlayer.Bonus,
				AverageStarts: apiPlayer.StartsPerNinety,
				ICTIndex:      float32(ictIndex),
				ICTIndexRank:  apiPlayer.ICTIndexRank,
			},
		}

		teamPlayersByID[newPlayer.Team.ID] = append(
			teamPlayersByID[TeamID(newPlayer.Team.ID)],
			newPlayer,
		)
	}

	for _, team := range teams {
		team.Players = teamPlayersByID[team.ID]
	}

	fixturesBody, err := getJsonBody(fixturesApi)
	if err != nil {
		panic(err)
	}

	var apiFixtures apiFixtures
	if err := json.Unmarshal(fixturesBody, &apiFixtures); err != nil {
		panic(err)
	}

	fixtures := make([]*Fixture, 0)
	for _, apiFixture := range apiFixtures {
		homeTeam := teamsByID[TeamID(apiFixture.HomeTeamID)]
		awayTeam := teamsByID[TeamID(apiFixture.AwayTeamID)]

		gameweek, ok := gameweeksByID[GameweekID(apiFixture.EventID)]
		if !ok {
			continue
		}

		newFixture := Fixture{
			Gameweek:           gameweek,
			HomeTeam:           homeTeam,
			AwayTeam:           awayTeam,
			HomeTeamDifficulty: apiFixture.HomeTeamDifficulty,
			AwayTeamDifficulty: apiFixture.AwayTeamDifficulty,
			DifficultyMajority: abs(apiFixture.HomeTeamDifficulty - apiFixture.AwayTeamDifficulty),
		}
		fixtures = append(fixtures, &newFixture)
	}
	data.Fixtures = fixtures

	return data, nil
}

func getJsonBody(endpoint string) ([]byte, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return body, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
