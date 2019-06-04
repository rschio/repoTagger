package repo

import (
	"encoding/json"
)

type repoSuggest struct {
	Stars   int `json:"stargazers_count"`
	License struct {
		Key string `jsont:"key"`
	} `json:"license"`
	Owner struct {
		Type string `json:"type"`
	} `json:"owner"`
}

// Suggest suggests tags to repository.
func Suggest(repoName string) ([]string, error) {
	data, err := getPageBody("https://api.github.com/repos/" + repoName)
	if err != nil {
		return nil, err
	}

	rSug := &repoSuggest{}
	err = json.Unmarshal(data, rSug)
	if err != nil {
		return nil, err
	}

	return suggest(rSug), nil
}

func suggest(rSug *repoSuggest) []string {
	suggestions := make([]string, 1)
	if rSug.Stars > 10000 {
		suggestions[0] = "very-popular"
	} else if rSug.Stars > 1000 {
		suggestions[0] = "popular"
	} else {
		suggestions[0] = "not-popular"
	}

	if rSug.Owner.Type != "" {
		suggestions = append(suggestions, rSug.Owner.Type+"-owner")
	}
	if rSug.License.Key != "" {
		suggestions = append(suggestions, rSug.License.Key)
	}

	return suggestions
}
