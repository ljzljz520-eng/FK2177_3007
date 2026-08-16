# Club Activity Scoring List

`clubscore` is a pure Go service library and command-line entry point for club activity records. An activity contains an activity number, name, responsible person, score, and participant count.

## Run

Use the CLI from the module root:

```bash
GOTOOLCHAIN=local go run ./cmd/clubscore help
GOTOOLCHAIN=local go run ./cmd/clubscore add --file activities.yaml --id A-001 --name "Campus Run" --leader Lin --score 92 --participants 80
GOTOOLCHAIN=local go run ./cmd/clubscore list --file activities.yaml
```

The archive is YAML and is written by `SaveFile`. It has no database or service dependency.

## API operations

`NewList` creates an empty doubly linked list. `Append` inserts at the tail, `Delete` removes by activity number, `Modify` changes a record, `Query` looks up a record by number, and `Save`/`Load` serialize the list with `gopkg.in/yaml.v3`. `SortByID` orders ascending by number. `SortByScore` orders descending by score and uses the number as a deterministic tie breaker.

The list owns its node links and protects every operation with a read/write mutex. `ToSlice` returns a copy of the activity values, so callers cannot mutate list state without an operation.

## Complexity

| Operation | Complexity |
| --- | --- |
| Tail append | O(1) expected, including the index update |
| Query by number | O(1) expected |
| Delete by number | O(1) expected |
| Modify by number | O(1) expected |
| Save or snapshot | O(n) time and O(n) output space |
| Load | O(n) time and O(n) space |
| Sort by number or score | O(n log n) time and O(n) temporary node references |

The expected constant-time operations rely on the Go map index; the linked links make tail insertion and removal independent of list length.
