package repo

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testServer struct {
	t *testing.T
}

func (ts *testServer) respJSON(w http.ResponseWriter, r *http.Request) {
	bs, _ := ioutil.ReadFile("mock.json")
	w.Header().Set("Link", `<https://api.github.com/user/0/starred?page=2>; rel="next",
    						<https://api.github.com/user/0/starred?page=500>; rel="last"`)
	w.Write(bs)
}

func (ts *testServer) notFoundPage(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(404), http.StatusNotFound)
}

func TestGetGithubRepos(t *testing.T) {
	ts := &testServer{t}

	tt := []struct {
		name        string
		handlerFn   http.HandlerFunc
		expectedErr error
	}{
		{"default", ts.respJSON, nil},
		{"not found", ts.notFoundPage, NotFoundErr(0)},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(tc.handlerFn)

			_, err := getGithubRepos(s.URL + "/%d")
			if err != tc.expectedErr {
				t.Fatalf("failed to get repos: %q", err)
			}
			s.Close()
		})
	}

}

func TestUnmarshalRepos(t *testing.T) {
	bs, err := ioutil.ReadFile("mock.json")
	if err != nil {
		t.Fatalf("failed to read mock json: %q", err)
	}
	rs, err := unmarshalRepos(bs)
	if err != nil {
		t.Fatalf("failed to unmarshal mock json: %q", err)
	}
	if len(rs) != 1 {
		t.Fatalf("expected 1 repos; got %d", len(rs))
	}
	rs, err = unmarshalRepos(nil)
	if err == nil {
		t.Fatalf("unmarshaled nil []byte")
	}
}

func TestSetTags(t *testing.T) {
	tt := []struct {
		name         string
		r            *Repo
		toSet        []string
		expectedTags []string
	}{
		{"nil prev tag", &Repo{}, []string{"tag1", "other tag"}, []string{"tag1", "other tag"}},
		{"set nil tag", &Repo{Tags: make([]string, 0)}, nil, nil},
		{"empty string", &Repo{Tags: make([]string, 0)}, []string{""}, []string{}},
		{"change tags", &Repo{Tags: []string{"hello", "world"}}, []string{"Foo"}, []string{"Foo"}},
		{"duplicated", &Repo{Tags: []string{"foo", "bar", "something"}}, []string{"Hello", "Hello"}, []string{"Hello"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.r.SetTags(tc.toSet...)
			if tc.expectedTags == nil && len(tc.r.Tags) == 0 {
				return
			}
			if len(tc.r.Tags) != len(tc.expectedTags) {
				t.Fatalf("tags of repo should be %v; got %v, len1 = %d len2 = %d", tc.expectedTags, tc.r.Tags, len(tc.expectedTags), len(tc.r.Tags))
			}

			for i, tag := range tc.r.Tags {
				if tag != tc.expectedTags[i] {
					t.Errorf("tags of repo should be %v; got %v", tc.expectedTags, tc.r.Tags)
				}
			}
		})

	}
}
