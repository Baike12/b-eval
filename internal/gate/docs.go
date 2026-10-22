package gate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// DocsCheck verifies a property of the documentation tree.
//
// The documentation checks are part of the gate rather than a separate
// script, because a gate that a developer does not run is not a gate.
type DocsCheck struct {
	// GateID is the stable identifier of the gate.
	GateID string
	// Description states what the gate protects.
	Description string
	// Root is the repository root.
	Root string
	// Check is the check body.
	Check func(ctx context.Context) Result
}

// ID returns the gate identifier.
func (c DocsCheck) ID() string { return c.GateID }

// Describe returns what the gate protects.
func (c DocsCheck) Describe() string { return c.Description }

// Run executes the check body.
func (c DocsCheck) Run(ctx context.Context) Result {
	return c.Check(ctx)
}

// cjkPattern matches CJK ideographs, CJK punctuation and full-width forms.
//
// The rule is that the repository is written in English only, and the check
// enforces it on every file, including the ones a reader never sees.
var cjkPattern = regexp.MustCompile(`[\x{4e00}-\x{9fff}\x{3400}-\x{4dbf}\x{3000}-\x{303f}\x{ff00}-\x{ffef}]`)

// excludedDirs are the directories the documentation gates skip.
//
// vendor/ holds third-party mirrors in their original language, and the
// frontend dependencies are not the platform's own text.
var excludedDirs = map[string]bool{
	"vendor":       true,
	"node_modules": true,
	".git":         true,
	"dist":         true,
	".venv":        true,
}

// translationDirs are the directories that hold the interface's
// translation catalogues.
//
// The catalogue of a non-English locale is the one text in the repository
// that is not written in English: it is the product's own second language,
// and a reader has to be able to read it. The keys stay English, so the
// catalogue is a table of values and not a body of prose.
var translationDirs = []string{
	filepath.Join("web", "src", "i18n"),
}

// isTranslationDir reports whether a directory holds translation
// catalogues.
func isTranslationDir(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	for _, dir := range translationDirs {
		if rel == dir {
			return true
		}
	}
	return false
}

// EnglishCheck verifies that no CJK text entered the repository.
func EnglishCheck(root string) Check {
	return DocsCheck{
		GateID:      "docs-english",
		Description: "Every file in the repository is written in English.",
		Root:        root,
		Check: func(ctx context.Context) Result {
			res := Result{GateID: "docs-english", Status: StatusPassed}
			var hits []string
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					if excludedDirs[info.Name()] {
						return filepath.SkipDir
					}
					if isTranslationDir(root, path) {
						return filepath.SkipDir
					}
					return nil
				}
				switch filepath.Ext(path) {
				case ".md", ".go", ".ts", ".tsx", ".json", ".yaml", ".yml", ".sh", ".html", ".css":
				default:
					return nil
				}
				data, readErr := os.ReadFile(path)
				if readErr != nil {
					return readErr
				}
				for i, line := range strings.Split(string(data), "\n") {
					if cjkPattern.MatchString(line) {
						rel, _ := filepath.Rel(root, path)
						hits = append(hits, fmt.Sprintf("%s:%d", rel, i+1))
					}
				}
				return nil
			})
			if err != nil {
				res.Status = StatusFailed
				res.Detail = err.Error()
				return res
			}
			if len(hits) > 0 {
				res.Status = StatusFailed
				res.Detail = fmt.Sprintf("CJK text found in %d place(s): %s", len(hits), strings.Join(hits[:min(len(hits), 20)], ", "))
			}
			res.Evidence = map[string]any{"violations": len(hits)}
			return res
		},
	}
}

// linkDefinitionPattern matches a reference-style link definition.
var linkDefinitionPattern = regexp.MustCompile(`(?m)^\[([^\]\s]+)\]:\s*(<[^>]+>|\S+)\s*$`)

// referenceUsagePattern matches a reference-style link usage.
var referenceUsagePattern = regexp.MustCompile(`\[([^\]\[]*)\]\[([^\]\[]*)\]`)

