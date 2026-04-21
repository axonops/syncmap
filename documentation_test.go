// Copyright 2026 AxonOps Limited.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syncmap_test

import (
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLLMs_TxtExists_AndUnderTokenBudget asserts the llms.txt file
// exists at the repo root and stays within the token budget that
// fits comfortably into an AI assistant's context window.
func TestLLMs_TxtExists_AndUnderTokenBudget(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("llms.txt")
	require.NoError(t, err, "llms.txt must exist at the repo root")

	words := len(strings.Fields(string(data)))
	assert.LessOrEqual(t, words, 2250,
		"llms.txt must stay under the 2250-word budget (got %d)", words)
	assert.Greater(t, words, 200,
		"llms.txt looks stubby (got %d words)", words)
}

// TestLLMs_FullTxtExists_AndIncludesSpecifiedSections asserts
// llms-full.txt is present and concatenates every canonical source
// listed in issue #17.
func TestLLMs_FullTxtExists_AndIncludesSpecifiedSections(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("llms-full.txt")
	require.NoError(t, err, "llms-full.txt must exist at the repo root")

	body := string(data)
	required := []string{
		"# syncmap — full documentation bundle",
		"# llms.txt",
		"# README.md",
		"# Package godoc (doc.go)",
		"# CONTRIBUTING.md",
		"# SECURITY.md",
		"# CHANGELOG.md",
		"# Full godoc reference (go doc -all)",
	}
	for _, header := range required {
		assert.Contains(t, body, header,
			"llms-full.txt must contain section header %q", header)
	}
}

// TestLLMs_FullTxtIsUpToDate re-runs the generator and asserts
// byte-equality with the committed file. If this fails, someone edited
// a source file and forgot to run `make llms-full`.
//
// This test intentionally does NOT use t.Parallel(): it overwrites the
// repo-root `llms-full.txt` while running, which would race with any
// other parallel test that reads that file (or its adjacent sources).
func TestLLMs_FullTxtIsUpToDate(t *testing.T) {
	committed, err := os.ReadFile("llms-full.txt")
	require.NoError(t, err)

	backup := t.TempDir() + "/llms-full.txt.committed"
	require.NoError(t, os.WriteFile(backup, committed, 0o644))
	t.Cleanup(func() {
		_ = os.WriteFile("llms-full.txt", committed, 0o644)
	})

	cmd := exec.Command("./scripts/gen-llms-full.sh")
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run(), "gen-llms-full.sh must exit 0")

	regenerated, err := os.ReadFile("llms-full.txt")
	require.NoError(t, err)
	assert.Equal(t, string(committed), string(regenerated),
		"llms-full.txt drift — run 'make llms-full' and commit the result")
}

// requiredExamples is the set pinned by issue #15 — every public
// symbol exposes a runnable godoc Example.
var requiredExamples = []string{
	"ExampleSyncMap",
	"ExampleSyncMap_Load",
	"ExampleSyncMap_Store",
	"ExampleSyncMap_LoadOrStore",
	"ExampleSyncMap_LoadAndDelete",
	"ExampleSyncMap_Delete",
	"ExampleSyncMap_Swap",
	"ExampleSyncMap_Clear",
	"ExampleSyncMap_Range",
	"ExampleSyncMap_Len",
	"ExampleSyncMap_Map",
	"ExampleSyncMap_Keys",
	"ExampleSyncMap_Values",
	"ExampleCompareAndSwap",
	"ExampleCompareAndDelete",
}

// TestExamples_AllRequiredExamplesExist parses example_test.go and
// asserts that every required godoc Example function is defined.
// Missing examples fail the build — this is the primary AI-assistant
// integration surface on pkg.go.dev.
func TestExamples_AllRequiredExamplesExist(t *testing.T) {
	t.Parallel()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "example_test.go", nil, 0)
	require.NoError(t, err)

	defined := map[string]struct{}{}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if strings.HasPrefix(fn.Name.Name, "Example") {
			defined[fn.Name.Name] = struct{}{}
		}
	}
	for _, required := range requiredExamples {
		_, ok := defined[required]
		assert.Truef(t, ok,
			"required example %q is missing from example_test.go", required)
	}
}

