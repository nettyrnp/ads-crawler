package entity

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestParseProvider(t *testing.T) {
	t.Parallel()

	t.Run("comments", testComments())
	t.Run("simple, 4 parts", testParse())
}

type entry struct {
	actual   string
	expected interface{}
}

func testComments() func(t *testing.T) {
	tcases := []entry{
		{"#Some comment", ""},
		{"# Some comment  ", ""},
		{"#Some comment  ", ""},
		{"  #Some comment", "  "},
		{" #Some comment ; abc", "  abc"},
		{"google.com, pub-5231479214411897, RESELLER, f08c47fec0942fa0 #Some comment", "google.com, pub-5231479214411897, RESELLER, f08c47fec0942fa0 "},
	}

	return func(t *testing.T) {
		for _, tc := range tcases {
			actual := reComment.ReplaceAllString(tc.actual, "")
			require.Equal(t, actual, tc.expected)
		}
	}
}

func testParse() func(t *testing.T) {
	tcases := []entry{
		{" google.com,   pub-5231479214411897, ReselLER, f08c47fec0942fa0", &Provider{
			DomainName:  "google.com",
			AccountID:   "pub-5231479214411897",
			AccountType: "reseller",
			CertAuthID:  "f08c47fec0942fa0",
		}},
		{"cnn.com,  pub-5231479214411897, DIRECT", &Provider{
			DomainName:  "cnn.com",
			AccountID:   "pub-5231479214411897",
			AccountType: "direct",
		}},
		{"google.com, pub-5231479214411897, ReselLER, f08c47fec0942fa0 #Some comment", &Provider{
			DomainName:  "google.com",
			AccountID:   "pub-5231479214411897",
			AccountType: "reseller",
			CertAuthID:  "f08c47fec0942fa0",
		}},
		{"google.com, pub-5231479214411897, #Some comment; ReselLER, f08c47fec0942fa0 #Another comment", &Provider{
			DomainName:  "google.com",
			AccountID:   "pub-5231479214411897",
			AccountType: "reseller",
			CertAuthID:  "f08c47fec0942fa0",
		}},
	}

	return func(t *testing.T) {
		for _, tc := range tcases {
			actual := reComment.ReplaceAllString(tc.actual, "")
			actual = strings.ToLower(actual)
			if !reProvider.MatchString(actual) {
				t.FailNow()
			}
			actualProvider, err := ParseProvider(actual)
			require.NoError(t, err)
			require.Equal(t, actualProvider, tc.expected)
		}
	}
}
