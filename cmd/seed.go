package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

type person struct {
	Index     string `csv:"Index"`
	UserID    string `csv:"User Id"`
	FirstName string `csv:"First Name"`
	LastName  string `csv:"Last Name"`
	Sex       string `csv:"Sex"`
	Email     string `csv:"Email"`
	Phone     string `csv:"Phone"`
	Dob       string `csv:"Date of birth"`
	JobTitle  string `csv:"Job Title"`
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed SQL DB with sample cvs data",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Seeding Database...")
		seed()
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

func seed() {
	start := time.Now()

	peopleFile, err := os.OpenFile("people-896000.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := peopleFile.Close(); err != nil {
			panic(err)
		}
	}()

	people := []*person{}
	if err := gocsv.UnmarshalFile(peopleFile, &people); err != nil {
		panic(err)
	}

	os.Remove("people.db")

	db, err := sql.Open("sqlite", "people.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS people (
		index_col TEXT,
		user_id TEXT,
		first_name TEXT,
		last_name TEXT,
		sex TEXT,
		email TEXT,
		phone TEXT,
		dob TEXT,
		job_title TEXT
	)`)
	if err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	stmt, err := tx.Prepare(`INSERT INTO people (index_col, user_id, first_name, last_name, sex, email, phone, dob, job_title)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, p := range people {
		if _, err := stmt.Exec(p.Index, p.UserID, p.FirstName, p.LastName, p.Sex, p.Email, p.Phone, p.Dob, p.JobTitle); err != nil {
			panic(err)
		}
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}

	fmt.Println("Total:", len(people))
	fmt.Println("Time:", time.Since(start))
}
