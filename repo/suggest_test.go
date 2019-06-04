package repo

import "testing"

func TestSuggest(t *testing.T) {
	r1 := &repoSuggest{Stars: 1000}
	r1.License.Key = "mit"
	r1.Owner.Type = "User"
	r2 := &repoSuggest{Stars: 90000}
	r3 := &repoSuggest{Stars: 5000}
	r3.Owner.Type = "Company"
	tt := []struct {
		name     string
		rSug     *repoSuggest
		expected []string
	}{
		{"notPop", r1, []string{"not-popular", "User-owner", "mit"}},
		{"veryPop", r2, []string{"very-popular"}},
		{"pop", r3, []string{"popular", "Company-owner"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			suggestions := suggest(tc.rSug)

			if len(suggestions) != len(tc.expected) {
				t.Fatalf("wrong number of suggestions, expected %d; got %d",
					len(tc.expected), len(suggestions))
			}

			for i, sug := range suggestions {
				if sug != tc.expected[i] {
					t.Fatalf("wrong suggestion, expected: %s; got %s", tc.expected[i], sug)
				}
			}
		})
	}
}
