package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rodaine/table"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var cache = make(map[string]interface{}, 0)

type StartingPlayer struct {
	Player       Player
	Fixture      Fixture
	OpposingTeam Team
	OverallRank  string
	TypeRank     string
}

func (sp StartingPlayer) Score() float32 {
	cacheKey := fmt.Sprintf("score_player_%d", sp.Player.ID)
	if val, exists := cache[cacheKey]; exists {
		return val.(float32)
	}

	chanceOfPlaying, ok := sp.Player.ChanceOfPlaying[sp.Fixture.Gameweek.ID]
	if !ok {
		chanceOfPlaying = 1
	}

	// i'm thinking that this prevents multiplying by 0 and by 1 has no effect anyway
	difficultyMajority := float32(sp.Fixture.DifficultyMajority + 1)

	score := sp.Player.Form *
		sp.Player.Stats.ICTIndex *
		difficultyMajority *
		sp.Player.Stats.AverageStarts *
		chanceOfPlaying

	cache[cacheKey] = score

	return score
}

func (sp StartingPlayer) WeightedPointsAverage() float32 {
	cacheKey := fmt.Sprintf("wppg_player_%d", sp.Player.ID)
	if val, exists := cache[cacheKey]; exists {
		return val.(float32)
	}

	// we load this here because it's very slow to make this request for all players
	playerHistory, _ := requestPlayerHistory(int(sp.Player.ID))

	// get all matches with similar difficulty majority
	teamFixtures := sp.Player.Team.Fixtures
	similarTeamFixtures := make(map[FixtureID]bool, 0)
	for _, fixture := range teamFixtures {
		if fixture.DifficultyMajority == sp.Fixture.DifficultyMajority {
			similarTeamFixtures[fixture.ID] = true
		}
	}

	if len(similarTeamFixtures) == 0 {
		return sp.Player.PointsPerGame
	}

	totalPoints := 0
	similarFixturesPlayerPlayedIn := 0
	for fixtureID, fixture := range playerHistory {
		if _, ok := similarTeamFixtures[fixtureID]; ok {
			totalPoints += fixture.Points
			similarFixturesPlayerPlayedIn++
		}
	}

	returnVal := float32(totalPoints) / float32(similarFixturesPlayerPlayedIn)

	cache[cacheKey] = returnVal

	return returnVal
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

type BestTeam struct {
	Goalkeepers []StartingPlayer
	Defenders   []StartingPlayer
	Midfielders []StartingPlayer
	Forwards    []StartingPlayer
}

func (bt *BestTeam) PlayerCount() int {
	return len(bt.Goalkeepers) +
		len(bt.Defenders) +
		len(bt.Midfielders) +
		len(bt.Forwards)
}

type TeamConfig struct {
	Players   []string `json:"players"`
	BankValue float32  `json:"bank_value"`
}

func main() {
	playerName := flag.String("player", "", "for specifying a player's name")
	gameWeekInt := flag.Int("gameweek", 0, "for specifying the gameweek")
	teamConfig := flag.String("config", "", "for specifying your current team config")
	flag.Parse()

	if *gameWeekInt == 0 {
		panic("You must provide a gameweek number")
	}

	data, err := BuildData()
	if err != nil {
		panic(err)
	}

	var gameweek *Gameweek
	if *gameWeekInt > 0 {
		gameweek = data.Gameweek(*gameWeekInt)
		if gameweek.Finished {
			fmt.Printf("\n%s is finished\n\n", gameweek.Name)
			return
		}
	} else {
		gameweek = data.CurrentGameweek()
	}

	previousGameweek := data.Gameweek(int(gameweek.ID) - 1)

	// set used in case multiple fixtures in one gameweek for a team
	likelyWinnerPlayersSet := make(map[PlayerID]StartingPlayer, 0)

	for _, fixture := range data.FixturesByGameWeek(*gameWeekInt) {
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

	rankedStartingPlayers := rankPlayers(likelyWinnerPlayers)

	if *playerName != "" {
		var matchingPlayer StartingPlayer
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		players := rankPlayers(data.GameweekPlayers(*gameWeekInt))
		for _, player := range players {
			flatString, _, _ := transform.String(t, player.Player.Name)
			if fuzzy.Match(*playerName, flatString) || fuzzy.Match(*playerName, player.Player.Name) {
				matchingPlayer = player
				break
			}
		}
		if matchingPlayer.Player.Name == "" {
			fmt.Printf("player '%s' not found\n", *playerName)
			return
		}
		fmt.Printf("Player: %s, Type: %s\n", matchingPlayer.Player.Name, matchingPlayer.Player.Type.Name)
		fmt.Printf("Team: %s\n", matchingPlayer.Player.Team.Name)
		fmt.Printf("Cost: %s\n", matchingPlayer.Player.Cost)
		fmt.Printf("Form: %.2f\n", matchingPlayer.Player.Form)
		fmt.Printf("Score: %.0f\n", matchingPlayer.Score())
		fmt.Printf("Picked: %.0f%%\n", matchingPlayer.Player.PickedPercentage)
		fmt.Printf("Overall Rank: %s, by Type: %s\n", matchingPlayer.OverallRank, matchingPlayer.TypeRank)
		fmt.Printf("Opposition: %s\n", matchingPlayer.OpposingTeam.Name)
		return
	}

	differentials := differentialPlayers(rankedStartingPlayers)
	bestTeam := createHighestScoringTeam(rankedStartingPlayers)

	if *teamConfig != "" {
		configFilePath, err := filepath.Abs(*teamConfig)
		if err != nil {
			panic(err)
		}
		jsonFile, err := os.Open(configFilePath)
		if err != nil {
			panic(err)
		}
		byteValue, err := io.ReadAll(jsonFile)
		if err != nil {
			panic(err)
		}
		var config TeamConfig
		err = json.Unmarshal(byteValue, &config)
		if err != nil {
			panic(err)
		}

		if len(config.Players) < 15 {
			panic(fmt.Errorf("team config not valid: you only have %d players but need 14", len(config.Players)))
		}

		gameweekPlayers := data.GameweekPlayers(*gameWeekInt)

		myGameweekPlayers := make([]StartingPlayer, 0)
		for _, myPlayer := range config.Players {
			for _, gameweekPlayer := range gameweekPlayers {
				if myPlayer == gameweekPlayer.Player.Name {
					myGameweekPlayers = append(myGameweekPlayers, gameweekPlayer)
				}
			}
		}

		myGameweekPlayers = sortStartingPlayersByScore(myGameweekPlayers)
		// checking again for consistency
		if len(myGameweekPlayers) < 15 {
			panic(fmt.Errorf("team config not valid: you only have %d players but need 14", len(myGameweekPlayers)))
		}

		bestTeam := createHighestScoringTeam(myGameweekPlayers)
		fmt.Printf("\nWith your current players, the best team you could pick for %s is:\n", gameweek.Name)
		headerFmt, columnFmt := tableFormat()
		tbl := table.New("Type", "Name", "Form", "Score", "Picked", "Cost", "Opponent")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
		appendOptions := AppendOptions{withPickedPercentage: true}
		appendToTable(tbl, bestTeam.Goalkeepers, appendOptions)
		appendToTable(tbl, bestTeam.Defenders, appendOptions)
		appendToTable(tbl, bestTeam.Midfielders, appendOptions)
		appendToTable(tbl, bestTeam.Forwards, appendOptions)
		tbl.Print()

		worstPlayer := myGameweekPlayers[14]
		cashAfterSale := worstPlayer.Player.RawCost + config.BankValue

		playersICanAfford := make([]StartingPlayer, 0)
		for _, potential := range gameweekPlayers {
			if potential.Player.RawCost <= cashAfterSale && potential.Player.Type.ID == worstPlayer.Player.Type.ID {
				playersICanAfford = append(playersICanAfford, potential)
			}
		}

		playersICanAfford = sortStartingPlayersByScore(playersICanAfford)
		topPick := playersICanAfford[0]

		fmt.Printf(
			"\nYou might want to consider selling %s and buying %s, who costs %s and has a score of %.0f.\n\n",
			worstPlayer.Player.Name,
			topPick.Player.Name,
			topPick.Player.Cost,
			topPick.Score(),
		)

		fmt.Printf("Type './simple-fantasy -gameweek %d -player %s' to find out more about him.\n\n", *gameWeekInt, topPick.Player.Name)

		// what could 2 transfers get you?
		secondWorstPlayer := myGameweekPlayers[13]
		cashAfterSale = worstPlayer.Player.RawCost + secondWorstPlayer.Player.RawCost + config.BankValue
		scoresAndPlayers := make(map[float32][]StartingPlayer, 0)
		sortedGameweekPlayers := sortStartingPlayersByScore(gameweekPlayers)
		for _, potentialFirstTransfer := range gameweekPlayers {
			if potentialFirstTransfer.Player.Type.ID != worstPlayer.Player.Type.ID && potentialFirstTransfer.Player.Type.ID != secondWorstPlayer.Player.Type.ID {
				continue
			}
			cashNow := cashAfterSale - potentialFirstTransfer.Player.RawCost
			var potentialSecondTransferType PlayerTypeID
			if potentialFirstTransfer.Player.Type.ID == worstPlayer.Player.Type.ID {
				potentialSecondTransferType = secondWorstPlayer.Player.Type.ID
			} else if potentialFirstTransfer.Player.Type.ID == secondWorstPlayer.Player.Type.ID {
				potentialSecondTransferType = worstPlayer.Player.Type.ID
			}
			for _, potentialSecondTransfer := range sortedGameweekPlayers {
				if (cashNow-potentialSecondTransfer.Player.RawCost) >= 0 &&
					potentialSecondTransfer.Player.ID != potentialFirstTransfer.Player.ID &&
					potentialSecondTransfer.Player.Type.ID == potentialSecondTransferType {
					combinedScore := potentialFirstTransfer.Score() + potentialSecondTransfer.Score()
					scoresAndPlayers[combinedScore] = append(
						scoresAndPlayers[combinedScore],
						[]StartingPlayer{potentialFirstTransfer, potentialSecondTransfer}...,
					)
				}
			}
		}

		scoreKeys := make([]float32, 0)
		for key := range scoresAndPlayers {
			scoreKeys = append(scoreKeys, key)
		}
		sort.Slice(scoreKeys, func(i, j int) bool {
			return scoreKeys[i] > scoreKeys[j]
		})
		bestPair := scoresAndPlayers[scoreKeys[0]]
		if len(bestPair) > 1 {
			formattedCash := fmt.Sprintf("£%.1fm", float32(cashAfterSale))
			fmt.Printf(
				"Or if you were willing to make two transfers you could sell %s and %s for %s and buy %s and %s, costing %s and %s, with scores %.0f and %.0f.\n\n",
				worstPlayer.Player.Name,
				secondWorstPlayer.Player.Name,
				formattedCash,
				bestPair[0].Player.Name,
				bestPair[1].Player.Name,
				bestPair[0].Player.Cost,
				bestPair[1].Player.Cost,
				bestPair[0].Score(),
				bestPair[1].Score(),
			)
		}

		fmt.Printf("(Scores may vary where team expected to draw.)\n\n")

		return
	}

	printOutput(bestTeam, differentials, gameweek)
}

func rankPlayers(players []StartingPlayer) []StartingPlayer {
	players = sortStartingPlayersByScore(players)
	rankedPlayers := make([]StartingPlayer, 0)
	overallRanking := 0
	typeRankings := make(map[PlayerTypeID]int, 0)

	for _, player := range players {
		overallRanking++
		typeRankings[player.Player.Type.ID]++

		player.OverallRank = ordinalNumber(overallRanking)
		player.TypeRank = ordinalNumber(typeRankings[player.Player.Type.ID])
		rankedPlayers = append(rankedPlayers, player)
	}

	return rankedPlayers
}

func createHighestScoringTeam(startingPlayers []StartingPlayer) BestTeam {
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
		for _, player := range startingPlayers {
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

	var bestTeam BestTeam
	bestTeam.Goalkeepers = highestScoringTeam["Goalkeeper"]
	bestTeam.Defenders = highestScoringTeam["Defender"]
	bestTeam.Midfielders = highestScoringTeam["Midfielder"]
	bestTeam.Forwards = highestScoringTeam["Forward"]

	return bestTeam
}

func tableFormat() (table.Formatter, table.Formatter) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	return headerFmt, columnFmt
}

func printOutput(bestTeam BestTeam, differentials BestTeam, gameweek *Gameweek) {
	headerFmt, columnFmt := tableFormat()

	tbl := table.New("Type", "Name", "Form", "PPG", "WPPG", "Score", "Rank (Type)", "Cost", "Opponent")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	if gameweek.IsCurrent {
		fmt.Printf("\nThe best team you could have played going into the current gameweek (deadline %s) was: \n", gameweek.Deadline)
	} else {
		fmt.Printf("\nThe best team you can play in %s (deadline %s) is: \n", gameweek.Name, gameweek.Deadline)
	}
	appendOptions := AppendOptions{withPickedPercentage: false, withRank: true}
	appendToTable(tbl, bestTeam.Goalkeepers, appendOptions)
	appendToTable(tbl, bestTeam.Defenders, appendOptions)
	appendToTable(tbl, bestTeam.Midfielders, appendOptions)
	appendToTable(tbl, bestTeam.Forwards, appendOptions)
	tbl.Print()

	fmt.Printf("\nDifferentials:\n")
	differentialsTbl := table.New("Type", "Name", "Form", "PPG", "WPPG", "Score", "Picked", "Rank (Type)", "Cost", "Opponent")
	differentialsTbl.
		WithHeaderFormatter(headerFmt).
		WithFirstColumnFormatter(columnFmt)
	appendOptions = AppendOptions{withPickedPercentage: true, withRank: true}
	appendToTable(differentialsTbl, differentials.Goalkeepers, appendOptions)
	appendToTable(differentialsTbl, differentials.Defenders, appendOptions)
	appendToTable(differentialsTbl, differentials.Midfielders, appendOptions)
	appendToTable(differentialsTbl, differentials.Forwards, appendOptions)
	differentialsTbl.Print()

	fmt.Println()

	playersToBuyNow := compareBestTeams(bestTeam, differentials)
	if playersToBuyNow.PlayerCount() > 0 && gameweek.IsNext {
		fmt.Println("Buy these players now!")
		playersToBuyTbl := table.New("Type", "Name", "Form", "PPG", "WPPG", "Score", "Picked", "Rank (Type)", "Cost", "Opponent")
		playersToBuyTbl.
			WithHeaderFormatter(headerFmt).
			WithFirstColumnFormatter(columnFmt)
		appendOptions = AppendOptions{withPickedPercentage: true, withRank: true}
		appendToTable(playersToBuyTbl, playersToBuyNow.Goalkeepers, appendOptions)
		appendToTable(playersToBuyTbl, playersToBuyNow.Defenders, appendOptions)
		appendToTable(playersToBuyTbl, playersToBuyNow.Midfielders, appendOptions)
		appendToTable(playersToBuyTbl, playersToBuyNow.Forwards, appendOptions)
		playersToBuyTbl.Print()
	}

	fmt.Println()

	fmt.Println("(PPG = Points Per Game, WPPG = Weighted Points Per Game (by match difficulty))")

	fmt.Println()
}

type AppendOptions struct {
	withPickedPercentage bool
	withRank             bool
}

func appendToTable(tbl table.Table, fixtureWinners []StartingPlayer, options AppendOptions) {
	for _, fixtureWinner := range fixtureWinners {
		playerName := fixtureWinner.Player.Name

		if fixtureWinner.Player.MostCaptained {
			playerName += " (C)"
		}

		row := []interface{}{
			fixtureWinner.Player.Type.ShortName,
			playerName,
			fixtureWinner.Player.Form,
			fmt.Sprintf("%.2f", fixtureWinner.Player.PointsPerGame),
			fmt.Sprintf("%.2f", fixtureWinner.WeightedPointsAverage()),
			fmt.Sprintf("%.0f", fixtureWinner.Score()),
		}

		if options.withPickedPercentage {
			row = append(row, fmt.Sprintf("%.0f%%", fixtureWinner.Player.PickedPercentage))
		}

		if options.withRank {
			row = append(row, fmt.Sprintf("%s (%s)", fixtureWinner.OverallRank, fixtureWinner.TypeRank))
		}

		row = append(row, []interface{}{
			fixtureWinner.Player.Cost,
			fixtureWinner.OpposingTeam.Name,
		}...)

		tbl.AddRow(row...)
	}
}

func sortStartingPlayersByScore(startingPlayers []StartingPlayer) []StartingPlayer {
	newSlice := make([]StartingPlayer, len(startingPlayers))
	copy(newSlice, startingPlayers)
	sort.Slice(newSlice, func(i, j int) bool {
		return newSlice[i].Score() > newSlice[j].Score()
	})
	return newSlice
}

func differentialPlayers(startingPlayers []StartingPlayer) BestTeam {
	players := make([]StartingPlayer, 0)
	for _, startingPlayer := range startingPlayers {
		if startingPlayer.Player.PickedPercentage < 15 {
			players = append(players, startingPlayer)
		}
	}
	return createHighestScoringTeam(players)
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

// this is sloppy and slow, but i'll refine it later
func compareBestTeams(a BestTeam, b BestTeam) BestTeam {
	var bestTeam BestTeam

	for _, aPlayer := range a.Goalkeepers {
		for _, bPlayer := range b.Goalkeepers {
			if aPlayer.Player.ID == bPlayer.Player.ID {
				bestTeam.Goalkeepers = append(bestTeam.Goalkeepers, aPlayer)
			}
		}
	}

	for _, aPlayer := range a.Defenders {
		for _, bPlayer := range b.Defenders {
			if aPlayer.Player.ID == bPlayer.Player.ID {
				bestTeam.Defenders = append(bestTeam.Defenders, aPlayer)
			}
		}
	}

	for _, aPlayer := range a.Midfielders {
		for _, bPlayer := range b.Midfielders {
			if aPlayer.Player.ID == bPlayer.Player.ID {
				bestTeam.Midfielders = append(bestTeam.Midfielders, aPlayer)
			}
		}
	}

	for _, aPlayer := range a.Forwards {
		for _, bPlayer := range b.Forwards {
			if aPlayer.Player.ID == bPlayer.Player.ID {
				bestTeam.Forwards = append(bestTeam.Forwards, aPlayer)
			}
		}
	}

	return bestTeam
}
