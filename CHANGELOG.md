# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.0] - 2026-11-15

### Removed

- **`SetPlayoffPairing()` and `RemovePlayoffPairing()`** — the two communicated through an in-memory set of unpaired players that `DumpTournament()` never wrote, so the documented two-step flow could not be used by a consumer that persists state between operations (issue #4). Removing a pairing and reloading left those players stranded: `SetPlayoffPairing()` rejected the very players it had just freed, and there was no way to put them back.
  - **Migration:** use `SwapPlayoffPlayers(idA, idB)`. Repeated swaps reach any arrangement of a bracket round's players, so nothing the removed pair could express is lost — a `RemovePlayoffPairing` plus two `SetPlayoffPairing` calls to turn `{a0,b0},{a1,b1}` into `{a0,a1},{b0,b1}` is the single call `SwapPlayoffPlayers(b0, a1)`. Unlike the removed pair, a swap is one atomic change to persisted state, so it is safe to dump and reload between any two bracket operations.
  - A no-show or concession is a result rather than a deletion: `AddPlayoffResult(survivor, 2, 0, 0)` advances the player who showed up and keeps both names in the bracket record.
  - The Swiss `SetPairing()` / `RemovePairing()` are unaffected. Their roster is every non-removed player, so a missing pairing is always detectable; a bracket round's roster is not re-derivable, which is the asymmetry behind this change.

### Fixed

- **The final round now counts towards tiebreakers once a tournament is finished** (issue #3). `calculateTiebreakers` walked a player's opponents over rounds `1..currentRound-1`, which is right while a tournament is running — the in-progress round's results are not final — but a finished tournament never advances past its last round: `NextRound()` returns at the `maxRounds` cap before incrementing, and `FinishTournament()` never increments. So on both paths the last round's opponents never entered `OpponentMatchWinPct` or `OpponentGameWinPct`. The last round scored points but counted for nothing else, which could mis-order players who finish level on points — including at the top of the final standings, and therefore the top cut `StartPlayoff()` seeds from `GetStandings()`. Tiebreakers now cover exactly the rounds that have been folded into the player totals, using the same rule `RecomputeStandings()` already applied, now shared between the two so they cannot drift apart again.
  - **This changes values for existing data.** A finished tournament's `OpponentMatchWinPct` and `OpponentGameWinPct` differ from prior versions for the same stored results, and players level on points may rank differently — that is the fix. Reloading an old dump recalculates them, because tiebreakers are derived, not stored; the export format is unchanged.
  - It also matters for the 0.6.0 result corrections: correcting a final-round result previously moved points without moving tiebreakers, which looked like a bug in the correction feature.
  - Swiss pairing is unaffected. `havePlayedBefore()` keeps its exclusive bound, which means "every round before the one being paired" and is correct as it stands — pairing never runs on a finished tournament.

- **`NextPlayoffRound()` no longer panics on a bracket round with too few pairings.** It paired winners two at a time through an unguarded index, so an odd number of winners read past the end of the slice — in a top-8 bracket, removing one quarterfinal and reloading produced an `index out of range` panic in the caller's process rather than an error. A round that cannot advance is now rejected with `"playoff round is incomplete - bracket cannot advance"`. A round short by one pairing in a top-4 bracket is rejected too: it leaves a single winner, which previously looked exactly like a champion and finished the playoff from a semifinal. So is an empty round, which previously advanced silently, leaving the playoff `in_progress` on a round with no players in it. No library operation can produce either state now, but a hand-edited dump or direct mutation of the exported `Playoff.Rounds` still can.

### Notes

- The export format is unchanged and this release reads dumps from earlier versions as-is. The fix removes transient state rather than persisting it, so there is nothing new to serialize.

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
