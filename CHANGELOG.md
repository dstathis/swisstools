# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.4.0]: https://github.com/dstathis/swisstools/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/dstathis/swisstools/compare/v0.1.0...v0.3.0
[0.1.0]: https://github.com/dstathis/swisstools/releases/tag/v0.1.0
