package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/rschio/repoTagger/repo"
	"github.com/rschio/repoTagger/storage"
	"github.com/rschio/repoTagger/storage/sqlite"
)

type server struct {
	store storage.Storage
}

func (s *server) getRepos(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	user := r.URL.Path[len("/repos/"):]
	repos, err := repo.GetGithubRepos(user)
	if err != nil {
		if _, ok := err.(repo.NotFoundErr); ok {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	for _, repository := range repos {
		s.store.InsertRepo(repository)
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *server) search(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	tag := r.URL.Path[len("/search/"):]
	repos, err := s.store.GetReposByTag(tag)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	if len(repos) == 0 {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(repos)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
	}
}

func (s *server) suggest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.URL.Path[len("/suggest/"):])
	if err != nil {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	repository, err := s.store.GetRepo(id)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			return
		}
		log.Println(err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	repoName := repository.URLHTTP[len("https://github.com/"):]
	suggestion, err := repo.Suggest(repoName)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(suggestion)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
	}
}

func (s *server) setTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.URL.Path[len("/tag/"):])
	if err != nil {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	repository, err := s.store.GetRepo(id)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			return
		}
		log.Println(err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	tags := r.FormValue("tags")
	ss := strings.Split(tags, ",")

	repository.SetTags(ss...)
	err = s.store.UpdateTags(repository)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func main() {
	dbPath := os.Getenv("REPOTAGGER_DBPATH")
	if dbPath == "" {
		dbPath = "repoTagger.db"
	}
	db, err := sqlite.New(dbPath)
	if err != nil {
		panic("failed to connect to db")
	}
	s := &server{store: db}

	port := os.Getenv("REPOTAGGER_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/repos/", s.getRepos)
	http.HandleFunc("/search/", s.search)
	http.HandleFunc("/suggest/", s.suggest)
	http.HandleFunc("/tag/", s.setTag)
	http.ListenAndServe(":"+port, nil)
}
