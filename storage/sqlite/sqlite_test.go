package sqlite

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/rschio/repoTagger/repo"
)

func TestNew(t *testing.T) {
	f, err := ioutil.TempFile(".", "testNewDb")
	if err != nil {
		t.Fatalf("failed to create temp file")
	}
	defer os.Remove(f.Name())

	db, err := New(f.Name())
	if err != nil {
		t.Errorf("database should be created")
	}

	err = db.(*service).Close()
	if err != nil {
		t.Errorf("datbase should close: %v", err)
	}
}

func tagsEq(t1, t2 []string) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i, tag1 := range t1 {
		if tag1 != t2[i] {
			return false
		}
	}
	return true
}

func repoEq(r1, r2 *repo.Repo) bool {
	if r1.ID != r2.ID || r1.Name != r2.Name ||
		r1.Desc != r2.Desc || r1.URLHTTP != r2.URLHTTP {
		return false
	}
	return tagsEq(r1.Tags, r2.Tags)
}

func TestInsertGet(t *testing.T) {
	f, err := ioutil.TempFile(".", "testNewDb")
	if err != nil {
		t.Fatalf("failed to create temp file")
	}
	defer os.Remove(f.Name())

	db, err := New(f.Name())
	if err != nil {
		t.Errorf("database should be created")
	}

	r1 := &repo.Repo{0, "Foo", "decrpition", "http://something.com", "go", []string{"H", "e"}}
	r2 := &repo.Repo{4, "Bar", "ha", "http://something.com", "", []string{}}
	err = db.InsertRepo(r1)
	if err != nil {
		t.Errorf("failed to insert repo: %v", err)
	}

	r3, err := db.GetRepo(r1.ID)
	if err != nil {
		t.Fatalf("failed to get repo %d: %v", r1.ID, err)
	}

	if !repoEq(r1, r3) {
		t.Fatalf("got different repos: %v, %v", r1, r3)
	}

	_, err = db.GetRepo(1)
	if err == nil {
		t.Fatalf("got invalid repo")
	}

	err = db.InsertRepo(r2)
	if err != nil {
		t.Errorf("failed to insert repo: %v", err)
	}

	r4, err := db.GetRepo(r2.ID)
	if err != nil {
		t.Fatalf("failed to get repo %d: %v", r2.ID, err)
	}

	if !repoEq(r2, r4) {
		t.Errorf("got different repos: %v, %v", r2, r4)
	}

}

func TestGetReposByTag(t *testing.T) {
	f, err := ioutil.TempFile(".", "testNewDb")
	if err != nil {
		t.Fatalf("failed to create temp file")
	}
	defer os.Remove(f.Name())

	db, err := New(f.Name())
	if err != nil {
		t.Errorf("database should be created")
	}

	r1 := &repo.Repo{0, "Foo", "decrpition", "http://something.com", "go",
		[]string{"document", "docker"}}
	r2 := &repo.Repo{4, "Bar", "ha", "http://something.com", "", []string{}}
	db.InsertRepo(r1)
	db.InsertRepo(r2)

	rs, err := db.GetReposByTag("")
	if err != nil {
		t.Fatalf("failed to get repos by tag: %v", err)
	}
	if len(rs) != 2 {
		t.Fatalf("got wrong number of repos by tag: %v", err)
	}
	if !repoEq(rs[0], r1) {
		if !repoEq(rs[0], r2) || !repoEq(rs[1], r1) {
			t.Fatalf("got wrong repos")
		}
	} else if !repoEq(rs[1], r2) {
		t.Fatalf("got wrong repos")
	}

	rs, err = db.GetReposByTag("doc")
	if err != nil {
		t.Fatalf("failed to get repos by tag: %v", err)
	}
	if len(rs) != 1 {
		t.Fatalf("got wrong number of repos by tag: %v", err)
	}
	if !repoEq(rs[0], r1) {
		t.Fatalf("got wrong repos")
	}

}

func TestUpdateTags(t *testing.T) {
	f, err := ioutil.TempFile(".", "testNewDb")
	if err != nil {
		t.Fatalf("failed to create temp file")
	}
	defer os.Remove(f.Name())

	db, err := New(f.Name())
	if err != nil {
		t.Errorf("database should be created")
	}

	r1 := &repo.Repo{0, "Foo", "decrpition", "http://something.com", "go",
		[]string{"document", "docker"}}
	r2 := &repo.Repo{4, "Bar", "ha", "http://something.com", "", []string{}}
	db.InsertRepo(r1)
	db.InsertRepo(r2)

	r1.Tags = []string{"notDocker", "notDocument", "otherThing"}
	r2.Tags = []string{"tag"}
	err = db.UpdateTags(r1)
	if err != nil {
		t.Fatalf("failet to update tags: %v", err)
	}
	err = db.UpdateTags(r2)
	if err != nil {
		t.Fatalf("failed to update tags: %v", err)
	}

	rs, err := db.GetReposByTag("doc")
	if err != nil {
		t.Fatalf("failed to get repos: %v", err)
	}

	if len(rs) > 0 {
		t.Fatalf("got repos bu shouldn't")
	}

	r3, err := db.GetRepo(r1.ID)
	if err != nil {
		t.Fatalf("failed to get repo")
	}
	r4, err := db.GetRepo(r2.ID)
	if err != nil {
		t.Fatalf("failed to get repo")
	}

	if !tagsEq(r1.Tags, r3.Tags) {
		t.Fatalf("got wrong tags")
	}
	if !tagsEq(r2.Tags, r4.Tags) {
		t.Fatalf("got wrong tags")
	}

}
