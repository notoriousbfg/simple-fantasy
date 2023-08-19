package main

import (
	"fmt"
	"os"
	"strconv"
)

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

	data, err := BuildData()
	if err != nil {
		panic(err)
	}

	for _, team := range data.Teams {
		fmt.Println(team)
	}

	// teamsMap := make(map[int]apiTeam, 0)
	// for _, team := range statsResp.Teams {
	// 	teamsMap[team.ID] = team
	// }

	// playerTypeMap := make(map[int]apiPlayerType, 0)
	// for _, playerType := range statsResp.PlayerTypes {
	// 	playerTypeMap[playerType.ID] = playerType
	// }

	// teamPlayerMap := make(map[int][]apiPlayer, 0)
	// for _, player := range statsResp.Players {
	// 	teamPlayerMap[player.TeamID] = append(teamPlayerMap[player.TeamID], player)
	// }

	// gameWeekFixtureMap := make(map[int][]apiFixture, 0)
	// for _, fixture := range fixturesList {
	// 	gameWeekFixtureMap[fixture.EventID] = append(gameWeekFixtureMap[fixture.EventID], fixture)
	// }

	gameWeekInt, err := strconv.Atoi(chosenGameweek)
	if err != nil {
		panic(err)
	}

	// gameweek := statsResp.Events[gameWeekInt-1]

	// fixtures := gameWeekFixtureMap[gameweek.ID]
	// likelyWinners := make([]apiTeam, 0)

	likelyWinners := make([]*Team, 0)

	for _, fixture := range data.FixturesByGameWeek(gameWeekInt) {
		if fixture.HomeTeamDifficulty < fixture.AwayTeamDifficulty {
			fixture.LikelyWinner = fixture.HomeTeam
			// likelyWinner.DifficultyMajority = fixture.AwayTeamDifficulty - fixture.HomeTeamDifficulty
			// likelyWinner.Opponent = &awayTeam
		} else if fixture.HomeTeamDifficulty > fixture.AwayTeamDifficulty {
			fixture.LikelyWinner = fixture.HomeTeam
			// likelyWinner.DifficultyMajority = fixture.HomeTeamDifficulty - fixture.AwayTeamDifficulty
			// likelyWinner.Opponent = &homeTeam
		}
	}

	// for _, fixture := range fixtures {
	// 	homeTeam := teamsMap[fixture.HomeTeamID]
	// 	awayTeam := teamsMap[fixture.AwayTeamID]

	// 	var likelyWinner apiTeam
	// 	if fixture.HomeTeamDifficulty < fixture.AwayTeamDifficulty {
	// 		likelyWinner = homeTeam
	// 		likelyWinner.DifficultyMajority = fixture.AwayTeamDifficulty - fixture.HomeTeamDifficulty
	// 		likelyWinner.Opponent = &awayTeam
	// 	} else if fixture.HomeTeamDifficulty > fixture.AwayTeamDifficulty {
	// 		likelyWinner = awayTeam
	// 		likelyWinner.DifficultyMajority = fixture.HomeTeamDifficulty - fixture.AwayTeamDifficulty
	// 		likelyWinner.Opponent = &homeTeam
	// 	}

	// 	if likelyWinner != (apiTeam{}) {
	// 		likelyWinners = append(likelyWinners, likelyWinner)
	// 	}
	// }

	// likelyWinnerMap := make(map[int]apiTeam, 0)
	// for _, winner := range likelyWinners {
	// 	likelyWinnerMap[winner.ID] = winner
	// }

	// likelyWinnerPlayersByType := make(map[int][]apiPlayer, 0)
	// for _, team := range likelyWinners {
	// 	for _, teamPlayer := range teamPlayerMap[team.ID] {
	// 		team := likelyWinnerMap[teamPlayer.TeamID]
	// 		teamPlayer.Team = &team
	// 		likelyWinnerPlayersByType[teamPlayer.TypeID] = append(
	// 			likelyWinnerPlayersByType[teamPlayer.TypeID],
	// 			teamPlayer,
	// 		)
	// 	}
	// }

	// var bestTeam bestTeam

	// var bestTeam bestTeam
	// for playerTypeID, players := range likelyWinnerPlayersByType {
	// 	// expensive, probably
	// 	sort.Slice(players, func(i, j int) bool {
	// 		playerIForm, err := strconv.ParseFloat(players[i].Form, 32)
	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		playerJForm, err := strconv.ParseFloat(players[j].Form, 32)
	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		if playerIForm != playerJForm {
	// 			return playerIForm > playerJForm
	// 		}

	// 		playerITeamDifficultyMajority := likelyWinnerMap[players[i].TeamID].DifficultyMajority
	// 		playerJTeamDifficultyMajority := likelyWinnerMap[players[j].TeamID].DifficultyMajority

	// 		return playerITeamDifficultyMajority > playerJTeamDifficultyMajority
	// 	})

	// 	playerType := playerTypeMap[playerTypeID]

	// 	// i'm sure there's a better way of doing this
	// 	if playerType.Name == "Goalkeepers" {
	// 		for i := 0; i < playerType.PlayerCount; i++ {
	// 			bestTeam.Goalkeepers = append(
	// 				bestTeam.Goalkeepers,
	// 				fmt.Sprintf("[%s] %s (%s)", players[i].Form, players[i].Name, players[i].Team.Opponent.Name),
	// 			)
	// 		}
	// 	}

	// 	if playerType.Name == "Defenders" {
	// 		for i := 0; i < playerType.PlayerCount; i++ {
	// 			bestTeam.Defenders = append(
	// 				bestTeam.Defenders,
	// 				fmt.Sprintf("[%s] %s (%s)", players[i].Form, players[i].Name, players[i].Team.Opponent.Name),
	// 			)
	// 		}
	// 	}

	// 	if playerType.Name == "Midfielders" {
	// 		for i := 0; i < playerType.PlayerCount; i++ {
	// 			bestTeam.Midfielders = append(
	// 				bestTeam.Midfielders,
	// 				fmt.Sprintf("[%s] %s (%s)", players[i].Form, players[i].Name, players[i].Team.Opponent.Name),
	// 			)
	// 		}
	// 	}

	// 	if playerType.Name == "Forwards" {
	// 		for i := 0; i < playerType.PlayerCount; i++ {
	// 			bestTeam.Forwards = append(
	// 				bestTeam.Forwards,
	// 				fmt.Sprintf("[%s] %s (%s)", players[i].Form, players[i].Name, players[i].Team.Opponent.Name),
	// 			)
	// 		}
	// 	}
	// }

	// fmt.Println()
	// fmt.Printf("The best team you could play in %s is: \n", gameweek.Name)
	// fmt.Println()

	// fmt.Println("Goalkeepers:")
	// for _, goalkeeper := range bestTeam.Goalkeepers {
	// 	fmt.Println(goalkeeper)
	// }
	// fmt.Println()

	// fmt.Println("Defenders:")
	// for _, defender := range bestTeam.Defenders {
	// 	fmt.Println(defender)
	// }
	// fmt.Println()

	// fmt.Println("Midfielders:")
	// for _, midfielder := range bestTeam.Midfielders {
	// 	fmt.Println(midfielder)
	// }
	// fmt.Println()

	// fmt.Println("Forwards:")
	// for _, forward := range bestTeam.Forwards {
	// 	fmt.Println(forward)
	// }
	// fmt.Println()
}
