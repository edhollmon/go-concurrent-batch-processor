# go-concurrent-batch-processor

A CLI tool for seeding and batch-processing large CSV datasets into SQLite.

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
