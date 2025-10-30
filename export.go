package swisstools

import (
	"encoding/json"
	"sort"
)

// Export format versioning (semantic versioning)
const exportVersion = "1.0.0"

// tournamentExport is the JSON schema for serializing a Tournament.
// Keep this separate from internal structs to avoid leaking private fields
// and to maintain backward compatibility via versioned schemas.
type tournamentExport struct {
	Version      string            `json:"version"`
	Config       TournamentConfig  `json:"config"`
	LastID       int               `json:"lastId"`
	CurrentRound int               `json:"currentRound"`
	Started      bool              `json:"started"`
	Finished     bool              `json:"finished"`
	Players      []playerExport    `json:"players"`
	Rounds       [][]pairingExport `json:"rounds"`
}

type playerExport struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Points         int       `json:"points"`
	Wins           int       `json:"wins"`
	Losses         int       `json:"losses"`
	Draws          int       `json:"draws"`
	GameWins       int       `json:"gameWins"`
	GameLosses     int       `json:"gameLosses"`
	GameDraws      int       `json:"gameDraws"`
	Notes          []string  `json:"notes"`
	Removed        bool      `json:"removed"`
	RemovedInRound int       `json:"removedInRound"`
	ExternalID     *int      `json:"externalID,omitempty"`
	Decklist       *Decklist `json:"decklist,omitempty"`
}

type pairingExport struct {
	PlayerA     int `json:"playerA"`
	PlayerB     int `json:"playerB"`
	PlayerAWins int `json:"playerAWins"`
	PlayerBWins int `json:"playerBWins"`
	Draws       int `json:"draws"`
}

// DumpTournament returns the tournament state serialized as JSON.
//
// Returns:
//   - []byte: JSON-encoded snapshot of the tournament
//   - error: non-nil if serialization fails
func (t *Tournament) DumpTournament() ([]byte, error) {
	// Serialize players in a stable order by ID
	playerIDs := make([]int, 0, len(t.players))
	for id := range t.players {
		playerIDs = append(playerIDs, id)
	}
	// Sort ascending by ID for deterministic output
	sort.Ints(playerIDs)

	players := make([]playerExport, 0, len(playerIDs))
	for _, id := range playerIDs {
		p := t.players[id]
		players = append(players, playerExport{
			ID:             id,
			Name:           p.name,
			Points:         p.points,
			Wins:           p.wins,
			Losses:         p.losses,
			Draws:          p.draws,
			GameWins:       p.gameWins,
			GameLosses:     p.gameLosses,
			GameDraws:      p.gameDraws,
			Notes:          append([]string(nil), p.notes...),
			Removed:        p.removed,
			RemovedInRound: p.removedInRound,
			ExternalID:     p.externalID,
			Decklist:       p.decklist,
		})
	}

	// Serialize rounds
	rounds := make([][]pairingExport, 0, len(t.rounds))
	for _, r := range t.rounds {
		out := make([]pairingExport, 0, len(r))
		for _, pr := range r {
			out = append(out, pairingExport{
				PlayerA:     pr.playera,
				PlayerB:     pr.playerb,
				PlayerAWins: pr.playeraWins,
				PlayerBWins: pr.playerbWins,
				Draws:       pr.draws,
			})
		}
		rounds = append(rounds, out)
	}

	payload := tournamentExport{
		Version:      exportVersion,
		Config:       t.config,
		LastID:       t.lastId,
		CurrentRound: t.currentRound,
		Started:      t.started,
		Finished:     t.finished,
		Players:      players,
		Rounds:       rounds,
	}

	return json.Marshal(payload)
}

// LoadTournament reconstructs a Tournament from a previously produced DumpTournament payload.
//
// Inputs:
//   - data: JSON-encoded tournament snapshot from DumpTournament
//
// Returns:
//   - Tournament: reconstructed tournament
//   - error: non-nil if the payload cannot be decoded
func LoadTournament(data []byte) (Tournament, error) {
	var payload tournamentExport
	if err := json.Unmarshal(data, &payload); err != nil {
		return Tournament{}, err
	}

	// Rebuild tournament
	t := Tournament{}
	t.config = payload.Config
	t.lastId = payload.LastID
	t.players = map[int]Player{}
	t.currentRound = payload.CurrentRound
	t.started = payload.Started
	t.finished = payload.Finished

	// Players
	for _, pe := range payload.Players {
		p := Player{
			name:           pe.Name,
			points:         pe.Points,
			wins:           pe.Wins,
			losses:         pe.Losses,
			draws:          pe.Draws,
			gameWins:       pe.GameWins,
			gameLosses:     pe.GameLosses,
			gameDraws:      pe.GameDraws,
			notes:          append([]string(nil), pe.Notes...),
			removed:        pe.Removed,
			removedInRound: pe.RemovedInRound,
			externalID:     pe.ExternalID,
			decklist:       pe.Decklist,
		}
		t.players[pe.ID] = p
	}

	// Rounds
	t.rounds = make([]Round, len(payload.Rounds))
	for i, r := range payload.Rounds {
		row := make([]Pairing, 0, len(r))
		for _, pr := range r {
			row = append(row, Pairing{
				playera:     pr.PlayerA,
				playerb:     pr.PlayerB,
				playeraWins: pr.PlayerAWins,
				playerbWins: pr.PlayerBWins,
				draws:       pr.Draws,
			})
		}
		t.rounds[i] = row
	}

	return t, nil
}
