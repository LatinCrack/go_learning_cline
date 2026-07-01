package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Wordlist holds a list of entries loaded from a dictionary file.
type Wordlist struct {
	Name    string   // Descriptive name (e.g. "users", "passwords").
	Path    string   // Source file path.
	Entries []string // The loaded entries.
}

// LoadWordlist reads a wordlist file and returns a Wordlist.
// Each non-empty, non-comment line (starting with #) becomes an entry.
// Leading/trailing whitespace is trimmed.
func LoadWordlist(name, path string) (*Wordlist, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open wordlist %s: %w", path, err)
	}
	defer f.Close()

	var entries []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // 1MB max line

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read wordlist %s at line %d: %w", path, lineNum, err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("wordlist %s is empty after filtering", path)
	}

	return &Wordlist{
		Name:    name,
		Path:    path,
		Entries: entries,
	}, nil
}

// Len returns the number of entries in the wordlist.
func (w *Wordlist) Len() int {
	return len(w.Entries)
}

// EntryPair represents a username/password combination to test.
type EntryPair struct {
	Username string
	Password string
	Index    int // Sequential index for tracking progress.
}

// GeneratePairs creates all username/password combinations as a channel.
// Uses cartesian product: for each user, iterate all passwords.
// The returned channel is closed when all pairs have been sent.
func GeneratePairs(users, passwords *Wordlist) <-chan EntryPair {
	total := users.Len() * passwords.Len()
	pairs := make(chan EntryPair, min(1000, total))

	go func() {
		defer close(pairs)
		idx := 0
		for _, user := range users.Entries {
			for _, pass := range passwords.Entries {
				pairs <- EntryPair{
					Username: user,
					Password: pass,
					Index:    idx,
				}
				idx++
			}
		}
	}()

	return pairs
}

// GeneratePairsSinglePassword creates pairs with a single password applied to all users.
// Used when only a password list is provided (common for known-username scenarios).
func GeneratePairsSinglePassword(users *Wordlist, password string) <-chan EntryPair {
	total := users.Len()
	pairs := make(chan EntryPair, min(1000, total))

	go func() {
		defer close(pairs)
		for idx, user := range users.Entries {
			pairs <- EntryPair{
				Username: user,
				Password: password,
				Index:    idx,
			}
		}
	}()

	return pairs
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}