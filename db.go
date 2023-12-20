package main

import (
	"database/sql"
	"github.com/gofiber/storage/sqlite3"
	"time"
)

func initDB(location string) *sqlite3.Storage {
	storage := sqlite3.New(sqlite3.Config{
		Database:        location,
		Table:           "sessions",
		Reset:           false,
		GCInterval:      10 * time.Second,
		MaxOpenConns:    100,
		MaxIdleConns:    100,
		ConnMaxLifetime: 1 * time.Second,
	})

	return storage
}

func dbInitStorage(db *sql.DB) error {
	const initDB string = `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER NOT NULL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			password BLOB NOT NULL,
			salt TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS score (
			id INTEGER NOT NULL PRIMARY KEY,
			user INTEGER NOT NULL,
			Challenge TEXT NOT NULL,
			points INTEGER NOT NULL,
			time DATETIME NOT NULL,
			unique (user, Challenge)
		);
		CREATE TABLE IF NOT EXISTS hints (
			id INTEGER NOT NULL PRIMARY KEY,
			user INTEGER NOT NULL,
			Challenge TEXT NOT NULL,
			hintid string NOT NULL,
			points INTEGER NOT NULL,
			unique (user, Challenge, hintid)
		);`
	if _, err := db.Exec(initDB); err != nil {
		return err
	}
	return nil
}

// Users

func dbUserGetName(db *sql.DB, id int) (string, error) {
	var name string
	row := db.QueryRow(`SELECT name FROM users WHERE id=$1;`, id)
	err := row.Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}

func dbUserGetScore(db *sql.DB, id int) (int, error) {
	var score int
	row := db.QueryRow(`SELECT COALESCE(SUM(score.points), 0) FROM score WHERE user=$1;`, id)
	err := row.Scan(&score)
	if err != nil {
		return 0, err
	}
	return score, nil
}

func dbUserRegister(db *sql.DB, username, hash, salt string) (int, error) {
	res, err := db.Exec("INSERT INTO users VALUES(NULL,?,?,?);", username, hash[:], salt)
	if err != nil {
		return 0, err
	}
	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return 0, err
	}
	return int(id), err
}

// Challenges

func dbChallengeAddSolve(db *sql.DB, userID int, challengeID string, challengePoints int) error {
	_, err := db.Exec("INSERT INTO score VALUES(NULL,?,?,?, ?);", userID, challengeID, challengePoints, time.Now().UTC())
	return err
}

func dbChallengeGetSolved(db *sql.DB, userID int) ([]string, error) {
	var Challenges []string

	rows, err := db.Query(`SELECT Challenge FROM score WHERE user=$1;`, userID)
	if err != nil {
		return Challenges, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var challenge string
		if err := rows.Scan(&challenge); err != nil {
			return Challenges, err
		}
		Challenges = append(Challenges, challenge)
	}
	if err = rows.Err(); err != nil {
		return Challenges, err
	}
	return Challenges, nil
}

// Score

func dbGetScoreboard(db *sql.DB) ([][]interface{}, error) {
	var scores [][]interface{}

	rows, err := db.Query(`SELECT users.id, users.name, COALESCE(SUM(score.points), 0) AS total_points
										FROM users
										LEFT JOIN score ON users.id = score.user
										GROUP BY users.id, users.name
										ORDER BY total_points DESC;`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var id int
		var name string
		var score int
		if err := rows.Scan(&id, &name, &score); err != nil {
			return scores, err
		}
		scores = append(scores, []interface{}{id, name, score})
	}
	if err = rows.Err(); err != nil {
		return scores, err
	}
	return scores, nil
}

// Hints

func dbInsertHint(db *sql.DB, userID int, challengeID, hintID string, cost int) error {
	_, err := db.Exec("INSERT INTO hints VALUES(NULL,?,?,?,?);", userID, challengeID, hintID, cost)
	if err != nil {
		return err
	}
	return nil
}

func dbHintGetBoughtIDs(db *sql.DB, userID int, challengeID string) ([]string, error) {
	var hintIDs []string

	rows, err := db.Query(`SELECT hintid FROM hints WHERE user=$1 AND challenge=$2;`, userID, challengeID)
	if err != nil {
		return hintIDs, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var hintID string
		if err := rows.Scan(&hintID); err != nil {
			return hintIDs, err
		}
		hintIDs = append(hintIDs, hintID)
	}
	if err = rows.Err(); err != nil {
		return hintIDs, err
	}
	return hintIDs, nil
}

func dbHintGetCost(db *sql.DB, userID int, challengeID string) (int, error) {
	var costs int
	row := db.QueryRow(
		`SELECT COALESCE(SUM(hints.points), 0) FROM hints  WHERE user=$1 AND challenge=$2;`, userID, challengeID)
	//goland:noinspection GoDirectComparisonOfErrors
	switch err := row.Scan(&costs); err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return costs, nil
	default:
		return 0, err
	}
}
