package sqlite

import (
	"database/sql"
	"log"

	// sqlite3 driver.
	_ "github.com/mattn/go-sqlite3"
	"github.com/rschio/repoTagger/repo"
	"github.com/rschio/repoTagger/storage"
)

type service struct {
	DB *sql.DB
}

func createDB(database *sql.DB) error {
	stmt := `
		CREATE TABLE IF NOT EXISTS repo (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			desc TEXT,
			url_http TEXT NOT NULL,
			lang TEXT
		);
		CREATE TABLE IF NOT EXISTS tag (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			repo_id INTEGER NOT NULL
		);
	`
	_, err := database.Exec(stmt)
	return err
}

// New returns a new storage with a sqlite database
// of path db. If db does not exists New create the
// file and tables repo and tag.
func New(db string) (storage.Storage, error) {
	database, err := sql.Open("sqlite3", db)
	if err != nil {
		log.Fatalf("failed to connect to db %s: %v", db, err)
		return nil, err
	}
	err = createDB(database)
	if err != nil {
		return nil, err
	}

	return &service{DB: database}, nil
}

func (s *service) Close() error { return s.DB.Close() }

func (s *service) deleteTagsOfRepo(id int) error {
	stmt := "DELETE FROM tag WHERE repo_id = ?;"
	_, err := s.DB.Exec(stmt, id)
	return err
}

func (s *service) insertTags(r *repo.Repo) error {
	stmt := "INSERT INTO tag (name, repo_id) VALUES (?, ?);"
	for _, tag := range r.Tags {
		_, err := s.DB.Exec(stmt, tag, r.ID)
		if err != nil {
			log.Printf("failed to insert tag %s: %v", tag, err)
			return err
		}
	}
	return nil
}

func (s *service) UpdateTags(r *repo.Repo) error {
	err := s.deleteTagsOfRepo(r.ID)
	if err != nil {
		return err
	}
	return s.insertTags(r)
}

func (s *service) InsertRepo(r *repo.Repo) error {
	stmt := "INSERT INTO repo (id, name, desc, url_http, lang) VALUES (?, ?, ?, ?, ?);"
	_, err := s.DB.Exec(stmt, r.ID, r.Name, r.Desc, r.URLHTTP, r.Lang)
	if err != nil {
		log.Printf("failed to insert repo %s: %v", r.Name, err)
		return err
	}
	return s.insertTags(r)
}

func (s *service) GetRepo(id int) (*repo.Repo, error) {
	stmt := "SELECT id, name, desc, url_http, lang FROM repo WHERE id = ?;"
	row := s.DB.QueryRow(stmt, id)
	r := &repo.Repo{}

	err := row.Scan(&r.ID, &r.Name, &r.Desc, &r.URLHTTP, &r.Lang)
	if err != nil {
		return nil, err
	}

	r.Tags, err = s.getTags(id)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *service) GetReposByTag(tag string) ([]*repo.Repo, error) {
	stmt := `SELECT r.id, r.name, r.desc, r.url_http, r.lang FROM 
		repo AS r JOIN tag WHERE tag.repo_id = r.id AND tag.name LIKE ? || '%';`

	// get all repos.
	if tag == "" {
		stmt = "SELECT id, name, desc, url_http, lang FROM repo;"
	}

	rows, err := s.DB.Query(stmt, tag)
	if err != nil {
		log.Printf("failed to get repos: %v", err)
		return nil, err
	}
	defer rows.Close()

	// tags can be incomplete word, so one tag
	// can generate more than one search and
	// one repo can have more than one of match
	// tag. Avoid duplicate repos with duplicated.
	duplicated := make(map[int]struct{})
	repos := make([]*repo.Repo, 0)

	for rows.Next() {
		r := &repo.Repo{}

		err = rows.Scan(&r.ID, &r.Name, &r.Desc, &r.URLHTTP, &r.Lang)
		if err != nil {
			log.Printf("failed to get repo IDs: %v", err)
			return nil, err
		}

		if _, ok := duplicated[r.ID]; ok {
			continue
		}

		r.Tags, err = s.getTags(r.ID)
		if err != nil {
			return nil, err
		}

		repos = append(repos, r)
		duplicated[r.ID] = struct{}{}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return repos, nil
}

func (s *service) getTags(repoID int) ([]string, error) {
	stmt := "SELECT name FROM tag WHERE repo_id = ?;"
	tags := make([]string, 0)

	rows, err := s.DB.Query(stmt, repoID)
	if err != nil {
		log.Printf("failed to get tags from repo %d: %v", repoID, err)
		return nil, err
	}

	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			log.Printf("failed to scan row: %v", err)
			return nil, err
		}
		tags = append(tags, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}
