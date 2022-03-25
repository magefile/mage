//go:build go1.18

package mage

import (
	"encoding/json"
	"net/http"
	dbg "runtime/debug"
	"sort"
	"strings"
	"time"
)

// set by ldflags when you "mage build"
var (
	commitHash = "<not set>"
	timestamp  = "<not set>"
	gitTag     = "<not set>"
)

func init() {
	info, ok := dbg.ReadBuildInfo()
	if !ok {
		return
	}
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			commitHash = kv.Value
		case "vcs.time":
			timestamp = kv.Value
		}
	}
}

// uses the commit hash to get the git tag (if one exists) from github
func getTag() string {
	debug.Println("requesting tag info from github via https://api.github.com/repos/magefile/mage/git/refs/tags")

	http.DefaultClient.Timeout = 300 * time.Millisecond
	resp, err := http.DefaultClient.Get("https://api.github.com/repos/magefile/mage/git/refs/tags")
	if err != nil {
		debug.Println("unable to request tag info from github:", err)
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		debug.Println("bad response requesting mage tag info from github:", resp.StatusCode)
		return ""
	}

	var tags []tag
	err = json.NewDecoder(resp.Body).Decode(&tags)
	if err != nil {
		debug.Println("error unmarshalling mage tag info from github:", err)
		return ""
	}

	var found []string
	for _, t := range tags {
		if t.Object.Sha == commitHash && strings.HasPrefix(t.Ref, "refs/tags/") {
			found = append(found, t.Ref[len("refs/tags/"):])
		}
	}
	if len(found) == 0 {
		debug.Println("no git tag found for commit hash:", commitHash)
		return "<no tag for this commit>"
	}
	if len(found) == 1 {
		return found[0]
	}
	// more than one tag for this commit, report the highest tag number.
	sort.Strings(found)
	return found[len(found)-1]
}

type tag struct {
	Ref    string `json:"ref"`
	NodeID string `json:"node_id"`
	URL    string `json:"url"`
	Object struct {
		Sha  string `json:"sha"`
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"object"`
}
