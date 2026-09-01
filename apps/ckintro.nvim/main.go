package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

/*
example of the config file:

```yaml
scan_dirs:
  - ~/code

exactDirs:
  - ~/code/lab/apps/ckintro.nvim
```
*/
type Config struct {
	ScanDirs  []string `yaml:"scan_dirs"`
	ExactDirs []string `yaml:"exact_dirs"`
}

func main() {
	var dirsToShow []string

	// Try to read config
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".ckintro.nvim.yaml")
	configData, err := os.ReadFile(configPath)

	if err != nil {
		if configDir, errConfig := os.UserConfigDir(); errConfig == nil {
			configPath = filepath.Join(configDir, "ckintro.nvim", "config.yaml")
			configData, err = os.ReadFile(configPath)
		}
	}

	if err == nil {
		var cfg Config
		if err := yaml.Unmarshal(configData, &cfg); err == nil {
			for _, dir := range cfg.ScanDirs {
				expanded := expandHome(dir, home)
				dirsToShow = append(dirsToShow, lsDirs(expanded)...)
			}
			for _, repo := range cfg.ExactDirs {
				dirsToShow = append(dirsToShow, expandHome(repo, home))
			}
		}
	}

	if len(dirsToShow) == 0 {
		fmt.Println("no dirs to show...")
		return
	}

	for _, d := range dirsToShow {
		fmt.Println(d)
	}
}

func expandHome(path, home string) string {
	if strings.HasPrefix(path, "~") {
		return filepath.Join(home, path[1:])
	}
	return path
}

func lsDirs(dir string) []string {
	var dirs []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return dirs
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPath := filepath.Join(dir, entry.Name())
		if _, err := os.Stat(dirPath); err == nil {
			dirs = append(dirs, dirPath)
		}
	}
	return dirs
}
