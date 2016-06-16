// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package client

import (
	"database/sql"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"time"
)

var dbMigrations = []struct {
	Version string
	Apply   func(db *sql.DB) error
}{
	{
		Version: "2016-06-15-22:58:03",
		Apply: func(db *sql.DB) error {
			_, err := db.Exec(`
			CREATE TABLE migrations (
			version TEXT
			)`)
			if err != nil {
				return err
			}
			_, err = db.Exec(insertIntoMigrations, "2016-06-15-22:58:03")
			return err
		},
	},
	{
		Version: "2016-06-15-23:22:24",
		Apply: func(db *sql.DB) error {
			_, err := db.Exec(`
			CREATE TABLE cachedPages (
			id INTEGER PRIMARY KEY ASC,
			cacheKey TEXT UNIQUE NOT NULL,
			lastFetched TIMESTAMP,
			body blob
			)`)
			return err
		},
	},
}

var createMigrationsTable = dbMigrations[0]

func (c *Client) setupDB() error {
	var firstRun bool
	rows, err := c.db.Query("select version from migrations")
	if sErr, ok := err.(sqlite3.Error); ok {
		if sErr.Error() == "no such table: migrations" {
			fmt.Println("[db] setting up database")
			err := createMigrationsTable.Apply(c.db)
			if err != nil {
				return errors.Wrap(err, "Creating migrations table")
			}
			firstRun = true
		} else {
			return errors.Wrap(err, "Checking migrations")
		}
	} else if err != nil {
		return errors.Wrap(err, "Checking migrations")
	}

	haveMigrations := make([]string, 0)
	for !firstRun && rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			return errors.Wrap(err, "Checking migrations")
		}
		haveMigrations = append(haveMigrations, version)
	}
	if firstRun {
		haveMigrations = []string{createMigrationsTable.Version}
	}

	haveIdx := 0
	stmt, err := c.db.Prepare(insertIntoMigrations)
	if err != nil {
		return errors.Wrap(err, "preparing insertIntoMigrations")
	}

	for _, m := range dbMigrations {
		if haveIdx >= len(haveMigrations) || haveMigrations[haveIdx] != m.Version {
			err := m.Apply(c.db)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Error performing migration %s", m.Version))
			}
			_, err = stmt.Exec(m.Version)
			if err != nil {
				// this is extremely bad
				// make a backup and restore it?
				return errors.Wrap(err, fmt.Sprintf("Recording migration %s", m.Version))
			}
			fmt.Println("[db] Applied migration", m.Version)
		} else {
			haveIdx++
		}
	}
	stmt.Close()

	if firstRun {
		fmt.Println("[db] database created")
	}

	stmtSelectExistsInCache, err = c.db.Prepare(selectExistsInCache)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	stmtSelectCacheData, err = c.db.Prepare(selectCacheData)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	stmtInsertCacheData, err = c.db.Prepare(insertCacheData)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	return nil
}

const insertIntoMigrations = `
INSERT INTO migrations (version) VALUES (?)`

const selectExistsInCache = `
SELECT id, lastFetched FROM cachedPages WHERE cacheKey = ?`

var stmtSelectExistsInCache *sql.Stmt

const selectCacheData = `
SELECT body FROM cachedPages WHERE id = ?`

var stmtSelectCacheData *sql.Stmt

const insertCacheData = `
INSERT INTO cachedPages
(cacheKey, lastFetched, body)
VALUES (?, ?, ?)`

var stmtInsertCacheData *sql.Stmt

const cacheStalePeriod = 196*time.Hour

// returns -1 if no match
func (c *Client) cacheCheck(u StoryURL) (int64, error) {
	row := stmtSelectExistsInCache.QueryRow(u.CacheKey())
	var id int64 = -1
	var lastUpdated time.Time
	err := row.Scan(&id, &lastUpdated)
	if err == sql.ErrNoRows {
		return -1, nil
	} else if err != nil {
		return -1, err
	}
	if time.Now().UTC().Add(-cacheStalePeriod).After(lastUpdated) {
		return -1, nil
	}
	return id, nil
}

func (c *Client) cacheGet(id int64) ([]byte, error) {
	row := stmtSelectCacheData.QueryRow(id)
	var b []byte
	err := row.Scan(&b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *Client) cachePut(u StoryURL, body []byte) error {
	_, err := stmtInsertCacheData.Exec(u.CacheKey(), time.Now().UTC(), body)
	return err
}
