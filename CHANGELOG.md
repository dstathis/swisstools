# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-04-14

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

[0.2.0]: https://github.com/dstathis/swisstools/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/dstathis/swisstools/releases/tag/v0.1.0
