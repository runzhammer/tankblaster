package tankblaster

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/runzhammer/tankblaster/pkg/buildinfo"
)

const (
	versionCheckURL     = "https://raw.githubusercontent.com/runzhammer/tankblaster/main/VERSION"
	versionDownloadURL  = "https://www.tankblaster.de"
	versionCheckTimeout = 3 * time.Second
)

type versionUpdateResult struct {
	available bool
	current   string
	latest    string
	err       error
}

type semanticVersion struct {
	major      int
	minor      int
	patch      int
	prerelease []string
}

func checkLatestVersion(ctx context.Context, url, current string) versionUpdateResult {
	currentVersion, ok := parseSemanticVersion(current)
	if !ok {
		return versionUpdateResult{current: current}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return versionUpdateResult{current: current, err: err}
	}
	client := &http.Client{Timeout: versionCheckTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return versionUpdateResult{current: current, err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return versionUpdateResult{current: current, err: fmt.Errorf("version check status %d", resp.StatusCode)}
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 128))
	if err != nil {
		return versionUpdateResult{current: current, err: err}
	}
	latest := strings.TrimSpace(string(data))
	latestVersion, ok := parseSemanticVersion(latest)
	if !ok {
		return versionUpdateResult{current: current, latest: latest}
	}
	return versionUpdateResult{
		available: compareSemanticVersions(latestVersion, currentVersion) > 0,
		current:   normalizeVersionString(currentVersion),
		latest:    normalizeVersionString(latestVersion),
	}
}

func currentSemanticVersionString() string {
	version := strings.TrimSpace(buildinfo.Version)
	if version == "" || strings.EqualFold(version, "dev") {
		if data, err := os.ReadFile("VERSION"); err == nil {
			version = strings.TrimSpace(string(data))
		}
	}
	return version
}

func parseSemanticVersion(value string) (semanticVersion, bool) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "v")
	if value == "" {
		return semanticVersion{}, false
	}
	if buildSplit := strings.SplitN(value, "+", 2); len(buildSplit) == 2 {
		value = buildSplit[0]
	}
	var prerelease []string
	if preSplit := strings.SplitN(value, "-", 2); len(preSplit) == 2 {
		value = preSplit[0]
		if preSplit[1] == "" {
			return semanticVersion{}, false
		}
		prerelease = strings.Split(preSplit[1], ".")
		for _, part := range prerelease {
			if part == "" || !semanticIdentifierValid(part) {
				return semanticVersion{}, false
			}
		}
	}
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return semanticVersion{}, false
	}
	nums := make([]int, 3)
	for i, part := range parts {
		if part == "" || !semanticNumberValid(part) {
			return semanticVersion{}, false
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return semanticVersion{}, false
		}
		nums[i] = n
	}
	return semanticVersion{major: nums[0], minor: nums[1], patch: nums[2], prerelease: prerelease}, true
}

func semanticNumberValid(value string) bool {
	if len(value) > 1 && value[0] == '0' {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func semanticIdentifierValid(value string) bool {
	for _, r := range value {
		if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '-' {
			continue
		}
		return false
	}
	return true
}

func compareSemanticVersions(left, right semanticVersion) int {
	if left.major != right.major {
		return compareInt(left.major, right.major)
	}
	if left.minor != right.minor {
		return compareInt(left.minor, right.minor)
	}
	if left.patch != right.patch {
		return compareInt(left.patch, right.patch)
	}
	return comparePrerelease(left.prerelease, right.prerelease)
}

func compareInt(left, right int) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func comparePrerelease(left, right []string) int {
	if len(left) == 0 && len(right) == 0 {
		return 0
	}
	if len(left) == 0 {
		return 1
	}
	if len(right) == 0 {
		return -1
	}
	count := minInt(len(left), len(right))
	for i := 0; i < count; i++ {
		leftNum, leftErr := strconv.Atoi(left[i])
		rightNum, rightErr := strconv.Atoi(right[i])
		switch {
		case leftErr == nil && rightErr == nil:
			if leftNum != rightNum {
				return compareInt(leftNum, rightNum)
			}
		case leftErr == nil:
			return -1
		case rightErr == nil:
			return 1
		case left[i] < right[i]:
			return -1
		case left[i] > right[i]:
			return 1
		}
	}
	return compareInt(len(left), len(right))
}

func normalizeVersionString(version semanticVersion) string {
	value := fmt.Sprintf("%d.%d.%d", version.major, version.minor, version.patch)
	if len(version.prerelease) > 0 {
		value += "-" + strings.Join(version.prerelease, ".")
	}
	return value
}
