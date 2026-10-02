package spoofer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DetectableApp represents an application definition from Discord's catalog.
type DetectableApp struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Executables []DetectableExecutable `json:"executables"`
}

// DetectableExecutable defines a registered process executable name and OS target.
type DetectableExecutable struct {
	IsLauncher bool   `json:"is_launcher"`
	Name       string `json:"name"`
	OS         string `json:"os"`
}

// Registry manages the catalog of detectable Discord applications.
type Registry struct {
	mu   sync.RWMutex
	apps map[string]DetectableApp
}

// LoadRegistry retrieves detectable applications from local cache or official Discord API.
func LoadRegistry(cacheDir string) (*Registry, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cacheFile := filepath.Join(cacheDir, "detectable_cache.json")

	// Try reading fresh cache
	if info, err := os.Stat(cacheFile); err == nil {
		if time.Since(info.ModTime()) < 24*time.Hour {
			if reg, err := loadRegistryFromFile(cacheFile); err == nil && len(reg.apps) > 0 {
				return reg, nil
			}
		}
	}

	// Fetch from Discord API
	reg, err := fetchRegistryFromRemote()
	if err == nil && len(reg.apps) > 0 {
		_ = saveRegistryToFile(cacheFile, reg)
		return reg, nil
	}

	// Fallback to existing stale cache if remote fetch failed
	if reg, errOld := loadRegistryFromFile(cacheFile); errOld == nil && len(reg.apps) > 0 {
		return reg, nil
	}

	return nil, fmt.Errorf("failed to obtain detectable applications registry: %w", err)
}

func loadRegistryFromFile(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var appList []DetectableApp
	if err := json.Unmarshal(data, &appList); err != nil {
		return nil, err
	}

	reg := &Registry{apps: make(map[string]DetectableApp, len(appList))}
	for _, app := range appList {
		reg.apps[app.ID] = app
	}
	return reg, nil
}

func saveRegistryToFile(path string, reg *Registry) error {
	reg.mu.RLock()
	defer reg.mu.RUnlock()

	appList := make([]DetectableApp, 0, len(reg.apps))
	for _, app := range reg.apps {
		appList = append(appList, app)
	}

	data, err := json.Marshal(appList)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func fetchRegistryFromRemote() (*Registry, error) {
	urls := []string{
		"https://discord.com/api/v9/applications/detectable",
		"https://discord.com/api/applications/detectable",
		"https://markterence.github.io/discord-quest-completer/detectable.json",
	}

	client := &http.Client{Timeout: 30 * time.Second}

	for _, url := range urls {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		var appList []DetectableApp
		if err := json.Unmarshal(body, &appList); err == nil && len(appList) > 0 {
			reg := &Registry{apps: make(map[string]DetectableApp, len(appList))}
			for _, app := range appList {
				reg.apps[app.ID] = app
			}
			return reg, nil
		}
	}

	return nil, fmt.Errorf("unable to reach any detectable applications catalog endpoint")
}

// ResolveExecutable finds the most appropriate executable name for the specified application ID and OS.
func (r *Registry) ResolveExecutable(appID string, currentOS string) (exeName string, gameTitle string, err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	app, ok := r.apps[appID]
	if !ok {
		return "", "", fmt.Errorf("application ID %s not found in detectable registry", appID)
	}

	gameTitle = app.Name

	// Map current runtime OS to Discord detectable OS tag
	targetOS := "win32"
	if currentOS == "linux" {
		targetOS = "linux"
	} else if currentOS == "darwin" {
		targetOS = "darwin"
	}

	// 1. Pass 1: exact OS match and is_launcher == false
	for _, exe := range app.Executables {
		if strings.EqualFold(exe.OS, targetOS) && !exe.IsLauncher && exe.Name != "" {
			return exe.Name, gameTitle, nil
		}
	}

	// 2. Pass 2: exact OS match (even if launcher)
	for _, exe := range app.Executables {
		if strings.EqualFold(exe.OS, targetOS) && exe.Name != "" {
			return exe.Name, gameTitle, nil
		}
	}

	// 3. Pass 3: On Linux, fallback to Windows .exe for Wine/Proton game detection (non-launcher first)
	if currentOS == "linux" {
		for _, exe := range app.Executables {
			if strings.EqualFold(exe.OS, "win32") && !exe.IsLauncher && strings.HasSuffix(strings.ToLower(exe.Name), ".exe") {
				return exe.Name, gameTitle, nil
			}
		}
		for _, exe := range app.Executables {
			if strings.EqualFold(exe.OS, "win32") && strings.HasSuffix(strings.ToLower(exe.Name), ".exe") {
				return exe.Name, gameTitle, nil
			}
		}
	}

	// 4. Pass 4: Any non-empty executable
	if len(app.Executables) > 0 && app.Executables[0].Name != "" {
		return app.Executables[0].Name, gameTitle, nil
	}

	return "", gameTitle, fmt.Errorf("no executable registered for application %s (%s)", appID, gameTitle)
}
