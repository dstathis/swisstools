## swisstools

Swiss tournament pairing and standings library for Go.

### Features
- Proper Swiss pairing with byes and late entries
- Match and game tracking with configurable points
- Tiebreakers: opponent match win %, game win %, opponent game win %
- Drop/remove players while preserving history
- Optional player metadata: external ID and structured decklist
- Versioned dump/load to JSON for persistence and resume

## Installation
```bash
go get github.com/dstathis/swisstools
```

## Quick start
```go
package main

import (
    "fmt"
    st "github.com/dstathis/swisstools"
)

func main() {
    t := st.NewTournament()
    t.AddPlayer("Alice")
    t.AddPlayer("Bob")
    t.AddPlayer("Charlie")

    // Optional metadata
    t.SetPlayerExternalID(1, 10101)
    t.SetPlayerDecklist(1, st.Decklist{Main: map[string]int{"Island": 20}})

    // Start and play round 1
    _ = t.StartTournament()
    for _, p := range t.GetRound() {
        if p.PlayerB() == st.BYE_OPPONENT_ID { continue }
        _ = t.AddResult(p.PlayerA(), 2, 1, 0)
    }
    _ = t.UpdatePlayerStandings()

    for _, s := range t.GetStandings() {
        fmt.Printf("%d. %s (%d points)\n", s.Rank, s.Name, s.Points)
    }

    // Persist and restore later
    data, _ := t.DumpTournament()
    restored, _ := st.LoadTournament(data)
    _ = restored.NextRound()
    _ = restored.Pair(false)
}
```

## Persistence (dump/load)
- Dump: `data, err := t.DumpTournament()`
- Load: `t2, err := swisstools.LoadTournament(data)`

The JSON schema is versioned independently of the Go module. See `exportVersion` in `export.go`.

## Versioning
- Go module versions are controlled via Git tags (e.g., `v1.2.3`).
- For major version 2+, the module path must include `/v2`.

## API
See [`docs/API.md`](docs/API.md) for the full API reference.

## License
MIT. See `LICENSE`.

Swisstools is a library for running swiss style tournaments.
