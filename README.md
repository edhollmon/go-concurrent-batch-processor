# go-concurrent-batch-processor

A CLI tool for seeding large CSV datasets into SQLite and running concurrent batch analysis over them.

## Requirements

- Go 1.25+

## Installation

```bash
git clone https://github.com/edhollmon/go-concurrent-batch-processor
cd go-concurrent-batch-processor
go build -o batch-processor .
```

## Usage

```bash
./batch-processor [command]
```

### Commands

#### `seed`

Reads `people-896000.csv` from the current directory, drops any existing `people.db`, and bulk-inserts all rows into a new SQLite database using a single transaction.

```bash
./batch-processor seed
```

Output:
```
INFO Seeding Database...
INFO Seeding complete total=896000 duration=4.2s
```

#### `countname`

Queries `people.db` and finds the letter of the alphabet that the most first names start with. Work is distributed across a pool of 4 concurrent goroutines, each processing batches of 50,000 rows, with results aggregated into a final tally.

Requires `people.db` to exist — run `seed` first.

```bash
./batch-processor countname
```

Output:
```
INFO Processing names...
INFO starting total_rows=896000 batch_size=50000 workers=4
INFO done letter=J count=72381
```

## Data Model

The `people` table schema:

| Column       | Type |
|--------------|------|
| `index_col`  | TEXT |
| `user_id`    | TEXT |
| `first_name` | TEXT |
| `last_name`  | TEXT |
| `sex`        | TEXT |
| `email`      | TEXT |
| `phone`      | TEXT |
| `dob`        | TEXT |
| `job_title`  | TEXT |

## Dependencies

| Package | Purpose |
|---------|---------|
| [`cobra`](https://github.com/spf13/cobra) | CLI framework |
| [`gocsv`](https://github.com/gocarina/gocsv) | CSV unmarshaling |
| [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) | Pure-Go SQLite driver (no CGO required) |
