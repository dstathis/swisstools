package swisstools

import (
	"bytes"
	"strings"
	"testing"
)

func TestAddPlayerName(t *testing.T) {
	tournament := NewTournament()
	err := tournament.AddPlayer("Dylan")
	if err != nil || tournament.players[1].name != "Dylan" {
		t.Fatalf("Expecting player name Dylan, got %s.", tournament.players[1].name)
	}
}

func TestAddPlayerEmpty(t *testing.T) {
	tournament := NewTournament()
	err := tournament.AddPlayer("")
	if err == nil {
		t.Fatalf("No name provided but AddPlayer did not return an error.")
	}
}

func TestAddResultBadId(t *testing.T) {
	tournament := NewTournament()
	err := tournament.AddResult(5, 2, 1, 0)
	if err == nil {
		t.Fatal("Bogus player submitted result but AddResult did not return an error.")
	}
}

func TestMultipleRounds(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")
	tournament.AddPlayer("Diana")

	// Test multiple rounds to ensure dynamic growth works
	for round := 1; round <= 5; round++ {
		tournament.Pair()

		// Add results for all pairings in this round
		pairings := tournament.GetRound()
		for _, pairing := range pairings {
			if pairing.playerb != -1 {
				// Add a simple win result
				err := tournament.AddResult(pairing.playera, 2, 1, 0)
				if err != nil {
					t.Fatalf("Failed to add result in round %d: %v", round, err)
				}
			}
			// Bye rounds already have results
		}

		if round < 5 {
			err := tournament.NextRound()
			if err != nil {
				t.Fatalf("Failed to advance to round %d: %v", round+1, err)
			}
		}
	}
	t.Log("Successfully completed 5 rounds with dynamic slice growth.")
}

func TestPairRePairing(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	// First pairing should work
	tournament.Pair()
	round1 := tournament.GetRound()
	if len(round1) == 0 {
		t.Fatal("Expected pairings in round 1")
	}

	// Re-pairing the same round should work and clear previous pairings
	tournament.Pair()
	round1After := tournament.GetRound()
	if len(round1After) == 0 {
		t.Fatal("Expected pairings after re-pairing round 1")
	}

	// Add results for the pairing so we can advance to next round
	for _, pairing := range round1After {
		if pairing.playerb != -1 {
			err := tournament.AddResult(pairing.playera, 2, 1, 0)
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
		}
	}

	// Should be able to pair after advancing to next round
	err := tournament.NextRound()
	if err != nil {
		t.Fatalf("Failed to advance to next round: %v", err)
	}
	tournament.Pair()
	round2 := tournament.GetRound()
	if len(round2) == 0 {
		t.Fatal("Expected pairings in round 2")
	}

	t.Log("Re-pairing functionality works correctly.")
}

func TestAddResultWithoutPairing(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	// Try to add result without calling Pair() first
	err := tournament.AddResult(1, 2, 0, 0)
	if err == nil {
		t.Fatal("Expected error when adding result without pairing, but got none")
	}
	if err.Error() != "round has no pairings - call Pair() first" {
		t.Fatalf("Expected specific error message, got: %v", err)
	}

	// Pair the round so we can advance, but don't add results
	tournament.Pair()

	// Try to advance to next round without adding results - should fail
	err = tournament.NextRound()
	if err == nil {
		t.Fatal("Expected NextRound to fail due to incomplete matches, but it succeeded")
	}
	if err.Error() != "incomplete match found - all matches must have results" {
		t.Fatalf("Expected specific error message, got: %v", err)
	}

	// After adding results, AddResult and NextRound should work
	err = tournament.AddResult(1, 2, 0, 0)
	if err != nil {
		t.Fatalf("Expected AddResult to work after Pair(), got error: %v", err)
	}

	// Now NextRound should succeed
	err = tournament.NextRound()
	if err != nil {
		t.Fatalf("Expected NextRound to succeed after completing matches, got error: %v", err)
	}
}

