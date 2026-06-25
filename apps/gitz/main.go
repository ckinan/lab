package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ScanDirs   []string `yaml:"scan_dirs"`
	ExactRepos []string `yaml:"exact_repos"`
	Sort       string   `yaml:"sort"`
}

type RepoStatus struct {
	Name             string
	Branch           string
	HasChanges       bool
	HasPendingPushes bool
	HasStashes       bool
	LastCommit       string
	LastCommitTime   int64
	Error            error
}

func main() {
	sortFlag := flag.String("sort", "", "Sort by 'name' or 'lastCommit'")
	flag.Parse()

	var targetRepos []string
	var sortBy = "name" // Default sort

	// Try to read config
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".gitz.yaml")
	configData, err := os.ReadFile(configPath)

	if err != nil {
		if configDir, errConfig := os.UserConfigDir(); errConfig == nil {
			configPath = filepath.Join(configDir, "gitz", "config.yaml")
			configData, err = os.ReadFile(configPath)
		}
	}

	if err == nil {
		var cfg Config
		if err := yaml.Unmarshal(configData, &cfg); err == nil {
			for _, dir := range cfg.ScanDirs {
				expanded := expandHome(dir, home)
				targetRepos = append(targetRepos, getReposFromDir(expanded)...)
			}
			for _, repo := range cfg.ExactRepos {
				targetRepos = append(targetRepos, expandHome(repo, home))
			}
			if cfg.Sort != "" {
				sortBy = cfg.Sort
			}
		}
	}

	// CLI path overrides YAML paths
	if len(flag.Args()) > 0 {
		targetRepos = getReposFromDir(flag.Arg(0))
	}

	// CLI sort flag overrides YAML sort
	if *sortFlag != "" {
		sortBy = *sortFlag
	}

	// If no config paths and no CLI path, default to current directory
	if len(targetRepos) == 0 {
		targetRepos = getReposFromDir(".")
	}

	targetRepos = unique(targetRepos)

	if len(targetRepos) == 0 {
		fmt.Println("No git repositories found.")
		return
	}

	var wg sync.WaitGroup
	results := make(chan RepoStatus, len(targetRepos))

	for _, repoPath := range targetRepos {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			results <- getRepoStatus(filepath.Base(path), path)
		}(repoPath)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var statuses []RepoStatus
	for st := range results {
		statuses = append(statuses, st)
	}

	// Sort the statuses
	sort.Slice(statuses, func(i, j int) bool {
		if sortBy == "lastCommit" {
			// Sort descending by time (newest first)
			if statuses[i].LastCommitTime == statuses[j].LastCommitTime {
				return strings.ToLower(statuses[i].Name) < strings.ToLower(statuses[j].Name)
			}
			return statuses[i].LastCommitTime > statuses[j].LastCommitTime
		}
		// Default: Sort alphabetically by name
		return strings.ToLower(statuses[i].Name) < strings.ToLower(statuses[j].Name)
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "REPOSITORY\tBRANCH\tUNCOMMITTED\tUNPUSHED\tSTASHED\tLAST COMMIT")
	for _, st := range statuses {
		if st.Error != nil {
			fmt.Fprintf(w, "%s\tError: %v\t\t\t\t\n", st.Name, st.Error)
			continue
		}

		changes := "-"
		if st.HasChanges {
			changes = "Yes"
		}
		pushes := "-"
		if st.HasPendingPushes {
			pushes = "Yes"
		}
		stashes := "-"
		if st.HasStashes {
			stashes = "Yes"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			st.Name, st.Branch, changes, pushes, stashes, st.LastCommit)
	}
	w.Flush()
}

func getReposFromDir(dir string) []string {
	var repos []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return repos
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		repoPath := filepath.Join(dir, entry.Name())
		gitPath := filepath.Join(repoPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			repos = append(repos, repoPath)
		}
	}
	return repos
}

func expandHome(path, home string) string {
	if strings.HasPrefix(path, "~") {
		return filepath.Join(home, path[1:])
	}
	return path
}

func unique(strSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func getRepoStatus(name, path string) RepoStatus {
	st := RepoStatus{Name: name}

	out, _ := runGit(path, "branch", "--show-current")
	st.Branch = strings.TrimSpace(out)
	if st.Branch == "" {
		out, _ = runGit(path, "rev-parse", "--abbrev-ref", "HEAD")
		st.Branch = strings.TrimSpace(out)
	}

	out, _ = runGit(path, "status", "--porcelain")
	st.HasChanges = len(strings.TrimSpace(out)) > 0

	out, _ = runGit(path, "status", "-sb")
	st.HasPendingPushes = strings.Contains(out, "ahead")

	out, _ = runGit(path, "stash", "list")
	st.HasStashes = len(strings.TrimSpace(out)) > 0

	// Get formatted date and unix timestamp
	out, _ = runGit(path, "log", "-1", "--date=iso-local", "--format=%cd (%h)|%ct")
	parts := strings.Split(strings.TrimSpace(out), "|")
	if len(parts) == 2 {
		st.LastCommit = parts[0]
		st.LastCommitTime, _ = strconv.ParseInt(parts[1], 10, 64)
	} else {
		st.LastCommit = strings.TrimSpace(out)
	}

	return st
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}
