package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type apiElement struct {
	ID     int    `json:"id"`
	Name   string `json:"web_name"`
	Form   string `json:"form"`
	TypeID int    `json:"element_type"`
	TeamID int    `json:"team"`
}

type apiElementType struct {
	ID          int    `json:"id"`
	Name        string `json:"plural_name"`
	PlayerCount int    `json:"squad_select"`
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
	Fixtures []*Fixture
	Teams    []*Team
}

func (d *Data) FixturesByGameWeek(gameweek int) []*Fixture {
	var fixtures []*Fixture
	for _, fixture := range d.Fixtures {
		if GameweekID(gameweek) == fixture.Gameweek.ID {
			fixtures = append(fixtures, fixture)
		}
	}
	return fixtures
}

type PlayerTypeID int

type PlayerType struct {
	ID              PlayerTypeID
	Name            string
	TeamPlayerCount int
}

type PlayerID int

type Player struct {
	ID   PlayerID
	Name string
	Form float32
	Team *Team
	Type PlayerType
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
	ID   GameweekID
	Name string
}

type Fixture struct {
	Gameweek           *Gameweek
	HomeTeam           *Team
	AwayTeam           *Team
	LikelyWinner       *Team
	HomeTeamDifficulty int
	AwayTeamDifficulty int
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
		gameweeksByID[gameweekID] = &Gameweek{
			ID:   gameweekID,
			Name: apiEvent.Name,
		}
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
	for _, apiPlayerType := range statsResp.ElementTypes {
		playerTypeID := PlayerTypeID(apiPlayerType.ID)
		playerTypesByID[playerTypeID] = PlayerType{
			ID:              playerTypeID,
			Name:            apiPlayerType.Name,
			TeamPlayerCount: apiPlayerType.PlayerCount,
		}
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

		newPlayer := Player{
			ID:   PlayerID(apiPlayer.ID),
			Name: apiPlayer.Name,
			Form: float32(playerForm),
			Team: playerTeam,
			Type: playerType,
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

		newFixture := Fixture{
			Gameweek:           gameweeksByID[GameweekID(apiFixture.EventID)],
			HomeTeam:           homeTeam,
			AwayTeam:           awayTeam,
			HomeTeamDifficulty: apiFixture.HomeTeamDifficulty,
			AwayTeamDifficulty: apiFixture.AwayTeamDifficulty,
		}
		fixtures = append(fixtures, &newFixture)
	}
	data.Fixtures = fixtures

	return data, nil
}

// func ImportFixtures() ([]Fixture, error) {
// 	fixturesBody, err := getJsonBody(fixturesApi)
// 	if err != nil {
// 		panic(err)
// 	}
// 	var apiFixtures apiFixtures
// 	if err := json.Unmarshal(fixturesBody, &apiFixtures); err != nil {
// 		panic(err)
// 	}
// 	fixtures := make([]Fixture, 0)
// 	for _, apiFixture := range apiFixtures {
// 		fixtures = append(fixtures, Fixture{})
// 	}
// 	return fixtures, nil
// }

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
