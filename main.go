package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
)

const (
	fixturesApi = "https://fantasy.premierleague.com/api/fixtures/"
	statsApi    = "https://fantasy.premierleague.com/api/bootstrap-static/"
)

type team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type event struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type player struct {
	ID     int    `json:"id"`
	Name   string `json:"web_name"`
	Form   string `json:"form"`
	TypeID int    `json:"element_type"`
	TeamID int    `json:"team"`
}

type playerType struct {
	ID          int    `json:"id"`
	Name        string `json:"plural_name"`
	PlayerCount int    `json:"squad_select"`
}

type stats struct {
	Teams       []team       `json:"teams"`
	Events      []event      `json:"events"`
	Players     []player     `json:"elements"`
	PlayerTypes []playerType `json:"element_types"`
}

type fixture struct {
	AwayTeamID         int `json:"team_a"`
	HomeTeamID         int `json:"team_h"`
	EventID            int `json:"event"`
	AwayTeamDifficulty int `json:"team_a_difficulty"`
	HomeTeamDifficulty int `json:"team_h_difficulty"`
}

type fixtures []fixture

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		panic("You must provide the gameweek number")
	}

	chosenGameweek := args[0]

	fixturesBody, err := getJsonBody(fixturesApi)
	if err != nil {
		panic(err)
	}
	var fixturesList fixtures
	if err := json.Unmarshal(fixturesBody, &fixturesList); err != nil {
		panic(err)
	}
	statsBody, err := getJsonBody(statsApi)
	if err != nil {
		panic(err)
	}
	var statsResp stats
	if err := json.Unmarshal(statsBody, &statsResp); err != nil {
		panic(err)
	}

	teamsMap := make(map[int]team, 0)
	for _, team := range statsResp.Teams {
		teamsMap[team.ID] = team
	}

	playerTypeMap := make(map[int]playerType, 0)
	for _, playerType := range statsResp.PlayerTypes {
		playerTypeMap[playerType.ID] = playerType
	}

	teamPlayerMap := make(map[int][]player, 0)
	for _, player := range statsResp.Players {
		teamPlayerMap[player.TeamID] = append(teamPlayerMap[player.TeamID], player)
	}

	gameWeekFixtureMap := make(map[int][]fixture, 0)
	for _, fixture := range fixturesList {
		gameWeekFixtureMap[fixture.EventID] = append(gameWeekFixtureMap[fixture.EventID], fixture)
	}

	gameWeekInt, err := strconv.Atoi(chosenGameweek)
	if err != nil {
		panic(err)
	}

	gameweek := statsResp.Events[gameWeekInt-1]

	fmt.Println()

	fmt.Printf("The best players you could play in %s are: \n", gameweek.Name)

	fmt.Println()

	fixtures := gameWeekFixtureMap[gameweek.ID]
	likelyWinners := make([]team, 0)

	for _, fixture := range fixtures {
		homeTeam := teamsMap[fixture.HomeTeamID]
		awayTeam := teamsMap[fixture.AwayTeamID]

		var likelyWinner team
		if fixture.HomeTeamDifficulty < fixture.AwayTeamDifficulty {
			likelyWinner = homeTeam
		} else if fixture.HomeTeamDifficulty > fixture.AwayTeamDifficulty {
			likelyWinner = awayTeam
		}

		if likelyWinner != (team{}) {
			likelyWinners = append(likelyWinners, likelyWinner)
		}
	}

	likeWinnerPlayersByType := make(map[int][]player, 0)
	for _, team := range likelyWinners {
		for _, teamPlayer := range teamPlayerMap[team.ID] {
			likeWinnerPlayersByType[teamPlayer.TypeID] = append(likeWinnerPlayersByType[teamPlayer.TypeID], teamPlayer)
		}
	}

	for playerTypeID, players := range likeWinnerPlayersByType {
		// expensive, probably
		sort.Slice(players, func(i, j int) bool {
			return players[i].Form > players[j].Form
		})

		playerType := playerTypeMap[playerTypeID]

		fmt.Printf("%s: \n", playerType.Name)

		for i := 0; i < playerType.PlayerCount; i++ {
			fmt.Println(players[i].Name)
		}

		fmt.Println()
	}

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
