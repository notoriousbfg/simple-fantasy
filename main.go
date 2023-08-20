package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

type likelyFixtureWinner struct {
	Player       Player
	Fixture      Fixture
	OpposingTeam Team
}

type bestTeam struct {
	Goalkeepers []likelyFixtureWinner
	Defenders   []likelyFixtureWinner
	Midfielders []likelyFixtureWinner
	Forwards    []likelyFixtureWinner
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

	gameweek := data.Gameweek(gameWeekInt)
	previousGameweek := data.Gameweek(gameWeekInt - 1)

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
			if previousGameweek != nil {
				player.MostCaptained = (previousGameweek.MostCaptainedID == player.ID)
			}

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
			goalkeepers[i],
		)
	}

	defType := data.PlayerType("Defender")
	defenders := filterWinnersByType(likelyWinners, "Defender")
	for i := 0; i < defType.TeamPlayerCount; i++ {
		bestTeam.Defenders = append(
			bestTeam.Defenders,
			defenders[i],
		)
	}

	midType := data.PlayerType("Midfielder")
	midfielders := filterWinnersByType(likelyWinners, "Midfielder")
	for i := 0; i < midType.TeamPlayerCount; i++ {
		bestTeam.Midfielders = append(
			bestTeam.Midfielders,
			midfielders[i],
		)
	}

	fwdType := data.PlayerType("Forward")
	forwards := filterWinnersByType(likelyWinners, "Forward")
	for i := 0; i < fwdType.TeamPlayerCount; i++ {
		bestTeam.Forwards = append(
			bestTeam.Forwards,
			forwards[i],
		)
	}
	printOutput(bestTeam, gameweek)
}

func printOutput(bestTeam bestTeam, gameweek *Gameweek) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Type", "Name", "Form", "Opponent")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	fmt.Printf("\nThe best team you could play in %s (deadline %s) is: \n", gameweek.Name, gameweek.Deadline)
	appendToTable(tbl, bestTeam.Goalkeepers)
	appendToTable(tbl, bestTeam.Defenders)
	appendToTable(tbl, bestTeam.Midfielders)
	appendToTable(tbl, bestTeam.Forwards)
	tbl.Print()
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

func appendToTable(tbl table.Table, fixtureWinners []likelyFixtureWinner) {
	for _, fixtureWinner := range fixtureWinners {
		playerName := fixtureWinner.Player.Name

		if fixtureWinner.Player.MostCaptained {
			playerName += " (C)"
		}

		tbl.AddRow(
			fixtureWinner.Player.Type.ShortName,
			playerName,
			fixtureWinner.Player.Form,
			fixtureWinner.OpposingTeam.Name,
		)
	}
}
