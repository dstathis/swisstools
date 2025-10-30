// Package swisstools provides utilities for running Swiss-system tournaments.
//
// Core capabilities include:
//   - Proper Swiss pairing across rounds with byes and late entries
//   - Recording match results at the game level with configurable points
//   - Computing standings with standard tiebreakers
//   - Dropping/removing players while preserving their history
//   - Optional player metadata (external ID and structured decklist)
//   - Versioned JSON dump/load to persist and resume tournaments
//
// Quick start:
//
//	t := swisstools.NewTournament()
//	t.AddPlayer("Alice")
//	t.AddPlayer("Bob")
//	_ = t.StartTournament()
//	for _, p := range t.GetRound() {
//		if p.PlayerB() == swisstools.BYE_OPPONENT_ID { continue }
//		_ = t.AddResult(p.PlayerA(), 2, 1, 0)
//	}
//	_ = t.UpdatePlayerStandings()
//
// See README for a longer example and usage notes.
package swisstools
