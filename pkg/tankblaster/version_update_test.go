package tankblaster

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseSemanticVersionRequiresMajorMinorPatch(t *testing.T) {
	valid := []string{"1.2.3", "v1.2.3", "1.2.3-alpha.1", "1.2.3+build.7"}
	for _, value := range valid {
		if _, ok := parseSemanticVersion(value); !ok {
			t.Fatalf("parseSemanticVersion(%q) failed", value)
		}
	}
	invalid := []string{"dev", "1", "1.2", "1.2.x", "1.02.3", "1.2.3-"}
	for _, value := range invalid {
		if _, ok := parseSemanticVersion(value); ok {
			t.Fatalf("parseSemanticVersion(%q) succeeded, want invalid", value)
		}
	}
}

func TestCompareSemanticVersions(t *testing.T) {
	tests := []struct {
		left  string
		right string
		want  int
	}{
		{"1.0.1", "1.0.0", 1},
		{"1.1.0", "1.0.9", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.0.0", "1.0.0", 0},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0-beta", "1.0.0-alpha", 1},
	}
	for _, tt := range tests {
		left, ok := parseSemanticVersion(tt.left)
		if !ok {
			t.Fatalf("parse left %q", tt.left)
		}
		right, ok := parseSemanticVersion(tt.right)
		if !ok {
			t.Fatalf("parse right %q", tt.right)
		}
		if got := compareSemanticVersions(left, right); normalizeCompare(got) != tt.want {
			t.Fatalf("compareSemanticVersions(%q, %q) = %d, want %d", tt.left, tt.right, got, tt.want)
		}
	}
}

func TestCheckLatestVersionOnlyReportsNewerSemanticVersion(t *testing.T) {
	version := "1.2.4\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(version))
	}))
	defer server.Close()

	result := checkLatestVersion(context.Background(), server.URL, "1.2.3")
	if !result.available || result.latest != "1.2.4" {
		t.Fatalf("newer result = %+v, want update to 1.2.4", result)
	}

	result = checkLatestVersion(context.Background(), server.URL, "1.2.4")
	if result.available {
		t.Fatalf("same version result = %+v, want no update", result)
	}

	version = "dev\n"
	result = checkLatestVersion(context.Background(), server.URL, "1.2.3")
	if result.available {
		t.Fatalf("non-semver remote result = %+v, want no update", result)
	}
}

func normalizeCompare(value int) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}
