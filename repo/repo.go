package repo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	reMaxPage = regexp.MustCompile(`<https:\/\/api.github.com\/user\/(\d+)\/starred\?page=(\d+)>; rel="last"`)
	client    = &http.Client{}
)

// Repo stores repository info.
type Repo struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Desc    string   `json:"description"`
	URLHTTP string   `json:"html_url"`
	Lang    string   `json:"language"`
	Tags    []string `json:"tags,omitempty"`
}

// NotFoundErr is used to know when
// repo is not found in GitHub.
type NotFoundErr int

func (e NotFoundErr) Error() string {
	return fmt.Sprintf("not found")
}

// SetTags discard the last tags map and set
// the new one.
func (r *Repo) SetTags(tags ...string) {
	r.Tags = make([]string, 0, len(tags))
	duplicated := make(map[string]struct{})
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := duplicated[tag]; ok {
			continue
		}
		duplicated[tag] = struct{}{}
		r.Tags = append(r.Tags, tag)
	}
}

// getLastPage returns the last page number or 0 if there is
// an error, pages start from 1.
func getLastPage(res *http.Response) int {
	if len(res.Header["Link"]) < 1 {
		return 0
	}
	reSs := reMaxPage.FindStringSubmatch(res.Header["Link"][0])
	if len(reSs) < 1 {
		return 0
	}
	// discard error, 0 will be treated as error.
	page, _ := strconv.Atoi(reSs[len(reSs)-1])
	return page
}

// GetGithubRepos returns all github repositories of user.
func GetGithubRepos(user string) ([]*Repo, error) {
	urlFormat := "https://api.github.com/users/" + user + "/starred?page=%d"
	return getGithubRepos(urlFormat)
}

func getGithubRepos(urlFormat string) ([]*Repo, error) {
	url := fmt.Sprintf(urlFormat, 1)
	res, err := requestPage(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		var notFound NotFoundErr
		return nil, notFound
	}

	nPages := getLastPage(res)
	allRepos := make([]*Repo, 0)
	if nPages > 1 {
		rs := getAllPages(urlFormat, nPages)
		allRepos = append(allRepos, rs...)
	}

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	repos, err := unmarshalRepos(bs)
	if err != nil {
		return nil, err
	}
	allRepos = append(allRepos, repos...)
	return allRepos, nil
}

// requestPage request the page with github header.
func requestPage(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	return client.Do(req)
}

func getPageBody(url string) ([]byte, error) {
	res, err := requestPage(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func unmarshalRepos(data []byte) ([]*Repo, error) {
	starreds := make([]*Repo, 0)
	if err := json.Unmarshal(data, &starreds); err != nil {
		return nil, err
	}
	return starreds, nil
}

func extractRepo(repoCh chan<- []*Repo, limit chan struct{}, url string) {
	// release space to another go routine can execute.
	defer func() { <-limit }()
	body, err := getPageBody(url)
	if err != nil {
		repoCh <- nil
		return
	}

	repos, err := unmarshalRepos(body)
	if err != nil {
		repoCh <- nil
		return
	}

	repoCh <- repos
}

// getAllPages request pages [2, nPages] concurrently and extract
// the repositories.
func getAllPages(urlFormat string, nPages int) []*Repo {
	repoCh := make(chan []*Repo)
	// limit the go routines.
	limit := make(chan struct{}, 100)
	done := make(chan struct{})

	allRepos := make([]*Repo, 0)
	go func() {
		for i := 2; i <= nPages; i++ {
			if rs := <-repoCh; rs != nil {
				allRepos = append(allRepos, rs...)
			}
		}
		done <- struct{}{}
	}()

	for i := 2; i <= nPages; i++ {
		url := fmt.Sprintf(urlFormat, i)
		// if there are less then limit of
		// go routines executing take your pass.
		limit <- struct{}{}
		go extractRepo(repoCh, limit, url)
	}

	<-done
	return allRepos
}