func TestUpdatePlayerStandings(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")   // ID: 1
	tournament.AddPlayer("Bob")     // ID: 2
	tournament.AddPlayer("Charlie") // ID: 3

	// Pair and see what we get (random pairing)
	tournament.Pair()
	pairings := tournament.GetRound()

	// Add results for all non-bye pairings
	for _, pairing := range pairings {
		if pairing.playerb != -1 {
			// Make player A win 2-1
			err := tournament.AddResult(pairing.playera, 2, 1, 0)
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
		}
		// Bye rounds already have results set by Pair()
	}

	// Update standings for round 1
	err := tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Check total wins + losses + draws equals number of players
	// (since with 3 players, we have 1 match + 1 bye = 2 match results total)
	totalWins := 0
	totalLosses := 0
	totalDraws := 0
	totalPoints := 0

	for _, player := range tournament.players {
		totalWins += player.wins
		totalLosses += player.losses
		totalDraws += player.draws
		totalPoints += player.points
	}

	// We should have 2 wins total (1 from match, 1 from bye), 1 loss, 0 draws
	if totalWins != 2 {
		t.Errorf("Expected 2 total wins, got %d", totalWins)
	}
	if totalLosses != 1 {
		t.Errorf("Expected 1 total loss, got %d", totalLosses)
	}
	if totalDraws != 0 {
		t.Errorf("Expected 0 total draws, got %d", totalDraws)
	}
	// With new points system: 2 wins = 6 points (3 each)
	if totalPoints != 6 {
		t.Errorf("Expected 6 total points with new system (3 per win), got %d", totalPoints)
	}

	t.Log("Player standings updated correctly for round 1")
}

func TestUpdatePlayerStandingsCumulative(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice") // ID: 1
	tournament.AddPlayer("Bob")   // ID: 2

	// Round 1: Alice beats Bob
	tournament.Pair()
	err := tournament.AddResult(1, 2, 0, 0) // Alice wins 2-0
	if err != nil {
		t.Fatalf("Failed to add result: %v", err)
	}

	// Round 2: Bob beats Alice - NextRound will update standings for round 1
	err = tournament.NextRound()
	if err != nil {
		t.Fatalf("Failed to advance to next round: %v", err)
	}

	// Check round 1 stats (after NextRound updates them)
	alice := tournament.players[1]
	bob := tournament.players[2]
	if alice.wins != 1 || alice.points != 3 {
		t.Errorf("After round 1: Alice should have 1 win, 3 points, got wins=%d, points=%d", alice.wins, alice.points)
	}
	if bob.losses != 1 || bob.points != 0 {
		t.Errorf("After round 1: Bob should have 1 loss, 0 points, got losses=%d, points=%d", bob.losses, bob.points)
	}
	tournament.Pair()
	err = tournament.AddResult(2, 2, 1, 0) // Bob wins 2-1
	if err != nil {
		t.Fatalf("Failed to add result: %v", err)
	}

	// Manually update standings for round 2 (since we're not calling NextRound again)
	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Check cumulative stats after round 2
	alice = tournament.players[1]
	bob = tournament.players[2]
	if alice.wins != 1 || alice.losses != 1 || alice.points != 3 {
		t.Errorf("After round 2: Alice should have 1 win, 1 loss, 3 points, got wins=%d, losses=%d, points=%d", alice.wins, alice.losses, alice.points)
	}
	if bob.wins != 1 || bob.losses != 1 || bob.points != 3 {
		t.Errorf("After round 2: Bob should have 1 win, 1 loss, 3 points, got wins=%d, losses=%d, points=%d", bob.wins, bob.losses, bob.points)
	}

	t.Log("Cumulative player standings working correctly across multiple rounds")
}

func TestUpdatePlayerStandingsIncompleteMatch(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	// Pair but don't add results
	tournament.Pair()

	// Should error due to incomplete match
	err := tournament.UpdatePlayerStandings()
	if err == nil {
		t.Fatal("Expected error for incomplete match, but got none")
	}
	if err.Error() != "incomplete match found - all matches must have results" {
		t.Fatalf("Expected specific error message, got: %v", err)
	}
}

func TestUpdatePlayerStandingsDraws(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	tournament.Pair()
	// Set up a drawn match (equal game wins)
	err := tournament.AddResult(1, 1, 1, 1) // Alice and Bob each win 1 game, 1 draw
	if err != nil {
		t.Fatalf("Failed to add result: %v", err)
	}

	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	alice := tournament.players[1]
	bob := tournament.players[2]

	// Both should have 1 draw and 1 point
	if alice.draws != 1 || alice.points != 1 {
		t.Errorf("Alice should have 1 draw and 1 point, got draws=%d, points=%d", alice.draws, alice.points)
	}
	if bob.draws != 1 || bob.points != 1 {
		t.Errorf("Bob should have 1 draw and 1 point, got draws=%d, points=%d", bob.draws, bob.points)
	}
}

