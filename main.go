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

	var bestTeam bestTeam

	gkType := data.PlayerType("Goalkeeper")
	goalkeepers := filterWinnersByType(likelyWinners, "Goalkeeper")
	for i := 0; i < gkType.TeamPlayerCount; i++ {
		bestTeam.Goalkeepers = append(
			bestTeam.Goalkeepers,
			fmt.Sprintf("[%v] %s", goalkeepers[i].Player.Form, goalkeepers[i].Player.Name),
		)
	}

	defType := data.PlayerType("Defender")
	defenders := filterWinnersByType(likelyWinners, "Defender")
	for i := 0; i < defType.TeamPlayerCount; i++ {
		bestTeam.Defenders = append(
			bestTeam.Defenders,
			fmt.Sprintf("[%v] %s", defenders[i].Player.Form, defenders[i].Player.Name),
		)
	}

	midType := data.PlayerType("Midfielder")
	midfielders := filterWinnersByType(likelyWinners, "Midfielder")
	for i := 0; i < midType.TeamPlayerCount; i++ {
		bestTeam.Midfielders = append(
			bestTeam.Midfielders,
			fmt.Sprintf("[%v] %s", midfielders[i].Player.Form, midfielders[i].Player.Name),
		)
	}

	fwdType := data.PlayerType("Forward")
	forwards := filterWinnersByType(likelyWinners, "Forward")
	for i := 0; i < fwdType.TeamPlayerCount; i++ {
		bestTeam.Forwards = append(
			bestTeam.Forwards,
			fmt.Sprintf("[%v] %s", forwards[i].Player.Form, forwards[i].Player.Name),
		)
	}

	gameweek := data.Gameweek(gameWeekInt)

	fmt.Printf("The best team you could play in %s is: \n\n", gameweek.Name)

	fmt.Println("Goalkeepers:")
	for _, goalkeeper := range bestTeam.Goalkeepers {
		fmt.Println(goalkeeper)
	}

	fmt.Println("\nDefenders:")
	for _, defender := range bestTeam.Defenders {
		fmt.Println(defender)
	}
	fmt.Println()

	fmt.Println("\nMidfielders:")
	for _, midfielder := range bestTeam.Midfielders {
		fmt.Println(midfielder)
	}
	fmt.Println()

	fmt.Println("\nForwards:")
	for _, forward := range bestTeam.Forwards {
		fmt.Println(forward)
	}
	fmt.Println()
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
