package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"unicode"

	"github.com/spf13/cobra"
)

const (
	batchSize  = 50_000
	numWorkers = 4 // Count be numCPU or numCPU * 2
)

// countnameCmd represents the countname command
var countnameCmd = &cobra.Command{
	Use:   "countname",
	Short: "Find the most common first-name starting letter in the database",
	Long: `countname queries the people database and determines which letter of the
alphabet is the most common starting letter among all first names.

It uses a concurrent worker pool to process rows in batches of 50,000,
distributing work across multiple goroutines for efficiency. Results from each
batch are aggregated into a final tally, and the winning letter along with its
count is logged on completion.

The people.db SQLite database must exist before running this command.
Run the seed command first if it does not.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("Processing names...")
		return countConcurrent()
	},
}

func init() {
	rootCmd.AddCommand(countnameCmd)
}

func countConcurrent() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db, err := sql.Open("sqlite", "people.db")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	var total int
	if err := db.QueryRow("SELECT COUNT(*) FROM people").Scan(&total); err != nil {
		return fmt.Errorf("count rows: %w", err)
	}
	slog.Info("starting", "total_rows", total, "batch_size", batchSize, "workers", numWorkers)

	batchCh := make(chan []string, numWorkers)
	resultCh := make(chan [26]int, numWorkers)
	errCh := make(chan error, 1)

	// producer: fetch batches sequentially and hand off to the worker pool
	go func() {
		defer close(batchCh)
		for batchNum, offset := 0, 0; offset < total; batchNum, offset = batchNum+1, offset+batchSize {
			if ctx.Err() != nil {
				return
			}
			rows, err := db.QueryContext(ctx,
				"SELECT first_name FROM people LIMIT ? OFFSET ?",
				batchSize, offset,
			)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("query batch %d: %w", batchNum, err):
				default:
				}
				return
			}
			names := make([]string, 0, batchSize)
			for rows.Next() {
				var name string
				if err := rows.Scan(&name); err != nil {
					rows.Close()
					select {
					case errCh <- fmt.Errorf("scan batch %d: %w", batchNum, err):
					default:
					}
					return
				}
				names = append(names, name)
			}
			rows.Close()
			if err := rows.Err(); err != nil {
				select {
				case errCh <- fmt.Errorf("rows batch %d: %w", batchNum, err):
				default:
				}
				return
			}
			select {
			case batchCh <- names:
			case <-ctx.Done():
				return
			}
		}
	}()

	// worker pool: count all letters in each batch and send the full tally
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for names := range batchCh {
				var counts [26]int
				for _, name := range names {
					if len(name) == 0 {
						continue
					}
					r := unicode.ToUpper(rune(name[0]))
					if r >= 'A' && r <= 'Z' {
						counts[r-'A']++
					}
				}
				resultCh <- counts
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// aggregate every letter from every batch
	var letterCounts [26]int
	for counts := range resultCh {
		for i, v := range counts {
			letterCounts[i] += v
		}
	}

	select {
	case err := <-errCh:
		return err
	default:
	}
	if ctx.Err() != nil {
		return fmt.Errorf("cancelled: %w", ctx.Err())
	}

	winIdx, winVal := 0, 0
	for i, v := range letterCounts {
		if v > winVal {
			winVal = v
			winIdx = i
		}
	}
	slog.Info("done", "letter", string(rune('A'+winIdx)), "count", winVal)
	return nil
}