func TestFormatPlayersWithActualData(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	tournament.Pair()
	// Alice wins 2-0
	err := tournament.AddResult(1, 2, 0, 0)
	if err != nil {
		t.Fatalf("Failed to add result: %v", err)
	}

	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Test that FormatPlayers doesn't crash and shows non-zero values
	// We can't easily test the exact output without capturing it, but we can
	// verify the function runs without panic
	var buf bytes.Buffer
	tournament.FormatPlayers(&buf)
	output := buf.String()

	// Should contain actual numbers, not just zeros
	if !strings.Contains(output, "1") || !strings.Contains(output, "0") {
		t.Errorf("FormatPlayers output should contain actual statistics, got: %s", output)
	}

	t.Log("FormatPlayers successfully displays actual player statistics")
}

func TestUpdatePlayerStandingsAtomic(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")   // ID: 1
	tournament.AddPlayer("Bob")     // ID: 2
	tournament.AddPlayer("Charlie") // ID: 3
	tournament.AddPlayer("Diana")   // ID: 4

	// Initial stats should be 0
	for _, player := range tournament.players {
		if player.wins != 0 || player.losses != 0 || player.points != 0 {
			t.Fatalf("Player %s should start with 0 stats", player.name)
		}
	}

	tournament.Pair()
	pairings := tournament.GetRound()

	// Add results for only SOME pairings, leaving others incomplete
	completedMatches := 0
	for _, pairing := range pairings {
		if pairing.playerb != -1 && completedMatches == 0 {
			// Complete only the first non-bye match
			err := tournament.AddResult(pairing.playera, 2, 1, 0)
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
			completedMatches++
			break // Leave other matches incomplete
		}
	}

	// Try to update standings - should fail due to incomplete matches
	err := tournament.UpdatePlayerStandings()
	if err == nil {
		t.Fatal("Expected error due to incomplete matches, but UpdatePlayerStandings succeeded")
	}
	if err.Error() != "incomplete match found - all matches must have results" {
		t.Fatalf("Expected specific error message, got: %v", err)
	}

	// Verify NO player stats were updated (atomic behavior)
	for _, player := range tournament.players {
		if player.wins != 0 || player.losses != 0 || player.points != 0 {
			t.Errorf("Player %s stats were modified despite incomplete matches: wins=%d, losses=%d, points=%d",
				player.name, player.wins, player.losses, player.points)
		}
	}

	t.Log("Atomic behavior verified - no stats updated when matches are incomplete")
}

func TestCorrectPointsSystem(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")   // ID: 1
	tournament.AddPlayer("Bob")     // ID: 2
	tournament.AddPlayer("Charlie") // ID: 3

	// Test wins and draws with correct points
	tournament.Pair()
	pairings := tournament.GetRound()

	// Set up specific results to test points system
	for _, pairing := range pairings {
		if pairing.playerb != -1 {
			// Alice beats Bob (assuming Alice is playera)
			err := tournament.AddResult(pairing.playera, 2, 1, 0)
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
		}
		// Bye already has results
	}

	// Test another round with a draw - NextRound will update standings for round 1
	err := tournament.NextRound()
	if err != nil {
		t.Fatalf("Failed to advance to next round: %v", err)
	}
	tournament.Pair()

	// Add a draw result
	pairings = tournament.GetRound()
	for _, pairing := range pairings {
		if pairing.playerb != -1 {
			// Set up a draw (equal games)
			err := tournament.AddResult(pairing.playera, 1, 1, 0)
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
		}
	}

	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Verify points system: win = 3 points, draw = 1 point, bye = 3 points
	totalPoints := 0
	totalWins := 0
	totalDraws := 0

	for _, player := range tournament.players {
		totalPoints += player.points
		totalWins += player.wins
		totalDraws += player.draws
	}

	// With 3 players across 2 rounds, we should have specific point totals
	// Round 1: 1 match + 1 bye = 6 points total (3 for winner + 0 for loser + 3 for bye)
	// Round 2: 1 draw + 1 bye = 5 points total (1 + 1 for draw + 3 for bye)
	// Total: 11 points
	expectedTotalPoints := 11
	if totalPoints != expectedTotalPoints {
		t.Errorf("Expected %d total points with correct system, got %d", expectedTotalPoints, totalPoints)
	}

	t.Logf("Correct points system verified: %d total points across %d wins and %d draws",
		totalPoints, totalWins, totalDraws)
}
