## swisstools API Reference

### Types

- TournamentConfig
  - Fields: `PointsForWin int`, `PointsForDraw int`, `PointsForLoss int`, `ByeWins int`, `ByeLosses int`, `ByeDraws int`

- Tournament
  - Represents a tournament; manage players, pairing, results, standings, and persistence.

- Player
  - Fields: `Name string`, `Points int`, `Wins int`, `Losses int`, `Draws int`, `GameWins int`, `GameLosses int`, `GameDraws int`, `Notes []string`, `Removed bool`, `RemovedInRound int`, `ExternalID *int`, `Decklist *Decklist`

- Decklist
  - Fields: `Main map[string]int`, `Sideboard map[string]int`

- PlayerStanding
  - Fields: `Rank int`, `PlayerID int`, `Name string`, `Points int`, `Wins int`, `Losses int`, `Draws int`, `Tiebreakers TiebreakerData`

- TiebreakerData
  - Fields: `GameWinPercentage float64`, `OpponentMatchWinPct float64`, `OpponentGameWinPct float64`

- Pairing
  - Methods: `PlayerA() int`, `PlayerB() int`, `PlayerAWins() int`, `PlayerBWins() int`, `Draws() int`

- Playoff
  - Fields: `Seeds []int`, `Rounds []Round`, `CurrentRound int`, `Finished bool`

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
- `SetMaxRounds(n int)` → set maximum number of rounds (0 = no limit)
- `GetMaxRounds() int`

### Player management
- `AddPlayer(name string) error`
- `RemovePlayerById(id int) error`
- `RemovePlayerByName(name string) error`
- `GetPlayerID(name string) (int, bool)`
- `GetPlayerById(id int) (Player, bool)`
- `GetPlayerByName(name string) (Player, bool)`
- `GetPlayers() map[int]Player` → copy of all players keyed by ID (includes removed)

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
- `GetRoundByNumber(round int) ([]Pairing, error)` → pairings for any round (1-based)
- `AddResult(id int, wins int, losses int, draws int) error` → add result for the match containing `id`
- `UpdatePlayerStandings() error` → validates every active player is paired and every match has results, then updates cumulative stats
- `NextRound() error` → calls `UpdatePlayerStandings()` and advances the round index; auto-finishes if max rounds reached
- `FinishTournament() error` → explicitly finish the tournament (records current round standings first)

Result corrections (any round, including rounds already completed):
- `SetResult(round int, id int, wins int, losses int, draws int) error` → write the result for the match containing `id` in the given 1-based Swiss round; does not touch cumulative stats, so follow it with `RecomputeStandings()`
- `RecomputeStandings() error` → discard every player's cumulative `Points`/`Wins`/`Losses`/`Draws`/`GameWins`/`GameLosses`/`GameDraws` and rebuild them by replaying the stored results of every completed Swiss round (rounds `1..currentRound-1`, plus the current round once the tournament is finished). Player identity and metadata are preserved
- `SetPlayoffResult(round int, id int, wins int, losses int, draws int) error` → write the result for the match containing `id` in the given 0-based playoff round, including a round the bracket has already advanced past and a playoff that has already finished

Manual pairing modification (current round only; pairings with recorded results are protected):
- `SwapPlayers(idA, idB int) error` → exchange the positions of two currently paired players
- `SetPairing(idA, idB int) error` → create a new pairing between two unpaired players; pass `BYE_OPPONENT_ID` as `idB` to assign a bye
- `RemovePairing(id int) error` → delete the pairing containing `id`; both players become unpaired until a follow-up `SetPairing`

### Standings
- `GetStandings() []PlayerStanding` → standings sorted by points and tiebreakers

### Persistence
- `DumpTournament() ([]byte, error)` → JSON snapshot (versioned schema, includes playoff if present)
- `LoadTournament(data []byte) (Tournament, error)` → reconstruct from snapshot

### Playoff (single-elimination bracket)
- `StartPlayoff(topN int) error` → seed top N players from Swiss standings into a bracket; N must be a power of 2; Swiss must be finished
- `GetPlayoff() *Playoff` → returns the playoff bracket, or nil if not started
- `GetPlayoffStatus() string` → `none | in_progress | finished`
- `GetPlayoffRound() []Pairing` → current playoff round pairings
- `GetPlayoffRoundByNumber(round int) ([]Pairing, error)` → pairings for a given 0-based playoff round
- `AddPlayoffResult(id int, wins int, losses int, draws int) error` → record result for a playoff match; draws are not allowed (one player must advance)
- `NextPlayoffRound() error` → validate results, advance winners; finishes playoff when the final is decided

Manual playoff pairing modification — the bracket is rearranged, never resized (current playoff round only; pairings with recorded results are protected):
- `SwapPlayoffPlayers(idA, idB int) error` → exchange the positions of two currently paired players in the bracket

### Notes
- All `Player` fields are exported and readable on values returned by `GetPlayerById`, `GetPlayerByName`, and `GetPlayers`.
- Removed players remain in history but are excluded from future pairings.
- Byes are represented by `playerB == BYE_OPPONENT_ID` in pairings.
- Playoff seeding uses standard bracket order: seed 1 vs seed N, seed 2 vs seed N-1, etc.
- Playoff pairings are adjusted by swapping players, so a bracket round always holds every player who belongs to it. Repeated swaps reach any arrangement of the round's players, which is why the bracket has no remove/set pair.
- The Swiss round does have `RemovePairing` / `SetPairing`, because its roster is "every non-removed player" — a fact about the player list, so a missing pairing is always detectable. A bracket round's roster after the first round is a fact about history that `SetPlayoffResult` may contradict, so a pairing deleted from it could not be reconstructed.
- Every bracket operation is a single atomic change to persisted state, so a consumer may `DumpTournament` and `LoadTournament` between any two of them.
- A no-show or concession is a result, not a deletion: `AddPlayoffResult(survivor, 2, 0, 0)` advances the player who showed up and keeps both names in the bracket record. Byes are not supported in the playoff.
- Playoff matches cannot end in a draw — one player must have more game wins to advance.
- Corrections never change pairings or bracket structure. `SetResult` does not re-pair any later round — those matches were actually played between those people, so only points, records and tiebreakers move. `SetPlayoffResult` likewise never rebuilds or truncates the bracket, so a correction that flips a semifinal leaves the final recording the players who actually played it.
- `RecomputeStandings()` is the canonical view of the cumulative totals. Because it treats the in-progress round of an unfinished tournament as not yet applied, it discards a contribution made by calling the additive `UpdatePlayerStandings()` out of band without advancing.
- Tiebreakers and standings derive from the cumulative totals, so they follow `RecomputeStandings()` automatically.

