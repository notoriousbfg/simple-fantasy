package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

type StartingPlayer struct {
	Player       Player
	Fixture      Fixture
	OpposingTeam Team
}

func (sp StartingPlayer) Score() float32 {
	chanceOfPlaying, ok := sp.Player.ChanceOfPlaying[sp.Fixture.Gameweek.ID]
	if !ok {
		chanceOfPlaying = 1
	}

	return sp.Player.Form *
		sp.Player.Stats.ICTIndex *
		float32(sp.Fixture.DifficultyMajority) *
		chanceOfPlaying
}

type StartingEleven map[string][]StartingPlayer

func (se StartingEleven) PlayerCount() int {
	count := 0
	for position := range se {
		for range position {
			count++
		}
	}
	return count
}

func (se StartingEleven) Score() float32 {
	score := float32(0)
	for _, players := range se {
		for _, player := range players {
			score += player.Score()
		}
	}
	return score
}

type bestTeam struct {
	Goalkeepers []StartingPlayer
	Defenders   []StartingPlayer
	Midfielders []StartingPlayer
	Forwards    []StartingPlayer
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

	var gameweek *Gameweek

	if gameWeekInt > 0 {
		gameweek = data.Gameweek(gameWeekInt)
	} else {
		gameweek = data.CurrentGameweek()
	}

	previousGameweek := data.Gameweek(int(gameweek.ID) - 1)

	// set used in case multiple fixtures in one gameweek for a team
	likelyWinnerPlayersSet := make(map[PlayerID]StartingPlayer, 0)
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

			// player already exists
			likelyWinnerPlayersSet[player.ID] = StartingPlayer{
				Player:       player,
				Fixture:      fixture,
				OpposingTeam: opposingTeam,
			}
		}
	}

	likelyWinnerPlayers := make([]StartingPlayer, 0)
	for _, player := range likelyWinnerPlayersSet {
		likelyWinnerPlayers = append(likelyWinnerPlayers, player)
	}

	sortStartingPlayers(likelyWinnerPlayers)

	positionCountCombinations := [][]int{
		{1, 3, 5, 2},
		{1, 4, 4, 2},
		{1, 5, 3, 2},
		{1, 3, 4, 3},
		{1, 4, 3, 3},
		{1, 5, 4, 1},
		{1, 5, 2, 3},
	}
	positionVariations := make(map[string]StartingEleven)

	for _, combination := range positionCountCombinations {
		key := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(combination)), "-"), "[]") // e.g. 1-3-5-2

		startingEleven := StartingEleven{}

		// assumes players are in descending score order
		for _, player := range likelyWinnerPlayers {
			position := player.Player.Type.Name
			switch position {
			case "Goalkeeper":
				if len(startingEleven[position]) < combination[0] {
					startingEleven[position] = append(startingEleven[position], player)
				}
			case "Defender":
				if len(startingEleven[position]) < combination[1] {
					startingEleven[position] = append(startingEleven[position], player)
				}
			case "Midfielder":
				if len(startingEleven[position]) < combination[2] {
					startingEleven[position] = append(startingEleven[position], player)
				}
			case "Forward":
				if len(startingEleven[position]) < combination[3] {
					startingEleven[position] = append(startingEleven[position], player)
				}
			}
		}

		positionVariations[key] = startingEleven
	}

	var highestScore float32
	var highestScoringTeam StartingEleven
	for _, startingEleven := range positionVariations {
		seScore := startingEleven.Score()
		if seScore > highestScore {
			highestScore = seScore
			highestScoringTeam = startingEleven
		}
	}

	var bestTeam bestTeam
	bestTeam.Goalkeepers = highestScoringTeam["Goalkeeper"]
	bestTeam.Defenders = highestScoringTeam["Defender"]
	bestTeam.Midfielders = highestScoringTeam["Midfielder"]
	bestTeam.Forwards = highestScoringTeam["Forward"]

	printOutput(bestTeam, gameweek)
}

func printOutput(bestTeam bestTeam, gameweek *Gameweek) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Type", "Name", "Form", "Score", "Cost", "Opponent")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	if gameweek.IsCurrent {
		fmt.Printf("\nThe best team you could have played going into the current gameweek (deadline %s) was: \n", gameweek.Deadline)
	} else {
		fmt.Printf("\nThe best team you can play in %s (deadline %s) is: \n", gameweek.Name, gameweek.Deadline)
	}
	appendToTable(tbl, bestTeam.Goalkeepers)
	appendToTable(tbl, bestTeam.Defenders)
	appendToTable(tbl, bestTeam.Midfielders)
	appendToTable(tbl, bestTeam.Forwards)
	tbl.Print()
}

func appendToTable(tbl table.Table, fixtureWinners []StartingPlayer) {
	for _, fixtureWinner := range fixtureWinners {
		playerName := fixtureWinner.Player.Name

		if fixtureWinner.Player.MostCaptained {
			playerName += " (C)"
		}

		tbl.AddRow(
			fixtureWinner.Player.Type.ShortName,
			playerName,
			fixtureWinner.Player.Form,
			fixtureWinner.Score(),
			fixtureWinner.Player.Cost,
			fixtureWinner.OpposingTeam.Name,
		)
	}
}

func sortStartingPlayers(startingPlayers []StartingPlayer) {
	sort.Slice(startingPlayers, func(i, j int) bool {
		return startingPlayers[i].Score() > startingPlayers[j].Score()
	})
}
