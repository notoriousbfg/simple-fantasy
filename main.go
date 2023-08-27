package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type StartingPlayer struct {
	Player       Player
	Fixture      Fixture
	OpposingTeam Team
	OverallRank  string
	TypeRank     string
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

	var playerName string
	if len(args) > 1 {
		playerName = args[1]
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

	rankedStartingPlayers := make([]StartingPlayer, 0)

	overallRanking := 0
	typeRankings := make(map[PlayerTypeID]int, 0)

	for _, player := range likelyWinnerPlayers {
		overallRanking++
		typeRankings[player.Player.Type.ID]++

		player.OverallRank = ordinalNumber(overallRanking)
		player.TypeRank = ordinalNumber(typeRankings[player.Player.Type.ID])
		rankedStartingPlayers = append(rankedStartingPlayers, player)
	}

	if playerName != "" {
		var matchingPlayer StartingPlayer
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		for _, player := range rankedStartingPlayers {
			// this is probably pretty sloppy
			result, _, _ := transform.String(t, player.Player.Name)
			if playerName == result {
				matchingPlayer = player
				break
			}
		}
		if matchingPlayer.Player.Name == "" {
			fmt.Printf("player '%s' not found or is not a probable winner for this gameweek\n", playerName)
			return
		}
		fmt.Printf("Player: %s, Type: %s\n", matchingPlayer.Player.Name, matchingPlayer.Player.Type.Name)
		fmt.Printf("Cost: %s\n", matchingPlayer.Player.Cost)
		fmt.Printf("Form: %.2f\n", matchingPlayer.Player.Form)
		fmt.Printf("Score: %.0f\n", matchingPlayer.Score())
		fmt.Printf("Overall Rank: %s, by Type: %s\n", matchingPlayer.OverallRank, matchingPlayer.TypeRank)
		fmt.Printf("Opposition: %s\n", matchingPlayer.OpposingTeam.Name)
		return
	}

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
		teamPlayerCounts := make(map[TeamID]int, 0)

		// assumes players are in descending score order
		for _, player := range rankedStartingPlayers {
			// you can only have 3 players from one team in your selection, continue to next ranking player
			if teamPlayerCounts[player.Player.Team.ID] >= 3 {
				continue
			}

			position := player.Player.Type.Name
			switch position {
			case "Goalkeeper":
				if len(startingEleven[position]) < combination[0] {
					startingEleven[position] = append(startingEleven[position], player)
					teamPlayerCounts[player.Player.Team.ID]++
				}
			case "Defender":
				if len(startingEleven[position]) < combination[1] {
					startingEleven[position] = append(startingEleven[position], player)
					teamPlayerCounts[player.Player.Team.ID]++
				}
			case "Midfielder":
				if len(startingEleven[position]) < combination[2] {
					startingEleven[position] = append(startingEleven[position], player)
					teamPlayerCounts[player.Player.Team.ID]++
				}
			case "Forward":
				if len(startingEleven[position]) < combination[3] {
					startingEleven[position] = append(startingEleven[position], player)
					teamPlayerCounts[player.Player.Team.ID]++
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
	tbl := table.New("Type", "Name", "Form", "Score", "Rank (Type)", "Cost", "Opponent")
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
			fmt.Sprintf("%.0f", fixtureWinner.Score()),
			fmt.Sprintf("%s (%s)", fixtureWinner.OverallRank, fixtureWinner.TypeRank),
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

func ordinalNumber(n int) string {
	if n >= 11 && n <= 13 {
		return fmt.Sprintf("%dth", n)
	}

	switch n % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}
