package gate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile writes a file below a directory, creating the directories.
func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEnglishCheckPassesOnAnEnglishTree(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "README.md", "# Title\n\nSome English prose.\n")
	res := EnglishCheck(root).Run(context.Background())
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, detail = %s", res.Status, res.Detail)
	}
}

func TestEnglishCheckFailsOnChineseText(t *testing.T) {
	root := t.TempDir()
	// The fixture is written with unicode escapes on purpose: the test
	// source itself must stay ASCII, because the English gate scans the
	// repository's sources too.
	writeFile(t, root, "README.md", "# Title\n\n\u8fd9\u662f\u4e2d\u6587\u3002\n")
	res := EnglishCheck(root).Run(context.Background())
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}
	if !strings.Contains(res.Detail, "README.md") {
		t.Errorf("the failure does not name the file: %s", res.Detail)
	}
}

func TestLinksCheckPassesOnResolvableLinks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Docs/a.md", "[b](b.md)\n\n[ref][r]\n\n[r]: <b.md>\n")
	writeFile(t, root, "Docs/b.md", "# b\n")
	res := LinksCheck(root).Run(context.Background())
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, detail = %s", res.Status, res.Detail)
	}
}

func TestLinksCheckFailsOnABrokenLink(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Docs/a.md", "[missing](missing.md)\n")
	res := LinksCheck(root).Run(context.Background())
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}
	if !strings.Contains(res.Detail, "missing.md") {
		t.Errorf("the failure does not name the target: %s", res.Detail)
	}
}

func TestLinksCheckFailsOnADanglingReference(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Docs/a.md", "See [the other document][missing].\n")
	res := LinksCheck(root).Run(context.Background())
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}
	if !strings.Contains(res.Detail, "no definition") {
		t.Errorf("the failure does not explain the dangling reference: %s", res.Detail)
	}
}

func TestLinksCheckIgnoresExternalTargets(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Docs/a.md", "[site](https://example.com) and [mail](mailto:a@b.c)\n")
	res := LinksCheck(root).Run(context.Background())
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, detail = %s", res.Status, res.Detail)
	}
}

func TestIndexCheckFailsWhenADocumentIsNotListed(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Docs/Features/index.md", "# Index\n\n[FEAT-001](platform/one.md)\n")
	writeFile(t, root, "Docs/Features/platform/one.md", "# one\n")
	writeFile(t, root, "Docs/Features/platform/two.md", "# two\n")
	writeFile(t, root, "Docs/Fix/index.md", "# Index\n")
	res := IndexCheck(root).Run(context.Background())
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}
	if !strings.Contains(res.Detail, "platform/two.md") {
		t.Errorf("the failure does not name the unlisted document: %s", res.Detail)
	}
}

func TestIndexCheckPassesWhenEveryDocumentIsListed(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Docs/Features/index.md", "# Index\n\n[FEAT-001](platform/one.md)\n")
	writeFile(t, root, "Docs/Features/platform/one.md", "# one\n")
	writeFile(t, root, "Docs/Fix/index.md", "# Index\n")
	res := IndexCheck(root).Run(context.Background())
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, detail = %s", res.Status, res.Detail)
	}
}

func TestClassifyAndAction(t *testing.T) {
	cases := map[Status]Classification{
		StatusPassed:  ClassificationReady,
		StatusFailed:  ClassificationBlocked,
		StatusSkipped: ClassificationInsufficientEvidence,
	}
	for status, want := range cases {
		if got := Classify(status); got != want {
			t.Errorf("Classify(%s) = %s, want %s", status, got, want)
		}
	}
	if ActionFor(ClassificationBlocked) != ActionFix {
		t.Errorf("a blocked gate must recommend a fix")
	}
}
