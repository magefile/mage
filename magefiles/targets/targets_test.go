package targets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// expectedReleaseFiles lists the release artifacts goreleaser should produce.
// Each entry uses %s as a placeholder for the version string.
var expectedReleaseFiles = []string{
	"mage_%s_checksums.txt",
	"mage_%s_DragonFlyBSD-64bit.tar.gz",
	"mage_%s_FreeBSD-64bit.tar.gz",
	"mage_%s_FreeBSD-ARM.tar.gz",
	"mage_%s_FreeBSD-ARM64.tar.gz",
	"mage_%s_Linux-64bit.tar.gz",
	"mage_%s_Linux-ARM.tar.gz",
	"mage_%s_Linux-ARM64.tar.gz",
	"mage_%s_macOS-64bit.tar.gz",
	"mage_%s_macOS-ARM64.tar.gz",
	"mage_%s_NetBSD-64bit.tar.gz",
	"mage_%s_NetBSD-ARM.tar.gz",
	"mage_%s_NetBSD-ARM64.tar.gz",
	"mage_%s_OpenBSD-64bit.tar.gz",
	"mage_%s_OpenBSD-ARM64.tar.gz",
	"mage_%s_Windows-64bit.zip",
	"mage_%s_Windows-ARM64.zip",
}

func TestRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping release test in short mode")
	}

	// goreleaser must run from the repo root where .goreleaser.yml lives.
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err := os.RemoveAll("../../dist")
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	version := "v1.0.99"

	dryRun := true
	if err := Release(version, &dryRun); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir("dist")
	if err != nil {
		t.Fatal(err)
	}

	releaserVer := strings.TrimPrefix(version, "v")

	// Build expected set, initially marking each as not found.
	expected := make(map[string]bool, len(expectedReleaseFiles))
	for _, pattern := range expectedReleaseFiles {
		expected[fmt.Sprintf(pattern, releaserVer)] = false
	}

	// Walk dist/ and match release artifacts.
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		isArtifact := strings.HasSuffix(name, ".tar.gz") ||
			strings.HasSuffix(name, ".zip") ||
			strings.HasSuffix(name, "_checksums.txt")
		if !isArtifact {
			continue
		}
		if _, ok := expected[name]; ok {
			expected[name] = true
		} else {
			t.Errorf("unexpected release artifact: %s", name)
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected release artifact not found: %s", name)
		}
	}
}
