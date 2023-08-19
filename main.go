package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
)

type likelyFixtureWinner struct {
	Player       Player
	Fixture      Fixture
	OpposingTeam Team
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
		panic("You must provide a gameweek number")
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
		var opposingTeam Team
		if fixture.HomeTeamDifficulty < fixture.AwayTeamDifficulty {
			likelyWinner = *fixture.HomeTeam
			opposingTeam = *fixture.AwayTeam
		} else if fixture.HomeTeamDifficulty > fixture.AwayTeamDifficulty {
			likelyWinner = *fixture.AwayTeam
			opposingTeam = *fixture.HomeTeam
		}
		for _, player := range likelyWinner.Players {
			likelyWinners = append(likelyWinners, likelyFixtureWinner{
				Player:       player,
				Fixture:      fixture,
				OpposingTeam: opposingTeam,
			})
		}
	}

	sort.Slice(likelyWinners, func(a, b int) bool {
		playerA := likelyWinners[a].Player
		playerB := likelyWinners[b].Player

		if playerA.Form != playerB.Form {
			return playerA.Form > playerB.Form
		}

		if playerA.Stats.ICTIndexRank != playerB.Stats.ICTIndexRank {
			return playerA.Stats.ICTIndexRank > playerB.Stats.ICTIndexRank
		}

		if playerA.Stats.AverageStarts != playerB.Stats.AverageStarts {
			return playerA.Stats.AverageStarts > playerB.Stats.AverageStarts
		}

		return likelyWinners[a].Fixture.DifficultyMajority > likelyWinners[b].Fixture.DifficultyMajority
	})

	var bestTeam bestTeam

	gkType := data.PlayerType("Goalkeeper")
	goalkeepers := filterWinnersByType(likelyWinners, "Goalkeeper")
	for i := 0; i < gkType.TeamPlayerCount; i++ {
		bestTeam.Goalkeepers = append(
			bestTeam.Goalkeepers,
			fmt.Sprintf("[%v] %s (%s)", goalkeepers[i].Player.Form, goalkeepers[i].Player.Name, goalkeepers[i].OpposingTeam.ShortName),
		)
	}

	defType := data.PlayerType("Defender")
	defenders := filterWinnersByType(likelyWinners, "Defender")
	for i := 0; i < defType.TeamPlayerCount; i++ {
		bestTeam.Defenders = append(
			bestTeam.Defenders,
			fmt.Sprintf("[%v] %s (%s)", defenders[i].Player.Form, defenders[i].Player.Name, defenders[i].OpposingTeam.ShortName),
		)
	}

	midType := data.PlayerType("Midfielder")
	midfielders := filterWinnersByType(likelyWinners, "Midfielder")
	for i := 0; i < midType.TeamPlayerCount; i++ {
		bestTeam.Midfielders = append(
			bestTeam.Midfielders,
			fmt.Sprintf("[%v] %s (%s)", midfielders[i].Player.Form, midfielders[i].Player.Name, midfielders[i].OpposingTeam.ShortName),
		)
	}

	fwdType := data.PlayerType("Forward")
	forwards := filterWinnersByType(likelyWinners, "Forward")
	for i := 0; i < fwdType.TeamPlayerCount; i++ {
		bestTeam.Forwards = append(
			bestTeam.Forwards,
			fmt.Sprintf("[%v] %s (%s)", forwards[i].Player.Form, forwards[i].Player.Name, forwards[i].OpposingTeam.ShortName),
		)
	}

	gameweek := data.Gameweek(gameWeekInt)

	printOutput(bestTeam, gameweek)
}

func printOutput(bestTeam bestTeam, gameweek *Gameweek) {
	fmt.Printf("The best team you could play in %s (deadline %s) is: \n", gameweek.Name, gameweek.Deadline)
	fmt.Printf("-- [Form] Name (Opponent) -- \n\n")

	fmt.Println("Goalkeepers:")
	for _, goalkeeper := range bestTeam.Goalkeepers {
		fmt.Println(goalkeeper)
	}

	fmt.Println("\nDefenders:")
	for _, defender := range bestTeam.Defenders {
		fmt.Println(defender)
	}

	fmt.Println("\nMidfielders:")
	for _, midfielder := range bestTeam.Midfielders {
		fmt.Println(midfielder)
	}

	fmt.Println("\nForwards:")
	for _, forward := range bestTeam.Forwards {
		fmt.Println(forward)
	}
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
