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
		}
	}
}

func (t *Tournament) GetPairings() [][2]int {
	return t.rounds[t.currentRound].pairings
}

type Round struct {
	pairings [][2]int
	results  []bool // Boolean values here represent the index of the player that won.
}

type Player struct {
	name   string
	points int
	notes  []string
}
