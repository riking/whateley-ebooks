// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package client

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func assetCacheKey(u *url.URL) string {
	if u.Host != "whateleyacademy.net" {
		panic(errors.Errorf("Bad host for asset cache: %s", u.String()))
	}
	return u.EscapedPath()
}

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
			body BLOB
			)`)
			return err
		},
	},
	{
		Version: "2016-07-02-07:02:32",
		Apply: func(db *sql.DB) error {
			_, err := db.Exec(`
			CREATE TABLE cachedAssets (
			id INTEGER PRIMARY KEY ASC,
			cacheKey TEXT UNIQUE NOT NULL,
			lastFetched TIMESTAMP,
			contentType TEXT,
			body BLOB
			)`)
			return err
		},
	},
}

var createMigrationsTable = dbMigrations[0]

func (c *WANetwork) setupDB() error {
	var firstRun bool
	rows, err := c.db.Query("select version from migrations")
	if sErr, ok := err.(sqlite3.Error); ok {
		if sErr.Error() == "no such table: migrations" {
			fmt.Fprintln(os.Stderr, "[db] setting up database")
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
			fmt.Fprintln(os.Stderr, "[db] Applied migration", m.Version)
		} else {
			haveIdx++
		}
	}
	stmt.Close()

	if firstRun {
		fmt.Fprintln(os.Stderr, "[db] database created")
	}

	stmtSelectStoryExistsInCache, err = c.db.Prepare(selectStoryExistsInCache)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	stmtSelectStoryCacheData, err = c.db.Prepare(selectStoryCacheData)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	stmtInsertStoryCacheData, err = c.db.Prepare(insertStoryCacheData)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	stmtUpdateStoryCacheData, err = c.db.Prepare(updateStoryCacheData)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	stmtSelectAssetExistsInCache, err = c.db.Prepare(selectAssetExistsInCache)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	stmtSelectAssetCacheData, err = c.db.Prepare(selectAssetCacheData)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	stmtInsertAssetCacheData, err = c.db.Prepare(insertAssetCacheData)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	stmtUpdateAssetCacheData, err = c.db.Prepare(updateAssetCacheData)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	stmtSearchStoryFulltext, err = c.db.Prepare(searchStoryFulltext)
	if err != nil {
		return errors.Wrap(err, "preparing statements")
	}

	return nil
}

const (
	insertIntoMigrations = `
INSERT INTO migrations (version) VALUES (?)`
	selectStoryExistsInCache = `
SELECT id, lastFetched FROM cachedPages WHERE cacheKey = ?`
	selectStoryCacheData = `
SELECT body FROM cachedPages WHERE id = ?`
	insertStoryCacheData = `
INSERT INTO cachedPages
(cacheKey, lastFetched, body)
VALUES (?, ?, ?)`
	updateStoryCacheData = `
UPDATE cachedPages
SET lastFetched=?, body=?
WHERE id = ?`
	selectAssetExistsInCache = `
SELECT id, lastFetched FROM cachedAssets WHERE cacheKey = ?`
	selectAssetCacheData = `
SELECT body, contentType FROM cachedAssets WHERE id = ?`
	insertAssetCacheData = `
INSERT INTO cachedAssets
(cacheKey, lastFetched, body, contentType)
VALUES (?, ?, ?, ?)`
	updateAssetCacheData = `
UPDATE cachedAssets
SET lastFetched=?, body=?, contentType=?
WHERE id = ?`
	searchStoryFulltext = `
SELECT cacheKey
FROM cachedPages
WHERE body LIKE '%' || ? || '%'`
)

var (
	stmtSelectStoryExistsInCache *sql.Stmt
	stmtSelectStoryCacheData     *sql.Stmt
	stmtInsertStoryCacheData     *sql.Stmt
	stmtUpdateStoryCacheData     *sql.Stmt
	stmtSelectAssetExistsInCache *sql.Stmt
	stmtSelectAssetCacheData     *sql.Stmt
	stmtInsertAssetCacheData     *sql.Stmt
	stmtUpdateAssetCacheData     *sql.Stmt

	stmtSearchStoryFulltext *sql.Stmt
)

const cacheStalePeriod = 1960 * time.Hour

var errExpired = errors.Errorf("cache entry expired")

// returns -1 if no match
func (c *WANetwork) cacheCheckStory(u StoryURL) (int64, error) {
	row := stmtSelectStoryExistsInCache.QueryRow(u.CacheKey())
	var id int64 = -1
	var lastUpdated time.Time
	err := row.Scan(&id, &lastUpdated)
	if err == sql.ErrNoRows {
		return -1, nil
	} else if err != nil {
		return -1, err
	}
	if !c.options.Offline && time.Now().UTC().Add(-cacheStalePeriod).After(lastUpdated) {
		return id, errExpired
	}
	return id, nil
}

func (c *WANetwork) cacheGetStory(id int64) ([]byte, error) {
	row := stmtSelectStoryCacheData.QueryRow(id)
	var b []byte
	err := row.Scan(&b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *WANetwork) cachePutStory(id int64, u StoryURL, body string) error {
	var err error
	if id == -1 {
		_, err = stmtInsertStoryCacheData.Exec(u.CacheKey(), time.Now().UTC(), body)
	} else {
		_, err = stmtUpdateStoryCacheData.Exec(time.Now().UTC(), body, id)
	}
	return err
}

func (c *WANetwork) cacheCheckAsset(u *url.URL) (int64, error) {
	row := stmtSelectAssetExistsInCache.QueryRow(assetCacheKey(u))
	var id int64 = -1
	var lastUpdated time.Time
	err := row.Scan(&id, &lastUpdated)
	if err == sql.ErrNoRows {
		return -1, nil
	} else if err != nil {
		return -1, err
	}
	if !c.options.Offline && time.Now().UTC().Add(-cacheStalePeriod).After(lastUpdated) {
		return id, errExpired
	}
	return id, nil
}

func (c *WANetwork) cacheGetAsset(id int64) ([]byte, string, error) {
	row := stmtSelectAssetCacheData.QueryRow(id)
	var b []byte
	var ct string
	err := row.Scan(&b, &ct)
	if err != nil {
		return nil, "", err
	}
	return b, ct, nil
}

func (c *WANetwork) cachePutAsset(id int64, u *url.URL, body []byte, contentType string) error {
	var err error
	if id == -1 {
		_, err = stmtInsertAssetCacheData.Exec(assetCacheKey(u), time.Now().UTC(), body, contentType)
	} else {
		_, err = stmtUpdateAssetCacheData.Exec(time.Now().UTC(), body, contentType, id)
	}
	return err
}

func (c *WANetwork) SearchFulltext(search string) ([]string, error) {
	rows, err := stmtSearchStoryFulltext.Query(search)
	if err != nil {
		return nil, err
	}
	var storyIDs []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		storyIDs = append(storyIDs, id)
	}
	return storyIDs, rows.Err()
}

func (c *WANetwork) DBTest() {
	rows, err := c.db.Query(
		"SELECT cacheKey " +
			"FROM cachedPages " +
			"WHERE body LIKE '%\u001c%' ")
	if err != nil {
		fmt.Println("err", err)
		return
	}
	for rows.Next() {
		var cacheKey string
		rows.Scan(&cacheKey)
		fmt.Println(cacheKey)
	}
	if rows.Err() != nil {
		fmt.Println("err", rows.Err())
	}
}
