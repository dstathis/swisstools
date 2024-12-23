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

type Player struct {
	name   string
	points int
	notes  []string
}

type Pairing struct {
	playera     int
	playerb     int
	playeraWins int
	playerbWins int
	draws       int
}

type Round = []Pairing

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
	players := []int{}
	for id, _ := range t.players {
		players = append(players, id)
	}
	for len(players) > 0 {
		if len(players) == 1 {
			t.rounds[t.currentRound] = append(t.rounds[t.currentRound], Pairing{playera: players[0], playerb: -1, playeraWins: 2, playerbWins: 0, draws: 0})
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
			t.rounds[t.currentRound] = append(t.rounds[t.currentRound], Pairing{playera: player0, playerb: player1, playeraWins: -1, playerbWins: -1, draws: -1})
		}
	}
}

func (t *Tournament) AddResult(id int, wins int, losses int, draws int) error {
	for i, pairing := range t.rounds[t.currentRound] {
		if pairing.playera == id {
			t.rounds[t.currentRound][i].playeraWins = wins
			t.rounds[t.currentRound][i].playerbWins = losses
			t.rounds[t.currentRound][i].draws = draws
			return nil
		}
		if pairing.playerb == id {
			t.rounds[t.currentRound][i].playerbWins = wins
			t.rounds[t.currentRound][i].playeraWins = losses
			t.rounds[t.currentRound][i].draws = draws
			return nil
		}
	}
	return errors.New("player not found")
}

func (t *Tournament) GetRound() []Pairing {
	return t.rounds[t.currentRound]
}
