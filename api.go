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
	fixturesApi       = "https://fantasy.premierleague.com/api/fixtures/"
	statsApi          = "https://fantasy.premierleague.com/api/bootstrap-static/"
	playerFixturesApi = "https://fantasy.premierleague.com/api/element-summary/"
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
	IsCurrent       bool      `json:"is_current"`
	IsNext          bool      `json:"is_next"`
	Finished        bool      `json:"finished"`
	MostCaptainedID int       `json:"most_captained"`
}

type apiElement struct {
	ID                       int     `json:"id"`
	Name                     string  `json:"web_name"`
	Form                     string  `json:"form"`
	PointsPerGame            string  `json:"points_per_game"`
	TotalPoints              int     `json:"total_points"`
	Cost                     int     `json:"now_cost"`
	TypeID                   int     `json:"element_type"`
	TeamID                   int     `json:"team"`
	Minutes                  int     `json:"minutes"`
	Goals                    int     `json:"goals_scored"`
	Assists                  int     `json:"assists"`
	Conceded                 int     `json:"goals_conceded"`
	CleanSheets              int     `json:"clean_sheets"`
	YellowCards              int     `json:"yellow_cards"`
	RedCards                 int     `json:"red_cards"`
	Bonus                    int     `json:"bonus"`
	Starts                   int     `json:"starts"`
	StartsPerNinety          float32 `json:"starts_per_90"`
	ICTIndex                 string  `json:"ict_index"`
	ICTIndexRank             int     `json:"ict_index_rank"`
	News                     string  `json:"news"`
	ChanceOfPlayingThisRound *int    `json:"chance_of_playing_this_round"`
	ChanceOfPlayingNextRound *int    `json:"chance_of_playing_next_round"`
	SelectedByPercent        string  `json:"selected_by_percent"`
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

type apiPlayerFixturesAndHistory struct {
	// Fixtures []apiPlayerFixture `json:"fixtures"`
	History []apiPlayerHistory `json:"history"`
}

// type apiPlayerFixture struct {
// }

type apiPlayerHistory struct {
	ElementID   int `json:"element"`
	FixtureID   int `json:"fixture"`
	Minutes     int `json:"minutes"`
	TotalPoints int `json:"total_points"`
}

type apiFixture struct {
	ID                 int `json:"id"`
	AwayTeamID         int `json:"team_a"`
	HomeTeamID         int `json:"team_h"`
	EventID            int `json:"event"`
	AwayTeamDifficulty int `json:"team_a_difficulty"`
	HomeTeamDifficulty int `json:"team_h_difficulty"`
}

type apiFixtures []apiFixture

type apiPicks struct {
	Picks        []apiPick       `json:"picks"`
	EntryHistory apiEntryHistory `json:"entry_history"`
}

type apiPick struct {
	Element   int  `json:"element"`
	IsCaptain bool `json:"is_captain"`
}

type apiEntryHistory struct {
	Bank float32 `json:"bank"`
}

type Data struct {
	PlayerTypes []PlayerType
	Gameweeks   []Gameweek
	Fixtures    []*Fixture
	Teams       []*Team
	Players     []Player
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

func (d *Data) CurrentGameweek() *Gameweek {
	for _, gameweek := range d.Gameweeks {
		if gameweek.IsCurrent {
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

func (d *Data) GameweekPlayers(gameweek int) []StartingPlayer {
	gameweekPlayers := make([]StartingPlayer, 0)
	for _, fixture := range d.FixturesByGameWeek(gameweek) {
		for _, player := range fixture.HomeTeam.Players {
			gameweekPlayers = append(gameweekPlayers, StartingPlayer{
				Player:       player,
				Fixture:      fixture,
				OpposingTeam: *fixture.AwayTeam,
			})
		}
		for _, player := range fixture.AwayTeam.Players {
			gameweekPlayers = append(gameweekPlayers, StartingPlayer{
				Player:       player,
				Fixture:      fixture,
				OpposingTeam: *fixture.HomeTeam,
			})
		}
	}
	return gameweekPlayers
}

func (d *Data) GameweekPlayerSet(gameweek GameweekID) map[PlayerID]StartingPlayer {
	playerSet := make(map[PlayerID]StartingPlayer, 0)
	for _, player := range d.GameweekPlayers(int(gameweek)) {
		playerSet[player.Player.ID] = player
	}
	return playerSet
}

func (d *Data) RequestManagerPicks(managerID int) TeamConfig {
	endpoint := fmt.Sprintf("https://fantasy.premierleague.com/api/entry/%d/event/%d/picks/", managerID, d.CurrentGameweek().ID)

	teamBody, err := getJsonBody(endpoint)
	if err != nil {
		panic(err)
	}

	var apiPicks apiPicks
	if err := json.Unmarshal(teamBody, &apiPicks); err != nil {
		panic(err)
	}

	gameweekPlayerSet := d.GameweekPlayerSet(d.CurrentGameweek().ID)

	players := make([]StartingPlayer, 0)
	for _, pick := range apiPicks.Picks {
		thisPlayer, ok := gameweekPlayerSet[PlayerID(pick.Element)]
		if ok {
			players = append(players, thisPlayer)
		}
	}

	return TeamConfig{
		Players:   players,
		BankValue: apiPicks.EntryHistory.Bank,
	}
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
	Starts        int
	AverageStarts float32
	MatchesPlayed float32
	ICTIndex      float32
	ICTIndexRank  int
}

type PlayerHistory struct {
	Fixture *Fixture
	Minutes int
}

type PlayerRoundProbability map[GameweekID]float32

type PlayerID int

type Player struct {
	ID               PlayerID
	Name             string
	Form             float32
	PointsPerGame    float32
	TotalPoints      int
	Cost             string
	RawCost          float32
	Team             *Team
	Type             PlayerType
	Stats            PlayerStats
	History          map[FixtureID]PlayerFixture
	ChanceOfPlaying  PlayerRoundProbability
	MostCaptained    bool
	PickedPercentage float32
}

type PlayerFixture struct {
	FixtureID FixtureID
	PlayerID  PlayerID
	Minutes   int
	Played    bool
	Points    int
}

type TeamID int

type Team struct {
	ID        TeamID
	Name      string
	ShortName string
	Players   []Player
	Fixtures  []Fixture
}

type GameweekID int

type Gameweek struct {
	ID              GameweekID
	Name            string
	Deadline        string
	IsCurrent       bool
	IsNext          bool
	Finished        bool
	MostCaptainedID PlayerID
}

type FixtureID int

type Fixture struct {
	ID                 FixtureID
	Gameweek           *Gameweek
	HomeTeam           *Team
	AwayTeam           *Team
	HomeTeamDifficulty int
	AwayTeamDifficulty int
	DifficultyMajority int
}

func (f *Fixture) Players() []Player {
	players := make([]Player, 0)
	players = append(players, f.HomeTeam.Players...)
	players = append(players, f.AwayTeam.Players...)
	return players
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

	var currentGameweekID GameweekID

	gameweeksByID := make(map[GameweekID]*Gameweek, 0)
	for _, apiEvent := range statsResp.Events {
		gameweekID := GameweekID(apiEvent.ID)
		if apiEvent.IsCurrent {
			currentGameweekID = gameweekID
		}
		gameweek := &Gameweek{
			ID:              gameweekID,
			Name:            apiEvent.Name,
			Deadline:        apiEvent.Deadline.Format("02 Jan 15:04"),
			IsCurrent:       apiEvent.IsCurrent,
			IsNext:          apiEvent.IsNext,
			Finished:        apiEvent.Finished,
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
	allPlayers := make([]Player, 0)
	for _, apiPlayer := range statsResp.Elements {
		playerForm, err := strconv.ParseFloat(apiPlayer.Form, 32)
		if err != nil {
			return &Data{}, err
		}

		playerPointsPerGame, err := strconv.ParseFloat(apiPlayer.PointsPerGame, 32)
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

		formattedCost := fmt.Sprintf("£%.1fm", float32(apiPlayer.Cost)/float32(10))

		var chanceOfPlayingThisRound float32
		if apiPlayer.ChanceOfPlayingThisRound == nil {
			chanceOfPlayingThisRound = 1
		} else {
			chanceOfPlayingThisRound = float32(*apiPlayer.ChanceOfPlayingThisRound) / 100
		}

		var chanceOfPlayingNextRound float32
		if apiPlayer.ChanceOfPlayingNextRound == nil {
			chanceOfPlayingNextRound = 1
		} else {
			chanceOfPlayingNextRound = float32(*apiPlayer.ChanceOfPlayingNextRound) / 100
		}

		chanceOfPlaying := map[GameweekID]float32{
			currentGameweekID:     chanceOfPlayingThisRound,
			currentGameweekID + 1: chanceOfPlayingNextRound, // assumes next round is gameweek ID + 1
		}

		pickedPercentage, err := strconv.ParseFloat(apiPlayer.SelectedByPercent, 32)
		if err != nil {
			return &Data{}, err
		}

		newPlayer := Player{
			ID:            PlayerID(apiPlayer.ID),
			Name:          apiPlayer.Name,
			Form:          float32(playerForm),
			PointsPerGame: float32(playerPointsPerGame),
			TotalPoints:   apiPlayer.TotalPoints,
			Cost:          formattedCost,
			RawCost:       float32(apiPlayer.Cost) / float32(10),
			Team:          playerTeam,
			Type:          playerType,
			Stats: PlayerStats{
				Minutes:       apiPlayer.Minutes,
				Goals:         apiPlayer.Goals,
				Assists:       apiPlayer.Assists,
				Conceded:      apiPlayer.Conceded,
				CleanSheets:   apiPlayer.CleanSheets,
				YellowCards:   apiPlayer.YellowCards,
				RedCards:      apiPlayer.RedCards,
				Bonus:         apiPlayer.Bonus,
				Starts:        apiPlayer.Starts,
				AverageStarts: apiPlayer.StartsPerNinety,
				ICTIndex:      float32(ictIndex),
				ICTIndexRank:  apiPlayer.ICTIndexRank,
			},
			ChanceOfPlaying:  chanceOfPlaying,
			PickedPercentage: float32(pickedPercentage),
		}

		teamPlayersByID[newPlayer.Team.ID] = append(
			teamPlayersByID[TeamID(newPlayer.Team.ID)],
			newPlayer,
		)
		allPlayers = append(allPlayers, newPlayer)
	}

	data.Players = allPlayers

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
			ID:                 FixtureID(apiFixture.ID),
			Gameweek:           gameweek,
			HomeTeam:           homeTeam,
			AwayTeam:           awayTeam,
			HomeTeamDifficulty: apiFixture.HomeTeamDifficulty,
			AwayTeamDifficulty: apiFixture.AwayTeamDifficulty,
			DifficultyMajority: abs(apiFixture.HomeTeamDifficulty - apiFixture.AwayTeamDifficulty),
		}
		fixtures = append(fixtures, &newFixture)

		if team, ok := teamsByID[TeamID(apiFixture.HomeTeamID)]; ok {
			team.Fixtures = append(team.Fixtures, newFixture)
		}

		if team, ok := teamsByID[TeamID(apiFixture.AwayTeamID)]; ok {
			team.Fixtures = append(team.Fixtures, newFixture)
		}
	}
	data.Fixtures = fixtures

	return data, nil
}

func requestPlayerHistory(apiPlayerID int) (map[FixtureID]PlayerFixture, error) {
	fixturesAndHistoryApiBody, err := getJsonBody(fmt.Sprintf("%s/%d", playerFixturesApi, apiPlayerID))
	if err != nil {
		return nil, err
	}
	var fixturesAndHistory apiPlayerFixturesAndHistory
	if err := json.Unmarshal(fixturesAndHistoryApiBody, &fixturesAndHistory); err != nil {
		return nil, err
	}
	fixturesToPlayerFixtures := make(map[FixtureID]PlayerFixture, 0)
	for _, fixture := range fixturesAndHistory.History {
		fixturesToPlayerFixtures[FixtureID(fixture.FixtureID)] = PlayerFixture{
			FixtureID: FixtureID(fixture.FixtureID),
			PlayerID:  PlayerID(fixture.ElementID),
			Minutes:   fixture.Minutes,
			Played:    fixture.Minutes > 0,
			Points:    fixture.TotalPoints,
		}
	}

	return fixturesToPlayerFixtures, nil
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
