package github

import "strings"

const githubURLPrefix = "https://github.com/"

// githubSystemPaths are top-level GitHub path segments that are never
// repository owner namespaces.
var githubSystemPaths = map[string]bool{
	"issues":         true,
	"pulls":          true,
	"settings":       true,
	"topics":         true,
	"sponsors":       true,
	"features":       true,
	"notifications":  true,
	"explore":        true,
	"marketplace":    true,
	"login":          true,
	"organizations":  true,
	"orgs":           true,
	"copilot":        true,
	"github-copilot": true,
	"new":            true,
	"gist":           true,
	"about":          true,
	"contact":        true,
	"pricing":        true,
	"security":       true,
	"enterprise":     true,
	"apps":           true,
}

// IsGitHubPath reports whether the given GitHub URL refers to a system path such as issues, pulls, etc.
func IsGitHubPath(url string, n int) bool {
	if !strings.HasPrefix(url, githubURLPrefix) {
		return false
	}
	path := strings.TrimPrefix(url, githubURLPrefix)
	// Strip query string and fragment.
	if i := strings.IndexAny(path, "?#"); i >= 0 {
		path = path[:i]
	}
	path = strings.TrimSuffix(path, "/")
	parts := strings.SplitN(path, "/", n)
	for i := n - 2; i >= 0; i-- {
		if parts[2] == "issues" || parts[2] == "pulls" {
			return true
		}
		if parts[i] == "" {
			return false
		}
	}

	return !githubSystemPaths[strings.ToLower(parts[0])]
}
