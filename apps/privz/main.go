package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

type Config struct {
	IgnoreCategories []string `yaml:"ignore_categories"`
	IgnoreKeywords   []string `yaml:"ignore_keywords"`
	SortBy           string   `yaml:"sort_by"` // category or name
}

type Device struct {
	Name     string
	Category string // CAMERA, AUDIO/MIC, DISPLAY, KEYBOARD, MOUSE/TOUCH, USB HARDWARE, INPUT/OTHER
	Path     string
	Details  string
}

type MediaState struct {
	IsConnected bool
	IsOn        bool
	ActivePIDs  []int
	ActiveNames []string
}

func main() {
	flag.Parse()
	cfg := loadConfig()
	printSnapshot(cfg)
}

func loadConfig() Config {
	cfg := Config{
		SortBy: "category",
	}

	home, _ := os.UserHomeDir()
	configPaths := []string{
		filepath.Join(home, ".privz.yaml"),
		filepath.Join(home, ".config", "privz", "config.yaml"),
	}

	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			_ = yaml.Unmarshal(data, &cfg)
			break
		}
	}

	return cfg
}

func printSnapshot(cfg Config) {
	camState, micState := getMediaStates()
	allDevices := scanAllDevices()

	ignoredCats := make(map[string]bool)
	for _, c := range cfg.IgnoreCategories {
		ignoredCats[strings.ToUpper(c)] = true
	}

	var filtered []Device
	for _, dev := range allDevices {
		if ignoredCats[dev.Category] {
			continue
		}
		skip := false
		for _, kw := range cfg.IgnoreKeywords {
			if strings.Contains(strings.ToLower(dev.Name), strings.ToLower(kw)) {
				skip = true
				break
			}
		}
		if !skip {
			filtered = append(filtered, dev)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Println("=== PRIVZ HARDWARE & PRIVACY RADAR ===")
	fmt.Println()

	// Special Section for Camera & Mic
	fmt.Println("--- SPECIAL PRIVACY SECTION (CAMERA & MICROPHONE) ---")
	fmt.Fprintf(w, "DEVICE\tSTATUS\tCONNECTION\tACTIVE PROCESS\n")

	camStatus := "[OFF]"
	if camState.IsOn {
		camStatus = "[ON] (STREAMING ACTIVE)"
	}
	camConn := "Disconnected"
	if camState.IsConnected {
		camConn = "Connected"
	}
	camProc := "-"
	if len(camState.ActiveNames) > 0 {
		camProc = fmt.Sprintf("%s (PIDs: %v)", strings.Join(camState.ActiveNames, ", "), camState.ActivePIDs)
	}
	fmt.Fprintf(w, "Camera\t%s\t%s\t%s\n", camStatus, camConn, camProc)

	micStatus := "[OFF]"
	if micState.IsOn {
		micStatus = "[ON] (RECORDING ACTIVE)"
	}
	micConn := "Disconnected"
	if micState.IsConnected {
		micConn = "Connected"
	}
	micProc := "-"
	if len(micState.ActiveNames) > 0 {
		micProc = fmt.Sprintf("%s (PIDs: %v)", strings.Join(micState.ActiveNames, ", "), micState.ActivePIDs)
	}
	fmt.Fprintf(w, "Microphone\t%s\t%s\t%s\n", micStatus, micConn, micProc)
	w.Flush()

	fmt.Println()
	fmt.Println("--- ALL CONNECTED HARDWARE DEVICES ---")
	fmt.Fprintf(w, "CATEGORY\tDEVICE NAME\tSYSTEM PATH\tDETAILS\n")

	sort.Slice(filtered, func(i, j int) bool {
		if strings.ToLower(cfg.SortBy) == "name" {
			if filtered[i].Name == filtered[j].Name {
				return filtered[i].Category < filtered[j].Category
			}
			return strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		}
		// Default: sort by category
		if filtered[i].Category == filtered[j].Category {
			return strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		}
		return filtered[i].Category < filtered[j].Category
	})

	for _, dev := range filtered {
		fmt.Fprintf(w, "[%s]\t%s\t%s\t%s\n", dev.Category, dev.Name, dev.Path, dev.Details)
	}
	w.Flush()
}

func getMediaStates() (MediaState, MediaState) {
	camState := MediaState{}
	micState := MediaState{}

	if matches, _ := filepath.Glob("/dev/video*"); len(matches) > 0 {
		camState.IsConnected = true
	}
	if cards, err := os.ReadFile("/proc/asound/cards"); err == nil && len(cards) > 0 {
		micState.IsConnected = true
	}

	entries, _ := os.ReadDir("/proc")
	for _, p := range entries {
		if !p.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(p.Name())
		if err != nil {
			continue
		}

		fdPath := filepath.Join("/proc", p.Name(), "fd")
		fds, err := os.ReadDir(fdPath)
		if err != nil {
			continue
		}

		hasCam := false
		hasMic := false

		for _, fd := range fds {
			target, err := os.Readlink(filepath.Join(fdPath, fd.Name()))
			if err != nil {
				continue
			}

			if strings.HasPrefix(target, "/dev/video") {
				hasCam = true
			}
			if strings.HasPrefix(target, "/dev/snd/pcm") && strings.HasSuffix(target, "c") {
				hasMic = true
			}
		}

		if hasCam || hasMic {
			comm, _ := os.ReadFile(filepath.Join("/proc", p.Name(), "comm"))
			procName := strings.TrimSpace(string(comm))
			if procName == "" {
				procName = fmt.Sprintf("PID %d", pid)
			}

			if hasCam {
				camState.IsOn = true
				camState.ActivePIDs = append(camState.ActivePIDs, pid)
				camState.ActiveNames = append(camState.ActiveNames, procName)
			}
			if hasMic {
				micState.IsOn = true
				micState.ActivePIDs = append(micState.ActivePIDs, pid)
				micState.ActiveNames = append(micState.ActiveNames, procName)
			}
		}
	}

	camState.ActiveNames = uniqueStrings(camState.ActiveNames)
	micState.ActiveNames = uniqueStrings(micState.ActiveNames)

	return camState, micState
}

func scanAllDevices() []Device {
	var devices []Device

	// 1. Cameras
	if matches, _ := filepath.Glob("/sys/class/video4linux/video*"); len(matches) > 0 {
		for _, m := range matches {
			nameBytes, _ := os.ReadFile(filepath.Join(m, "name"))
			name := strings.TrimSpace(string(nameBytes))
			if name == "" {
				name = filepath.Base(m)
			}
			devices = append(devices, Device{
				Name:     name,
				Category: "CAMERA",
				Path:     m,
				Details:  "Video4Linux device",
			})
		}
	}

	// 2. Sound Cards / Microphones
	if cards, err := os.ReadFile("/proc/asound/cards"); err == nil {
		lines := strings.Split(string(cards), "\n")
		for _, l := range lines {
			if strings.Contains(l, ": ") && !strings.HasPrefix(strings.TrimSpace(l), "---") {
				parts := strings.SplitN(l, ": ", 2)
				if len(parts) == 2 {
					devices = append(devices, Device{
						Name:     strings.TrimSpace(parts[1]),
						Category: "AUDIO/MIC",
						Path:     "/proc/asound/cards",
						Details:  "ALSA Sound Card",
					})
				}
			}
		}
	}

	// 3. Displays / Monitors
	if matches, _ := filepath.Glob("/sys/class/drm/card*-*"); len(matches) > 0 {
		for _, m := range matches {
			statusBytes, _ := os.ReadFile(filepath.Join(m, "status"))
			status := strings.TrimSpace(string(statusBytes))
			if status == "connected" {
				devices = append(devices, Device{
					Name:     filepath.Base(m),
					Category: "DISPLAY",
					Path:     m,
					Details:  "DRM Display Connector (Connected)",
				})
			}
		}
	}

	// 4. Input Devices (Keyboards, Mice)
	if data, err := os.ReadFile("/proc/bus/input/devices"); err == nil {
		blocks := strings.Split(string(data), "\n\n")
		for _, b := range blocks {
			var name, handler string
			lines := strings.Split(b, "\n")
			for _, l := range lines {
				if strings.HasPrefix(l, "N: Name=") {
					name = strings.Trim(strings.TrimPrefix(l, "N: Name="), "\"")
				}
				if strings.HasPrefix(l, "H: Handlers=") {
					handler = strings.TrimPrefix(l, "H: Handlers=")
				}
			}
			if name != "" {
				cat := "INPUT/OTHER"
				lower := strings.ToLower(name)
				if strings.Contains(lower, "keyboard") {
					cat = "KEYBOARD"
				} else if strings.Contains(lower, "mouse") || strings.Contains(lower, "track") || strings.Contains(lower, "touch") {
					cat = "MOUSE/TOUCH"
				}
				devices = append(devices, Device{
					Name:     name,
					Category: cat,
					Path:     handler,
					Details:  "evdev input device",
				})
			}
		}
	}

	// 5. USB Devices
	if matches, _ := filepath.Glob("/sys/bus/usb/devices/*"); len(matches) > 0 {
		for _, m := range matches {
			prodBytes, err := os.ReadFile(filepath.Join(m, "product"))
			if err != nil {
				continue
			}
			prod := strings.TrimSpace(string(prodBytes))
			mfgBytes, _ := os.ReadFile(filepath.Join(m, "manufacturer"))
			mfg := strings.TrimSpace(string(mfgBytes))
			full := prod
			if mfg != "" && !strings.Contains(prod, mfg) {
				full = mfg + " " + prod
			}
			devices = append(devices, Device{
				Name:     full,
				Category: "USB HARDWARE",
				Path:     m,
				Details:  "USB Bus Device",
			})
		}
	}

	return devices
}

func uniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range slice {
		if _, val := keys[entry]; !val {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
