package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
)

type likelyFixtureWinner struct {
	Player  Player
	Fixture Fixture
}

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
	gameWeekInt, err := strconv.Atoi(chosenGameweek)
	if err != nil {
		panic(err)
	}

	data, err := BuildData()
	if err != nil {
		panic(err)
	}

	likelyWinners := make([]likelyFixtureWinner, 0)
	for _, fixture := range data.FixturesByGameWeek(gameWeekInt) {
		var likelyWinner Team
		if fixture.HomeTeamDifficulty < fixture.AwayTeamDifficulty {
			likelyWinner = *fixture.HomeTeam
		} else if fixture.HomeTeamDifficulty > fixture.AwayTeamDifficulty {
			likelyWinner = *fixture.HomeTeam
		}
		for _, player := range likelyWinner.Players {
			likelyWinners = append(likelyWinners, likelyFixtureWinner{
				Player:  player,
				Fixture: fixture,
			})
		}
	}

	sort.Slice(likelyWinners, func(a, b int) bool {
		playerA := likelyWinners[a].Player
		playerB := likelyWinners[b].Player

		if playerA.Form != playerB.Form {
			return playerA.Form > playerB.Form
		}

		return likelyWinners[a].Fixture.DifficultyMajority > likelyWinners[b].Fixture.DifficultyMajority
	})

	// var bestTeam bestTeam

	// var bestTeam bestTeam
	// for playerTypeID, players := range likelyWinnerPlayersByType {
	// 	// expensive, probably
	// sort.Slice(players, func(i, j int) bool {
	// 	playerIForm, err := strconv.ParseFloat(players[i].Form, 32)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	playerJForm, err := strconv.ParseFloat(players[j].Form, 32)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	if playerIForm != playerJForm {
	// 		return playerIForm > playerJForm
	// 	}

	// 	playerITeamDifficultyMajority := likelyWinnerMap[players[i].TeamID].DifficultyMajority
	// 	playerJTeamDifficultyMajority := likelyWinnerMap[players[j].TeamID].DifficultyMajority

	// 	return playerITeamDifficultyMajority > playerJTeamDifficultyMajority
	// })

	// 	playerType := playerTypeMap[playerTypeID]

	goalkeepers := filterWinnersByType(likelyWinners, "Goalkeeper")

	fmt.Print(goalkeepers)

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

func filterWinnersByType(likelyWinners []likelyFixtureWinner, playerType string) []likelyFixtureWinner {
	out := make([]likelyFixtureWinner, 0)
	for _, likelyWinner := range likelyWinners {
		if likelyWinner.Player.Type.Name == playerType {
			out = append(out, likelyWinner)
		}
	}
	return out
}
