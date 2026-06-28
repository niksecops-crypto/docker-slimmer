package optimizer

import (
	"bufio"
	"os"
	"strings"
)

// ParsedDockerfile holds structured information extracted from an existing Dockerfile.
type ParsedDockerfile struct {
	BaseImage      string
	IsMultiStage   bool
	HasDistroless  bool
	HasNobodyUser  bool
	PackageManager string
	Stages         []Stage
}

// Stage represents one FROM block in a Dockerfile.
type Stage struct {
	Name  string
	Image string
	Lines []string
}

// ParseFile reads a Dockerfile from disk and returns structured metadata.
func ParseFile(path string) (*ParsedDockerfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	res := parse(scanner)
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// ParseString parses Dockerfile content from a string (useful in tests).
func ParseString(content string) *ParsedDockerfile {
	return parse(bufio.NewScanner(strings.NewReader(content)))
}

func parse(scanner *bufio.Scanner) *ParsedDockerfile {
	result := &ParsedDockerfile{}
	var current *Stage

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		upper := strings.ToUpper(line)

		switch {
		case strings.HasPrefix(upper, "FROM"):
			parts := strings.Fields(line)
			if len(parts) == 0 || !strings.EqualFold(parts[0], "from") {
				continue
			}
			// Filter out --platform flag if present
			var filteredParts []string
			for _, p := range parts {
				if !strings.HasPrefix(strings.ToLower(p), "--platform=") {
					filteredParts = append(filteredParts, p)
				}
			}
			if len(filteredParts) < 2 {
				continue
			}
			image := filteredParts[1]
			name := ""
			if len(filteredParts) >= 4 && strings.EqualFold(filteredParts[2], "as") {
				name = filteredParts[3]
			}
			s := Stage{Image: image, Name: name}
			result.Stages = append(result.Stages, s)
			current = &result.Stages[len(result.Stages)-1]

			if result.BaseImage == "" {
				result.BaseImage = image
			}
			if strings.Contains(strings.ToLower(image), "distroless") {
				result.HasDistroless = true
			}

		case current != nil:
			current.Lines = append(current.Lines, line)

			if strings.Contains(upper, "APT-GET") {
				result.PackageManager = "apt"
			} else if strings.Contains(upper, "APK") {
				result.PackageManager = "apk"
			}

			if strings.HasPrefix(upper, "USER") {
				userParts := strings.Fields(upper)
				if len(userParts) >= 2 && userParts[0] == "USER" {
					u := userParts[1]
					if u == "NOBODY" || u == "65534" {
						result.HasNobodyUser = true
					}
				}
			}
		}
	}

	result.IsMultiStage = len(result.Stages) > 1
	return result
}

// AnalyzeIssues returns a list of human-readable improvement suggestions.
func AnalyzeIssues(df *ParsedDockerfile) []string {
	var issues []string

	if !df.IsMultiStage {
		issues = append(issues, "Single-stage build detected: consider multi-stage to separate build and runtime")
	}
	if !df.HasDistroless {
		issues = append(issues, "Base image is not distroless: runtime attack surface can be reduced significantly")
	}
	if !df.HasNobodyUser {
		issues = append(issues, "No explicit non-root user: add USER nobody or USER 65534")
	}
	if df.PackageManager == "apt" {
		issues = append(issues, "apt-get detected: ensure rm -rf /var/lib/apt/lists/* follows each install")
	}

	return issues
}
