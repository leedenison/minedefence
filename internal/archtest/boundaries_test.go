package archtest

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The rules below are the executable form of the dependency table in
// docs/architecture.md. Change both together, and only with an ADR.

// allowOnly lists, for headless packages, every non-standard-library package
// allowed anywhere in their transitive dependencies. Entries ending in "/..."
// match the path and everything under it. Paths are relative to the module
// unless they contain a dot in the first element.
var allowOnly = map[string][]string{
	"internal/sim/...": {"internal/sim/...", "internal/balance", "assets", "github.com/BurntSushi/toml/..."},
	"internal/balance": {"internal/balance", "assets", "github.com/BurntSushi/toml/..."},
	"assets":           {"assets"},
}

// forbid lists packages that must not appear in the transitive dependencies
// of presentation and network packages.
var forbid = map[string][]string{
	"internal/netplay/...": {"github.com/hajimehoshi/ebiten/...", "internal/render/...", "internal/ui/...", "internal/app/..."},
	"internal/render/...":  {"internal/netplay/...", "internal/ui/...", "internal/app/..."},
	"internal/ui/...":      {"internal/netplay/...", "internal/app/..."},
}

// simForbiddenStd lists standard library packages the simulation must not
// import directly because they break determinism or headlessness. Use
// math/rand/v2 with an explicit seeded source held in the World instead of
// the global functions.
var simForbiddenStd = []string{
	"time", "math/rand", "crypto/rand", "os", "os/exec", "net", "net/http",
	"sync", "sync/atomic", "runtime", "unsafe", "syscall",
}

func TestHeadlessPackagesOnlyUseAllowedDeps(t *testing.T) {
	mod := modulePath(t)
	for pkg, allowed := range allowOnly {
		for _, dep := range nonStdDeps(t, mod, pkg) {
			if !matchesAny(mod, dep, allowed) {
				t.Errorf("%s depends on %s, which is not in its allow list", pkg, dep)
			}
		}
	}
}

func TestForbiddenDeps(t *testing.T) {
	mod := modulePath(t)
	for pkg, banned := range forbid {
		for _, dep := range nonStdDeps(t, mod, pkg) {
			if matchesAny(mod, dep, banned) {
				t.Errorf("%s must not depend on %s", pkg, dep)
			}
		}
	}
}

func TestSimDirectStdImports(t *testing.T) {
	mod := modulePath(t)
	out := goList(t, "-f", `{{.ImportPath}}{{range .Imports}} {{.}}{{end}}`, mod+"/internal/sim/...")
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		for _, imp := range fields[1:] {
			for _, bad := range simForbiddenStd {
				if imp == bad {
					t.Errorf("%s imports %s; the simulation must stay deterministic and headless", fields[0], imp)
				}
			}
		}
	}
}

// modulePath returns the module path. It also reads go.mod, go.sum and every
// .go file in the module so that the go test cache, which only tracks files
// the test process itself opens, is invalidated when any import changes.
// Without this, results computed by the go list subprocess would be cached.
func modulePath(t *testing.T) string {
	t.Helper()
	fields := strings.Split(strings.TrimSpace(goList(t, "-m", "-f", "{{.Path}}\n{{.Dir}}")), "\n")
	if len(fields) != 2 {
		t.Fatalf("unexpected go list -m output: %q", fields)
	}
	root := fields[1]
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") || d.Name() == "go.mod" || d.Name() == "go.sum" {
			_, err = os.ReadFile(path)
		}
		return err
	})
	if err != nil {
		t.Fatalf("reading module sources: %v", err)
	}
	return fields[0]
}

// nonStdDeps returns the non-standard-library transitive dependencies of the
// packages matching pattern, including those packages themselves.
func nonStdDeps(t *testing.T, mod, pattern string) []string {
	t.Helper()
	out := goList(t, "-deps", "-f", `{{if not .Standard}}{{.ImportPath}}{{end}}`, mod+"/"+pattern)
	return strings.Fields(out)
}

func matchesAny(mod, dep string, patterns []string) bool {
	for _, p := range patterns {
		if !strings.Contains(strings.SplitN(p, "/", 2)[0], ".") {
			p = mod + "/" + p
		}
		if prefix, ok := strings.CutSuffix(p, "/..."); ok {
			if dep == prefix || strings.HasPrefix(dep, prefix+"/") {
				return true
			}
		} else if dep == p {
			return true
		}
	}
	return false
}

func goList(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("go", append([]string{"list"}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		t.Fatalf("go list %v: %v\n%s", args, err, stderr)
	}
	return string(out)
}
