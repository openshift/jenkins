package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"
)

const (
	pluginVersionsAPIURL     = "https://updates.jenkins.io/plugin-versions.json"
	jenkinsLTSChecksumURLFmt = "https://get.jenkins.io/war-stable/%s/jenkins.war.sha256"
	jenkinsLTSDownloadURLFmt = "https://get.jenkins.io/war-stable/%s/jenkins.war"
	outputFile               = "artifacts.lock.yaml"
	maxRetries               = 3
	artifactsLockTemplate    = `---
metadata:
  version: "1.0"
artifacts:
{{- range .}}
  - download_url: {{.DownloadURL | printf "%q"}}
    checksum: {{.Checksum | printf "%q"}}
    filename: {{.Filename | printf "%q"}}
{{- end}}
`
)

// Artifact represents a single artifact in the output
type Artifact struct {
	DownloadURL string
	Checksum    string
	Filename    string
}

// Plugin represents the internal plugin data structure
type Plugin struct {
	Name        string
	Version     string
	SHA256      string
	DownloadURL string
	Error       string
}

// PluginRequest represents a plugin with its requested version
type PluginRequest struct {
	Name    string
	Version string
}

// PluginVersions represents the plugin-versions.json structure
type PluginVersions struct {
	Plugins map[string]map[string]VersionData `json:"plugins"`
}

// VersionData represents data for a specific version of a plugin
type VersionData struct {
	SHA256 string `json:"sha256"`
	URL    string `json:"url"`
}

// CoreData represents Jenkins core information
type CoreData struct {
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
	URL     string `json:"url"`
}

// base64ToHex converts base64-encoded checksum to hex
func base64ToHex(b64 string) string {
	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return b64 // Return original if decode fails
	}
	return fmt.Sprintf("%x", decoded)
}

// fetchPluginVersionsWithRetry fetches the plugin-versions.json API with retry logic
func fetchPluginVersionsWithRetry() (*PluginVersions, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(delay)
		}

		resp, err := http.Get(pluginVersionsAPIURL)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %v", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %v", err)
			continue
		}

		var pv PluginVersions
		if err := json.Unmarshal(body, &pv); err != nil {
			lastErr = fmt.Errorf("failed to parse JSON: %v", err)
			continue
		}

		return &pv, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}

// fetchJenkinsLTSChecksumWithRetry fetches Jenkins LTS checksum from remote .sha256 file with retry logic
func fetchJenkinsLTSChecksumWithRetry(version string) (*CoreData, error) {
	checksumURL := fmt.Sprintf(jenkinsLTSChecksumURLFmt, version)
	downloadURL := fmt.Sprintf(jenkinsLTSDownloadURLFmt, version)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
		}

		resp, err := http.Get(checksumURL)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %v", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d: jenkins LTS version %s not found", resp.StatusCode, version)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %v", err)
			continue
		}

		// Format: "checksum jenkins.war"
		checksumLine := strings.TrimSpace(string(body))
		parts := strings.Fields(checksumLine)
		if len(parts) < 1 {
			lastErr = fmt.Errorf("invalid checksum file format")
			continue
		}

		sha256Hex := parts[0]

		return &CoreData{
			Version: version,
			SHA256:  sha256Hex,
			URL:     downloadURL,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}

// extractPlugin extracts plugin information from plugin-versions API
func extractPlugin(req PluginRequest, pv *PluginVersions) Plugin {
	// Get version-specific data from plugin-versions API
	pluginVersions, exists := pv.Plugins[req.Name]
	if !exists {
		return Plugin{
			Name:    req.Name,
			Version: req.Version,
			Error:   "plugin not found",
		}
	}

	versionData, exists := pluginVersions[req.Version]
	if !exists {
		return Plugin{
			Name:    req.Name,
			Version: req.Version,
			Error:   fmt.Sprintf("version %s not found", req.Version),
		}
	}

	// Convert base64 SHA256 to hex
	sha256Hex := base64ToHex(versionData.SHA256)

	return Plugin{
		Name:        req.Name,
		Version:     req.Version,
		SHA256:      sha256Hex,
		DownloadURL: versionData.URL,
	}
}