// mechanicalDoc matches trivially-generated one-liners like
// "Foo returns a string." — we want real prose that tells a reader
// how and when to use the symbol, not a restatement of the signature.
// The trailing period is required: a first line without one is a
// wrapped multi-line paragraph, which is always fine.
var mechanicalDoc = regexp.MustCompile(`^\w+ (returns|is|creates) [\w ]+\.$`)

// TestDocumentation_EveryExportedSymbolHasGodoc parses the package
// and asserts every exported symbol has a doc comment of at least
// 20 characters that is not a mechanical one-liner.
func TestDocumentation_EveryExportedSymbolHasGodoc(t *testing.T) {
	t.Parallel()
	fset := token.NewFileSet()

	entries, err := os.ReadDir(".")
	require.NoError(t, err)
	files := map[string]*ast.File{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
		require.NoError(t, err, "parse %s", name)
		if f.Name.Name != "syncmap" {
			continue
		}
		files[name] = f
	}
	require.NotEmpty(t, files, "no syncmap package files found")

	docPkg, err := doc.NewFromFiles(fset, sortedFiles(files), "github.com/axonops/syncmap")
	require.NoError(t, err)

	checkDoc := func(name, text string) {
		t.Run(name, func(t *testing.T) {
			trimmed := strings.TrimSpace(text)
			assert.GreaterOrEqual(t, len(trimmed), 20,
				"symbol %q has a doc comment shorter than 20 characters: %q", name, text)
			// Only flag as mechanical when the ENTIRE doc is a single
			// line matching the pattern. Multi-line docs with the same
			// crisp opening sentence are canonical Go godoc style.
			if !strings.Contains(trimmed, "\n") {
				assert.False(t, mechanicalDoc.MatchString(trimmed),
					"symbol %q has a mechanical one-line doc: %q", name, trimmed)
			}
		})
	}
	for _, c := range docPkg.Consts {
		for _, n := range c.Names {
			checkDoc(n, c.Doc)
		}
	}
	for _, v := range docPkg.Vars {
		for _, n := range v.Names {
			checkDoc(n, v.Doc)
		}
	}
	for _, f := range docPkg.Funcs {
		checkDoc(f.Name, f.Doc)
	}
	for _, typ := range docPkg.Types {
		checkDoc(typ.Name, typ.Doc)
		for _, m := range typ.Methods {
			checkDoc(typ.Name+"."+m.Name, m.Doc)
		}
		for _, f := range typ.Funcs {
			checkDoc(f.Name, f.Doc)
		}
	}
}

// TestReadmeQuickStart_Compiles extracts the README Quick Start code
// block, compiles it in a fresh temporary module, runs it, and
// verifies it produces the documented output. This catches drift
// between the README's copy-paste snippet and the library API.
func TestReadmeQuickStart_Compiles(t *testing.T) {
	if testing.Short() {
		t.Skip("-short set; skipping compilation test")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("go toolchain not on PATH: %v", err)
	}
	t.Parallel()

	readme, err := os.ReadFile("README.md")
	require.NoError(t, err)

	snippet, ok := extractQuickStartBlock(string(readme))
	require.True(t, ok, "could not find the Quick Start go code block in README.md")

	repoDir, err := os.Getwd()
	require.NoError(t, err)
	repoDir, err = filepath.Abs(repoDir)
	require.NoError(t, err)

	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "main.go"), []byte(snippet), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "go.mod"),
		[]byte(fmt.Sprintf("module quickstart\n\ngo 1.26\n\nrequire github.com/axonops/syncmap v0.0.0\n\nreplace github.com/axonops/syncmap => %s\n", repoDir)),
		0o644))

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmp
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())

	build := exec.Command("go", "build", "-o", "main", ".")
	build.Dir = tmp
	build.Stderr = os.Stderr
	require.NoError(t, build.Run(), "Quick Start snippet must compile")

	run := exec.Command("./main")
	run.Dir = tmp
	out, err := run.Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "hits: 1",
		"Quick Start output must include the documented first line")
	assert.Contains(t, string(out), "incremented",
		"Quick Start output must confirm the CompareAndSwap succeeded")
}

