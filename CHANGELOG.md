# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.0] - 2026-11-15

### Added

- **Result corrections** — an organizer can fix a mis-entered result in a round that has already been completed, later in the same event or after it is over, and have standings, tiebreakers and final ranks follow:
  - `SetResult(round, id, wins, losses, draws int) error` writes the game result for the match containing `id` in the given 1-based Swiss round, which may be a round already completed. It does not touch cumulative player stats.
  - `RecomputeStandings() error` discards every player's `Points`, `Wins`, `Losses`, `Draws`, `GameWins`, `GameLosses` and `GameDraws` and rebuilds them by replaying the stored results of every completed Swiss round (rounds `1..currentRound-1`, plus the current round once the tournament is finished). `Name`, `Notes`, `Removed`, `RemovedInRound`, `ExternalID` and `Decklist` are preserved. It errors without changing anything if a completed round holds an incomplete result, the same atomicity `UpdatePlayerStandings()` already guarantees.
  - `SetPlayoffResult(round, id, wins, losses, draws int) error` writes the game result for the match containing `id` in the given 0-based playoff round, which may be a round the bracket has already advanced past. Unlike `AddPlayoffResult()`, it also works after the playoff has finished, so a mis-entered final can still be fixed.
- Byes, dropped players and late entries all replay from stored history with no special case, because `rounds` records what actually happened.

### Notes

- **Corrections do not re-pair.** Correcting round 1 does not re-pair round 2: those matches were actually played between those people, so a correction fixes the record of a game, not the history of who sat down together. Only points, records and tiebreakers move.
- **Corrections do not rebuild or truncate the bracket.** Correcting a semifinal does not discard or re-pair the final, even when the correction flips who won that semifinal — the bracket then records that the player who lost still advanced, because those two people did play the final. `SetPlayoffResult()` never writes `Playoff.Rounds`, `Playoff.CurrentRound` or `Playoff.Finished`. Correcting the current, not-yet-advanced playoff round needs no special handling: `NextPlayoffRound()` reads it when it advances, so the corrected winner progresses.
- `UpdatePlayerStandings()` is additive and remains so. `RecomputeStandings()` is canonical, and by its rule the in-progress round of an unfinished tournament has not been applied — so it discards a contribution made by calling `UpdatePlayerStandings()` out of band without advancing the round.
- This release is purely additive: `AddResult()`, `AddPlayoffResult()`, `Pair()`, `NextRound()`, `NextPlayoffRound()` and `UpdatePlayerStandings()` are unchanged, and the export schema stays at version `1.0.0`.

### Fixed

- `TestMinimum33PercentRule` was flaky. It recorded results by pairing position, so a particular random first-round pairing made the player under test lose her matches and the assertions fail. It now records results by player identity, which also makes both of her opponents' raw percentages fall below the floor, so the test actually exercises the rule it is named for. Test-only change; no library behavior is affected.

## [0.5.0] - 2026-11-01

### Added

- **Manual pairing modification** — tournament organizers can now adjust pairings within a round without regenerating the whole round:
  - `SwapPlayers(idA, idB int) error` exchanges two currently paired players' positions.
  - `SetPairing(idA, idB int) error` pairs two currently unpaired players; passing `BYE_OPPONENT_ID` as `idB` assigns a bye.
  - `RemovePairing(id int) error` deletes a pairing from the current round; both players become unpaired.
  - Pairings with already-recorded results are protected from modification.
- **Manual playoff pairing modification** — the same three operations as `SwapPlayoffPlayers` / `SetPlayoffPairing` / `RemovePlayoffPairing` on the single-elimination bracket. Byes are not supported in the playoff variant; `SetPlayoffPairing` only accepts players that were previously removed from the current bracket round via `RemovePlayoffPairing`.

### Changed

- `UpdatePlayerStandings()` now also validates that every active (non-removed) player has a pairing in the current round, so that `RemovePairing` without a follow-up `SetPairing` is caught instead of silently zeroing a player's round. Flows through to `NextRound()`.
- `NextPlayoffRound()` now also validates that no players are currently unpaired in the bracket round (set by `RemovePlayoffPairing` and cleared by `SetPlayoffPairing`).

## [0.4.0] - 2026-09-17

### Added

- **Final standings including dropped players** — `GetFinalStandings()` returns standings with dropped players ranked in place by the same keys as `GetStandings()` (points, then OMW%, GW%, OGW%), rather than omitted. A dropped player falls naturally as the remaining rounds are played, but one who drops while ahead can still rank highly.
- `PlayerStanding` gained `Removed` and `RemovedInRound`, so callers can flag dropped players in a standings display.

### Changed

- `GetStandings()` is unchanged in behavior and still omits dropped players. It remains the view for deciding who is still competing — `StartPlayoff()` seeds the top cut from it, so a dropped player must never appear there. Use `GetFinalStandings()` for display and for recording results.

## [0.3.0] - 2026-04-15

### Added

- **Player listing** — `GetPlayers()` returns a copy of all players keyed by ID (includes removed players).
- **Round history** — `GetRoundByNumber(round int)` retrieves pairings for any past or current round by 1-based index.
- **Match results retrieval** — `PlayerAWins()`, `PlayerBWins()`, and `Draws()` accessor methods on `Pairing`.
- **Explicit tournament finish** — `FinishTournament()` to explicitly mark the tournament as finished.
- **Max rounds** — `SetMaxRounds(n int)` and `GetMaxRounds()` to cap Swiss rounds; `NextRound()` auto-finishes when the cap is reached.
- **Single-elimination playoff bracket** — `StartPlayoff(topN int)` seeds top N players (power of 2) from Swiss standings into a bracket with standard seeding (1 vs N, 2 vs N-1, etc.).
  - `GetPlayoff()`, `GetPlayoffStatus()`, `GetPlayoffRound()`, `GetPlayoffRoundByNumber()` for reading bracket state.
  - `AddPlayoffResult()` and `NextPlayoffRound()` for recording results and advancing the bracket.
- Playoff and max rounds state included in `DumpTournament()`/`LoadTournament()` serialization.

### Changed

- **Player type fields exported** — All `Player` struct fields are now exported (`Points`, `Wins`, `Losses`, `Draws`, `GameWins`, `GameLosses`, `GameDraws`, `Notes`, `Removed`, `RemovedInRound`, `ExternalID`, `Decklist`). Previously only `Name` was exported.

## [0.1.0] - 2025-10-30

### Added

- Swiss-system tournament pairing with byes and late entries.
- Match result recording at the game level with configurable points.
- Standings with tiebreakers (opponent match win %, game win %, opponent game win %).
- Player removal with history preservation.
- Optional player metadata (external ID, structured decklist).
- Versioned JSON dump/load for tournament persistence.
- Tournament configuration (custom points for win/draw/loss, bye scoring).

[0.6.0]: https://github.com/dstathis/swisstools/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/dstathis/swisstools/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/dstathis/swisstools/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/dstathis/swisstools/compare/v0.1.0...v0.3.0
[0.1.0]: https://github.com/dstathis/swisstools/releases/tag/v0.1.0
