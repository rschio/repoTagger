package storage

import "github.com/rschio/repoTagger/repo"

// Storage is the interface that abstract the data storage.
type Storage interface {
	// InsertRepo insert the repository into storage.
	InsertRepo(*repo.Repo) error
	// GetReposByTag search all the repositories that has
	// a tag starting with string and return the repositories
	// slice and error.
	GetReposByTag(string) ([]*repo.Repo, error)
	// UpdateTags delete the old tags of r and
	// set the new ones.
	UpdateTags(r *repo.Repo) error
	// GetRepo returns the repo by id.
	GetRepo(id int) (*repo.Repo, error)
}
