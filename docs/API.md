## swisstools API Reference

### Types

- TournamentConfig
  - Fields: `PointsForWin int`, `PointsForDraw int`, `PointsForLoss int`, `ByeWins int`, `ByeLosses int`, `ByeDraws int`

- Tournament
  - Represents a tournament; manage players, pairing, results, standings, and persistence.

- Decklist
  - Fields: `Main map[string]int`, `Sideboard map[string]int`

- PlayerStanding
  - Fields: `Rank int`, `PlayerID int`, `Name string`, `Points int`, `Wins int`, `Losses int`, `Draws int`, `Tiebreakers TiebreakerData`

- TiebreakerData
  - Fields: `GameWinPercentage float64`, `OpponentMatchWinPct float64`, `OpponentGameWinPct float64`

- Pairing
  - Methods: `PlayerA() int`, `PlayerB() int`

### Constants
- `BYE_OPPONENT_ID = -1`
- `UNINITIALIZED_RESULT = -1`

### Constructors
- `DefaultConfig() TournamentConfig`
- `NewTournament() Tournament`
- `NewTournamentWithConfig(config TournamentConfig) Tournament`

### Tournament state
- `GetStatus() string` → `setup | in_progress | finished`
- `GetCurrentRound() int`
- `GetPlayerCount() int`

### Player management
- `AddPlayer(name string) error`
- `RemovePlayerById(id int) error`
- `RemovePlayerByName(name string) error`
- `GetPlayerID(name string) (int, bool)`
- `GetPlayerById(id int) (Player, bool)`
- `GetPlayerByName(name string) (Player, bool)`

Optional metadata:
- `SetPlayerExternalID(id int, externalID int) error`
- `ClearPlayerExternalID(id int) error`
- `GetPlayerExternalID(id int) (*int, bool)`
- `SetPlayerDecklist(id int, deck Decklist) error`
- `ClearPlayerDecklist(id int) error`
- `GetPlayerDecklist(id int) (*Decklist, bool)`

### Pairing and results
- `StartTournament() error` → pairs first round
- `Pair(allowRepair bool) error` → pair current round; if `allowRepair` is true, clears existing pairings for that round
- `GetRound() []Pairing` → current round pairings
- `AddResult(id int, wins int, losses int, draws int) error` → add result for the match containing `id`
- `UpdatePlayerStandings() error` → validates all matches are complete, then updates cumulative stats
- `NextRound() error` → calls `UpdatePlayerStandings()` and advances the round index

### Standings
- `GetStandings() []PlayerStanding` → standings sorted by points and tiebreakers

### Persistence
- `DumpTournament() ([]byte, error)` → JSON snapshot (versioned schema)
- `LoadTournament(data []byte) (Tournament, error)` → reconstruct from snapshot

### Notes
- Removed players remain in history but are excluded from future pairings.
- Byes are represented by `playerB == BYE_OPPONENT_ID` in pairings.