// TestGovernance_NoticeFileExists asserts NOTICE is present at the
// repo root and carries the AxonOps and upstream Richard Gooding
// attribution required for the fork.
func TestGovernance_NoticeFileExists(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("NOTICE")
	require.NoError(t, err, "NOTICE must exist at the repo root (Apache 2.0 § 4(d))")

	s := string(body)
	assert.Contains(t, s, "AxonOps Limited",
		"NOTICE must name AxonOps Limited")
	assert.Contains(t, s, "Apache License",
		"NOTICE must reference the Apache License")
	assert.Contains(t, s, "rgooding/go-syncmap",
		"NOTICE must credit the upstream repository")
	assert.Contains(t, s, "Richard Gooding",
		"NOTICE must credit the upstream author by name")
}

// TestGovernance_SecurityPolicyExists asserts SECURITY.md is present
// and carries the private-reporting contact.
func TestGovernance_SecurityPolicyExists(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("SECURITY.md")
	require.NoError(t, err, "SECURITY.md must exist at the repo root")

	s := string(body)
	assert.Contains(t, s, "oss@axonops.com",
		"SECURITY.md must carry the AxonOps oss@axonops.com reporting contact")
	assert.Contains(t, s, "Supported versions",
		"SECURITY.md must document supported versions")
}

// TestGovernance_ContributingExists asserts CONTRIBUTING.md is present
// and carries the load-bearing policy sections.
func TestGovernance_ContributingExists(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("CONTRIBUTING.md")
	require.NoError(t, err, "CONTRIBUTING.md must exist at the repo root")

	s := string(body)
	for _, section := range []string{
		"Contributor License Agreement",
		"Code of Conduct",
		"Attribution policy",
		"Branching and commits",
		"Test requirements",
		"Releases",
	} {
		assert.Contains(t, s, section, "CONTRIBUTING.md must contain %q", section)
	}
}

// TestGovernance_CLADocumentExists asserts CLA.md is present and
// carries the legally load-bearing sections.
func TestGovernance_CLADocumentExists(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("CLA.md")
	require.NoError(t, err, "CLA.md must exist at the repo root")

	s := string(body)
	for _, section := range []string{
		"Grant of copyright licence",
		"Grant of patent licence",
		"Representations",
		"AxonOps",
	} {
		assert.Contains(t, s, section, "CLA.md must contain %q", section)
	}
}

// TestGovernance_CodeOfConductExists asserts CODE_OF_CONDUCT.md is
// present, is derived from the Contributor Covenant, and carries the
// AxonOps enforcement contact.
func TestGovernance_CodeOfConductExists(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("CODE_OF_CONDUCT.md")
	require.NoError(t, err, "CODE_OF_CONDUCT.md must exist at the repo root")

	s := string(body)
	assert.Contains(t, s, "Contributor Covenant",
		"CODE_OF_CONDUCT.md must be derived from the Contributor Covenant")
	assert.Contains(t, s, "oss@axonops.com",
		"CODE_OF_CONDUCT.md must carry the AxonOps enforcement contact")
	assert.NotContains(t, s, "[INSERT CONTACT METHOD]",
		"CODE_OF_CONDUCT.md must have the contact placeholder filled in")
}

// TestGovernance_CLAWorkflowExists asserts the CLA Assistant workflow
// is present and points at the syncmap repository.
func TestGovernance_CLAWorkflowExists(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile(".github/workflows/cla.yml")
	require.NoError(t, err, ".github/workflows/cla.yml must exist")

	s := string(body)
	assert.Contains(t, s, "axonops/syncmap",
		"cla.yml must reference this repository (not a mask copy-paste)")
	assert.Contains(t, s, "CLA_ASSISTANT_PAT",
		"cla.yml must wire the CLA_ASSISTANT_PAT secret for branch-protection bypass")
	assert.NotContains(t, s, "axonops/mask",
		"cla.yml must not reference axonops/mask (adapt repo-specific values)")
}