// inlineLinkPattern matches an inline link.
var inlineLinkPattern = regexp.MustCompile(`\[([^\]\[]*)\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)

// fencePattern matches a fenced code block.
var fencePattern = regexp.MustCompile("(?s)```.*?```")

// LinksCheck verifies that every local link in the documentation resolves.
//
// A broken link is a defect: it is a promise the documentation makes to the
// reader that the repository does not keep.
func LinksCheck(root string) Check {
	return DocsCheck{
		GateID:      "docs-links",
		Description: "Every local link in the documentation resolves to an existing file.",
		Root:        root,
		Check: func(ctx context.Context) Result {
			res := Result{GateID: "docs-links", Status: StatusPassed}
			files, err := markdownFiles(root)
			if err != nil {
				res.Status = StatusFailed
				res.Detail = err.Error()
				return res
			}
			var broken []string
			for _, file := range files {
				data, readErr := os.ReadFile(file)
				if readErr != nil {
					res.Status = StatusFailed
					res.Detail = readErr.Error()
					return res
				}
				text := stripCodeBlocks(string(data))
				broken = append(broken, checkFileLinks(root, file, text)...)
			}
			if len(broken) > 0 {
				res.Status = StatusFailed
				res.Detail = fmt.Sprintf("%d broken link(s):\n%s", len(broken), strings.Join(broken, "\n"))
			}
			res.Evidence = map[string]any{"broken": len(broken), "files": len(files)}
			return res
		},
	}
}

// checkFileLinks returns the broken links of one file.
func checkFileLinks(root, file, text string) []string {
	var broken []string
	rel, _ := filepath.Rel(root, file)
	dir := filepath.Dir(file)

	definitions := map[string]string{}
	for _, m := range linkDefinitionPattern.FindAllStringSubmatch(text, -1) {
		definitions[m[1]] = strings.Trim(m[2], "<>")
	}

	// A reference-style usage whose id has no definition is a dangling
	// reference: the reader sees the link text and nothing behind it.
	withoutDefinitions := linkDefinitionPattern.ReplaceAllString(text, "")
	for _, m := range referenceUsagePattern.FindAllStringSubmatch(withoutDefinitions, -1) {
		id := m[2]
		if id == "" {
			id = m[1]
		}
		if id == "" {
			continue
		}
		if _, ok := definitions[id]; !ok {
			broken = append(broken, fmt.Sprintf("%s: reference [%s] has no definition", rel, id))
		}
	}

	for _, m := range inlineLinkPattern.FindAllStringSubmatch(text, -1) {
		target := m[2]
		if isExternal(target) {
			continue
		}
		if !resolves(dir, target) {
			broken = append(broken, fmt.Sprintf("%s: link target %q does not exist", rel, target))
		}
	}

	for id, target := range definitions {
		if isExternal(target) {
			continue
		}
		if !resolves(dir, target) {
			broken = append(broken, fmt.Sprintf("%s: link definition [%s] points at %q, which does not exist", rel, id, target))
		}
	}
	return broken
}

// isExternal reports whether a link target leaves the repository.
func isExternal(target string) bool {
	switch {
	case strings.HasPrefix(target, "http://"),
		strings.HasPrefix(target, "https://"),
		strings.HasPrefix(target, "mailto:"),
		strings.HasPrefix(target, "#"),
		strings.HasPrefix(target, "tel:"):
		return true
	}
	return false
}

// resolves reports whether a relative link target exists on disk.
func resolves(dir, target string) bool {
	target = strings.Trim(target, "<>")
	if idx := strings.Index(target, "#"); idx >= 0 {
		target = target[:idx]
	}
	if target == "" {
		return true
	}
	path := filepath.Join(dir, filepath.FromSlash(target))
	_, err := os.Stat(path)
	return err == nil
}

// stripCodeBlocks removes fenced code blocks from a document.
//
// A link inside a code block is an example, not a link, and checking it
// would report a defect that is not one.
func stripCodeBlocks(text string) string {
	return fencePattern.ReplaceAllString(text, "")
}

// IndexCheck verifies that every document appears in its folder's index.
//
// The index is the entry point of its folder. A document missing from it is
// a document a reader cannot find.
func IndexCheck(root string) Check {
	return DocsCheck{
		GateID:      "docs-index",
		Description: "Every feature and fix document appears in its folder's index.",
		Root:        root,
		Check: func(ctx context.Context) Result {
			res := Result{GateID: "docs-index", Status: StatusPassed}
			var missing []string
			for _, folder := range []string{"Docs/Features", "Docs/Fix"} {
				dir := filepath.Join(root, folder)
				indexPath := filepath.Join(dir, "index.md")
				indexData, err := os.ReadFile(indexPath)
				if err != nil {
					res.Status = StatusFailed
					res.Detail = fmt.Sprintf("index %s is missing: %v", folder, err)
					return res
				}
				indexText := string(indexData)
				err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() || info.Name() == "index.md" {
						return nil
					}
					if filepath.Ext(path) != ".md" {
						return nil
					}
					rel, _ := filepath.Rel(dir, path)
					if !strings.Contains(indexText, filepath.ToSlash(rel)) {
						missing = append(missing, fmt.Sprintf("%s is not listed in %s/index.md", filepath.ToSlash(rel), folder))
					}
					return nil
				})
				if err != nil {
					res.Status = StatusFailed
					res.Detail = err.Error()
					return res
				}
			}
			if len(missing) > 0 {
				res.Status = StatusFailed
				sort.Strings(missing)
				res.Detail = fmt.Sprintf("%d document(s) missing from an index:\n%s", len(missing), strings.Join(missing, "\n"))
			}
			res.Evidence = map[string]any{"missing": len(missing)}
			return res
		},
	}
}

// markdownFiles lists the markdown files of the repository, skipping the
// directories that are not the project's own text.
func markdownFiles(root string) ([]string, error) {
	var out []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if excludedDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".md" {
			out = append(out, path)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
