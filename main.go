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
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	DifficultyMajority int
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

type bestTeam struct {
	Goalkeepers []string
	Defenders   []string
	Midfielders []string
	Forwards    []string
}

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

	fmt.Printf("The best team you could play in %s is: \n", gameweek.Name)

	fmt.Println()

	fixtures := gameWeekFixtureMap[gameweek.ID]
	likelyWinners := make([]team, 0)

	for _, fixture := range fixtures {
		homeTeam := teamsMap[fixture.HomeTeamID]
		awayTeam := teamsMap[fixture.AwayTeamID]

		var likelyWinner team
		if fixture.HomeTeamDifficulty < fixture.AwayTeamDifficulty {
			likelyWinner = homeTeam
			likelyWinner.DifficultyMajority = fixture.AwayTeamDifficulty - fixture.HomeTeamDifficulty
		} else if fixture.HomeTeamDifficulty > fixture.AwayTeamDifficulty {
			likelyWinner = awayTeam
			likelyWinner.DifficultyMajority = fixture.HomeTeamDifficulty - fixture.AwayTeamDifficulty
		}

		if likelyWinner != (team{}) {
			likelyWinners = append(likelyWinners, likelyWinner)
		}
	}

	likelyWinnerMap := make(map[int]team, 0)
	for _, winner := range likelyWinners {
		likelyWinnerMap[winner.ID] = winner
	}

	likeWinnerPlayersByType := make(map[int][]player, 0)
	for _, team := range likelyWinners {
		for _, teamPlayer := range teamPlayerMap[team.ID] {
			likeWinnerPlayersByType[teamPlayer.TypeID] = append(likeWinnerPlayersByType[teamPlayer.TypeID], teamPlayer)
		}
	}

	var bestTeam bestTeam
	for playerTypeID, players := range likeWinnerPlayersByType {
		// expensive, probably
		sort.Slice(players, func(i, j int) bool {
			playerIForm, err := strconv.ParseFloat(players[i].Form, 32)
			if err != nil {
				panic(err)
			}

			playerJForm, err := strconv.ParseFloat(players[j].Form, 32)
			if err != nil {
				panic(err)
			}

			if playerIForm != playerJForm {
				return playerIForm > playerJForm
			}

			playerITeamDifficultyMajority := likelyWinnerMap[players[i].TeamID].DifficultyMajority
			playerJTeamDifficultyMajority := likelyWinnerMap[players[j].TeamID].DifficultyMajority

			return playerITeamDifficultyMajority > playerJTeamDifficultyMajority
		})

		playerType := playerTypeMap[playerTypeID]

		// i'm sure there's a better way of doing this
		if playerType.Name == "Goalkeepers" {
			for i := 0; i < playerType.PlayerCount; i++ {
				bestTeam.Goalkeepers = append(
					bestTeam.Goalkeepers,
					fmt.Sprintf("[%s] %s", players[i].Form, players[i].Name),
				)
			}
		}

		if playerType.Name == "Defenders" {
			for i := 0; i < playerType.PlayerCount; i++ {
				bestTeam.Defenders = append(
					bestTeam.Defenders,
					fmt.Sprintf("[%s] %s", players[i].Form, players[i].Name),
				)
			}
		}

		if playerType.Name == "Midfielders" {
			for i := 0; i < playerType.PlayerCount; i++ {
				bestTeam.Midfielders = append(
					bestTeam.Midfielders,
					fmt.Sprintf("[%s] %s", players[i].Form, players[i].Name),
				)
			}
		}

		if playerType.Name == "Forwards" {
			for i := 0; i < playerType.PlayerCount; i++ {
				bestTeam.Forwards = append(
					bestTeam.Forwards,
					fmt.Sprintf("[%s] %s", players[i].Form, players[i].Name),
				)
			}
		}
	}

	fmt.Println("Goalkeepers:")
	for _, goalkeeper := range bestTeam.Goalkeepers {
		fmt.Println(goalkeeper)
	}
	fmt.Println()

	fmt.Println("Defenders:")
	for _, defender := range bestTeam.Defenders {
		fmt.Println(defender)
	}
	fmt.Println()

	fmt.Println("Midfielders:")
	for _, midfielder := range bestTeam.Midfielders {
		fmt.Println(midfielder)
	}
	fmt.Println()

	fmt.Println("Forwards:")
	for _, forward := range bestTeam.Forwards {
		fmt.Println(forward)
	}
	fmt.Println()
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