// TestGovernance_ContributorsWorkflowExists asserts the contributors
// regeneration workflow is present.
func TestGovernance_ContributorsWorkflowExists(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile(".github/workflows/contributors.yml")
	require.NoError(t, err, ".github/workflows/contributors.yml must exist")

	s := string(body)
	assert.Contains(t, s, "signatures/version1/cla.json",
		"contributors.yml must trigger on the CLA signatures file")
	assert.Contains(t, s, "scripts/generate-contributors.sh",
		"contributors.yml must invoke the generator script")
}

// TestGovernance_ContributorsFileIsGenerated asserts CONTRIBUTORS.md
// matches the output of the generator script (i.e. nobody has
// hand-edited it).
func TestGovernance_ContributorsFileIsGenerated(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skipf("jq not on PATH: %v", err)
	}
	t.Parallel()

	committed, err := os.ReadFile("CONTRIBUTORS.md")
	require.NoError(t, err, "CONTRIBUTORS.md must exist")

	tmpOut := filepath.Join(t.TempDir(), "CONTRIBUTORS.md")
	cmd := exec.Command("./scripts/generate-contributors.sh",
		"signatures/version1/cla.json", tmpOut)
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run(), "generate-contributors.sh must exit 0")

	regenerated, err := os.ReadFile(tmpOut)
	require.NoError(t, err)
	assert.Equal(t, string(committed), string(regenerated),
		"CONTRIBUTORS.md drift — run 'scripts/generate-contributors.sh' and commit the result")
}

// TestGovernance_SignaturesFileIsValid asserts the signatures file
// exists, is valid JSON, and carries the expected schema skeleton.
func TestGovernance_SignaturesFileIsValid(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("signatures/version1/cla.json")
	require.NoError(t, err, "signatures/version1/cla.json must exist")

	s := string(body)
	assert.Contains(t, s, "signedContributors",
		"signatures/version1/cla.json must carry the signedContributors key")
}

// TestGovernance_ChangelogHasV1Entry asserts CHANGELOG.md is present
// and carries the v1.0.0 entry that documents the fork changes.
func TestGovernance_ChangelogHasV1Entry(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("CHANGELOG.md")
	require.NoError(t, err, "CHANGELOG.md must exist at the repo root")

	s := string(body)
	assert.Contains(t, s, "## [1.0.0]",
		"CHANGELOG must contain a ## [1.0.0] section")
	assert.Contains(t, s, "Keep a Changelog",
		"CHANGELOG must reference the Keep a Changelog format")
	assert.Contains(t, s, "rgooding/go-syncmap",
		"CHANGELOG must credit the upstream fork origin")
	assert.Contains(t, s, "Items",
		"CHANGELOG must record the Items → Values breaking rename from #12")
	assert.Contains(t, s, "Values",
		"CHANGELOG must record the Items → Values breaking rename from #12")
}

// quickStartHeadingPattern locates any H2 heading containing
// "Quick Start" (case-insensitive), tolerating emoji decoration.
var quickStartHeadingPattern = regexp.MustCompile(`(?mi)^##\s+.*Quick\s+Start`)

// extractQuickStartBlock finds the first ```go ... ``` fence
// following the Quick Start heading.
func extractQuickStartBlock(body string) (string, bool) {
	loc := quickStartHeadingPattern.FindStringIndex(body)
	if loc == nil {
		return "", false
	}
	after := body[loc[1]:]
	start := strings.Index(after, "```go")
	if start < 0 {
		return "", false
	}
	start += len("```go")
	if start < len(after) && after[start] == '\n' {
		start++
	}
	end := strings.Index(after[start:], "```")
	if end < 0 {
		return "", false
	}
	return after[start : start+end], true
}

// sortedFiles returns the map's values ordered by file name so
// doc.NewFromFiles sees a deterministic input slice.
func sortedFiles(m map[string]*ast.File) []*ast.File {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	out := make([]*ast.File, 0, len(m))
	for _, n := range names {
		out = append(out, m[n])
	}
	return out
}

// ensure errors import stays live when a future edit trims an assertion;
// keeping it avoids a goimports flip-flop.
var _ = errors.Is
