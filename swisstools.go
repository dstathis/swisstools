package swisstools

import (
	"errors"
	"io"
	"math/rand"
	"time"

	"github.com/olekukonko/tablewriter"
)

type Tournament struct {
	lastId       int // Most recent player id to be assigned.
	players      map[int]Player
	currentRound int
	rounds       []Round
}

func NewTournament() Tournament {
	rand.Seed(time.Now().Unix())
	tournament := Tournament{}
	tournament.lastId = 0
	tournament.players = map[int]Player{}
	tournament.currentRound = 1 // Index round starting with 1 to make the round numbers human readable.
	tournament.rounds = make([]Round, 2)
	return tournament
}

func (t *Tournament) AddPlayer(name string) error {
	if name == "" {
		return errors.New("empty name")
	}
	t.lastId++
	player := Player{}
	player.points = 0
	player.name = name
	player.notes = []string{}
	t.players[t.lastId] = player
	return nil
}

func (t *Tournament) FormatPlayers(w io.Writer) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Name", "Wins", "Losses", "Points"})
	for _, player := range t.players {
		table.Append([]string{player.name, "0", "0", "0"})
	}
	table.Render()
}

func (t *Tournament) NextRound() {
	t.currentRound++
}

func (t *Tournament) Pair() {
	round := &t.rounds[t.currentRound]
	players := []int{}
	for id, _ := range t.players {
		players = append(players, id)
	}
	for len(players) > 0 {
		if len(players) == 1 {
			round.pairings = append(round.pairings, [2]int{players[0], -1})
			round.results = append(round.results, [3]int{2, 0, 0})
			players = players[:0]
		} else {
			// Choose 2 random players and delete them from the list.
			playerIndex := rand.Intn(len(players))
			player0 := players[playerIndex]
			players[playerIndex] = players[len(players)-1]
			players = players[:len(players)-1]
			playerIndex = rand.Intn(len(players))
			player1 := players[playerIndex]
			players[playerIndex] = players[len(players)-1]
			players = players[:len(players)-1]
			round.pairings = append(round.pairings, [2]int{player0, player1})
			round.results = append(round.results, [3]int{-1, -1, -1})
		}
	}
}

func (t *Tournament) AddResult(id int, wins int, losses int, draws int) error {
	round := t.rounds[t.currentRound]
	for i, players := range round.pairings {
		if players[0] == id {
			round.results[i][0] = wins
			round.results[i][1] = losses
			round.results[i][2] = draws
			return nil
		}
		if players[1] == id {
			round.results[i][1] = wins
			round.results[i][0] = losses
			round.results[i][2] = draws
			return nil
		}
	}
	return errors.New("player not found")
}

func (t *Tournament) GetPairings() [][2]int {
	return t.rounds[t.currentRound].pairings
}

func (t *Tournament) GetResults() [][3]int {
	return t.rounds[t.currentRound].results
}

type Round struct {
	pairings [][2]int
	results  [][3]int // [<first player wins>, <second player wins>, <draws>]
}

type Player struct {
	name   string
	points int
	notes  []string
}