// processPlugins processes all plugins concurrently
func processPlugins(requests []PluginRequest, pv *PluginVersions) []Plugin {
	results := make(chan Plugin, len(requests))
	var wg sync.WaitGroup

	// Process each plugin in a goroutine
	for _, req := range requests {
		wg.Add(1)
		go func(r PluginRequest) {
			defer wg.Done()
			results <- extractPlugin(r, pv)
		}(req)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	plugins := make([]Plugin, 0, len(requests))
	for plugin := range results {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// readPluginFile reads plugin names and versions from a file (format: name:version)
func readPluginFile(filename string) ([]PluginRequest, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var requests []PluginRequest
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Extract plugin name and version
		parts := strings.SplitN(line, ":", 2)
		name := strings.TrimSpace(parts[0])
		version := ""
		if len(parts) > 1 {
			version = strings.TrimSpace(parts[1])
		}
		requests = append(requests, PluginRequest{
			Name:    name,
			Version: version,
		})
	}

	return requests, scanner.Err()
}

func main() {
	var pluginRequests []PluginRequest
	var pluginsFile string
	var jenkinsFile string
	outputFilePath := outputFile // Default: artifacts.lock.yaml

	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: jenkins-plugin-checker --plugins <file-path> --jenkins <file-path> [--output <file-path>]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Required:")
		fmt.Fprintln(os.Stderr, "  --plugins <file>  Path to plugins file (format: plugin-name:version)")
		fmt.Fprintln(os.Stderr, "  --jenkins <file>  Path to file containing Jenkins LTS version (e.g., 2.504.2)")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Optional:")
		fmt.Fprintln(os.Stderr, "  --output <file>   Path to output file (default: artifacts.lock.yaml)")
		os.Exit(1)
	}

	// Parse flags
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--plugins":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "Error: --plugins requires a file path")
				os.Exit(1)
			}
			pluginsFile = os.Args[i+1]
			i++ // Skip next arg
		case "--jenkins":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "Error: --jenkins requires a file path")
				os.Exit(1)
			}
			jenkinsFile = os.Args[i+1]
			i++ // Skip next arg
		case "--output":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "Error: --output requires a file path")
				os.Exit(1)
			}
			outputFilePath = os.Args[i+1]
			i++ // Skip next arg
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", os.Args[i])
			os.Exit(1)
		}
	}

	// Validate required arguments
	if pluginsFile == "" || jenkinsFile == "" {
		fmt.Fprintln(os.Stderr, "Error: both --plugins and --jenkins flags are required")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Usage: jenkins-plugin-checker --plugins <file-path> --jenkins <file-path> [--output <file-path>]")
		os.Exit(1)
	}

	// Read plugins from file
	requests, err := readPluginFile(pluginsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading plugins file: %v\n", err)
		os.Exit(1)
	}
	pluginRequests = requests

	if len(pluginRequests) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no plugins to check")
		os.Exit(1)
	}

	// Read Jenkins version from file
	jenkinsVersionBytes, err := os.ReadFile(jenkinsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading Jenkins version file: %v\n", err)
		os.Exit(1)
	}
	jenkinsVersion := strings.TrimSpace(string(jenkinsVersionBytes))
	if jenkinsVersion == "" {
		fmt.Fprintln(os.Stderr, "Error: Jenkins version file is empty")
		os.Exit(1)
	}

	// Fetch plugin versions API with retry
	pv, err := fetchPluginVersionsWithRetry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching plugin versions API: %v\n", err)
		os.Exit(1)
	}

	// Fetch Jenkins checksum and process plugins concurrently
	var jenkinsCore *CoreData
	var jenkinsFetchErr error
	jenkinsDone := make(chan bool)

	// Fetch Jenkins checksum in background
	go func() {
		jenkinsCore, jenkinsFetchErr = fetchJenkinsLTSChecksumWithRetry(jenkinsVersion)
		jenkinsDone <- true
	}()

	// Process plugins concurrently
	results := processPlugins(pluginRequests, pv)

	// Wait for Jenkins fetch to complete
	<-jenkinsDone
	if jenkinsFetchErr != nil {
		fmt.Fprintf(os.Stderr, "Error fetching Jenkins %s: %v\n", jenkinsVersion, jenkinsFetchErr)
		os.Exit(1)
	}

	// Collect errors
	errorsByType := make(map[string][]string)
	var allArtifacts []Artifact

	// Add Jenkins core first
	allArtifacts = append(allArtifacts, Artifact{
		DownloadURL: jenkinsCore.URL,
		Checksum:    "sha256:" + jenkinsCore.SHA256,
		Filename:    "jenkins.war",
	})

	// Add plugin artifacts
	for _, plugin := range results {
		if plugin.Error != "" {
			errorsByType[plugin.Error] = append(errorsByType[plugin.Error], plugin.Name)
			continue
		}

		filename := filepath.Base(plugin.DownloadURL)
		if filename == "" || filename == "." {
			filename = plugin.Name + ".hpi"
		}

		// Build artifact
		allArtifacts = append(allArtifacts, Artifact{
			DownloadURL: plugin.DownloadURL,
			Checksum:    "sha256:" + plugin.SHA256,
			Filename:    filename,
		})
	}

	// If there are any errors, don't output YAML, exit with error
	if len(errorsByType) > 0 {
		fmt.Fprintln(os.Stderr, "Error: Failed to fetch information for the following plugins:")
		for errType, pluginNames := range errorsByType {
			fmt.Fprintf(os.Stderr, "\n%s:\n", errType)
			for _, name := range pluginNames {
				fmt.Fprintf(os.Stderr, "  - %s\n", name)
			}
		}
		os.Exit(1)
	}

	// Sort artifacts: jenkins.war first, then plugins alphabetically
	sort.Slice(allArtifacts, func(i, j int) bool {
		// jenkins.war always comes first
		if allArtifacts[i].Filename == "jenkins.war" {
			return true
		}
		if allArtifacts[j].Filename == "jenkins.war" {
			return false
		}
		// Sort other artifacts alphabetically
		return allArtifacts[i].Filename < allArtifacts[j].Filename
	})

	// Parse template
	tmpl, err := template.New("output").Parse(artifactsLockTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	// Create output file
	file, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file %s: %v\n", outputFilePath, err)
		os.Exit(1)
	}
	defer file.Close()

	// Execute template and write to file
	err = tmpl.Execute(file, allArtifacts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	// Calculate statistics
	totalInputPlugins := len(pluginRequests)
	totalOutputPlugins := len(allArtifacts) - 1 // Exclude jenkins.war
	jenkinsWarCount := 1

	// Print summary
	fmt.Fprintf(os.Stderr, "Successfully generated %s\n\n", outputFilePath)
	fmt.Fprintln(os.Stderr, "Summary:")
	fmt.Fprintf(os.Stderr, "  Input plugins:    %d\n", totalInputPlugins)
	fmt.Fprintf(os.Stderr, "  Output plugins:   %d\n", totalOutputPlugins)
	fmt.Fprintf(os.Stderr, "  Jenkins WAR:      %d\n", jenkinsWarCount)
	fmt.Fprintf(os.Stderr, "  Total artifacts:  %d\n", len(allArtifacts))
}
