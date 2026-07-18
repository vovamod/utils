package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func newTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "templates_test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	enBase := `<html><body><h1>{title}</h1><div>{body}</div><p>Footer: {footer}</p></body></html>`
	writeFile(t, filepath.Join(tmpDir, "en", "layout.html"), enBase)

	enSnippets := `{
		"button": "<button>{btn_text}</button>",
		"input": "<input type='text' placeholder='{hint}'>"
	}`
	writeFile(t, filepath.Join(tmpDir, "en", "elements.json"), enSnippets)

	ukSnippets := `{
		"button": "<button>{btn_text} (uk)</button>"
	}`
	writeFile(t, filepath.Join(tmpDir, "uk", "elements.json"), ukSnippets)

	store := NewStore("en")
	if err = store.LoadFromDir(tmpDir); err != nil {
		t.Fatalf("failed to load templates: %v", err)
	}
	return store, tmpDir
}

func TestRenderer_Render(t *testing.T) {
	store, _ := newTestStore(t)
	renderer := NewTemplator(store)

	t.Run("Full Format", func(t *testing.T) {
		req := Trequest{
			BaseName: "layout",
			Objects:  []string{"elements.button", "elements.input"},
			Data: []map[string]string{
				{"{title}": "Billing Portal"},
				{"{btn_text}": "Submit"},
				{"{hint}": "Enter ID"},
				{"{footer}": "2026 Corp"},
			},
		}

		result, err := renderer.Format("en", req)
		if err != nil {
			t.Fatalf("Format failed: %v", err)
		}

		expectedFragments := []string{
			"<h1>Billing Portal</h1>",
			"<button>Submit</button>",
			"placeholder='Enter ID'",
			"Footer: 2026 Corp",
		}
		for _, frag := range expectedFragments {
			if !strings.Contains(result, frag) {
				t.Errorf("Result missing fragment [%s]. Got: %s", frag, result)
			}
		}

		placeholders := []string{"{title}", "{body}", "{btn_text}", "{hint}"}
		for _, ph := range placeholders {
			if strings.Contains(result, ph) {
				t.Errorf("Placeholder [%s] was not replaced in output", ph)
			}
		}
	})

	t.Run("Redacted Body (No Objects)", func(t *testing.T) {
		req := Trequest{
			BaseName: "layout",
			Objects:  nil,
			Data: []map[string]string{
				{"{title}": "Empty Page"},
				{"{footer}": "No Body Here"},
			},
		}

		result, err := renderer.Format("en", req)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(result, "<button>") {
			t.Error("Result should not contain snippets")
		}
	})

	t.Run("Missing Base Error", func(t *testing.T) {
		req := Trequest{BaseName: "non_existent"}
		_, err := renderer.Format("en", req)
		if err == nil {
			t.Error("Expected error for missing base template, got nil")
		}
	})
}

func TestRenderer_LanguageFallback(t *testing.T) {
	store, _ := newTestStore(t)
	renderer := NewTemplator(store)

	t.Run("Base falls back when language has no override", func(t *testing.T) {
		req := Trequest{
			BaseName: "layout",
			Data:     []map[string]string{{"{title}": "UK Page", "{footer}": "x"}},
		}

		result, err := renderer.Format("uk", req)
		if err != nil {
			t.Fatalf("expected fallback to en base, got error: %v", err)
		}
		if !strings.Contains(result, "<h1>UK Page</h1>") {
			t.Errorf("expected fallback-rendered base, got: %s", result)
		}
	})

	t.Run("Snippet uses language-specific version when present", func(t *testing.T) {
		req := Trequest{
			BaseName: "layout",
			Objects:  []string{"elements.button"},
			Data:     []map[string]string{{"{title}": "t", "{footer}": "f", "{btn_text}": "Go"}},
		}

		result, err := renderer.Format("uk", req)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(result, "<button>Go (uk)</button>") {
			t.Errorf("expected uk-specific snippet, got: %s", result)
		}
	})

	t.Run("Snippet falls back to default language when missing entirely", func(t *testing.T) {
		req := Trequest{
			BaseName: "layout",
			Objects:  []string{"elements.input"},
			Data:     []map[string]string{{"{title}": "t", "{footer}": "f", "{hint}": "search"}},
		}

		result, err := renderer.Format("uk", req)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(result, "placeholder='search'") {
			t.Errorf("expected fallback to en snippet, got: %s", result)
		}
	})

	t.Run("Unknown language falls back entirely to default", func(t *testing.T) {
		req := Trequest{
			BaseName: "layout",
			Objects:  []string{"elements.button"},
			Data:     []map[string]string{{"{title}": "t", "{footer}": "f", "{btn_text}": "Hi"}},
		}

		result, err := renderer.Format("fr", req)
		if err != nil {
			t.Fatalf("expected full fallback to en, got error: %v", err)
		}
		if !strings.Contains(result, "<button>Hi</button>") {
			t.Errorf("expected en snippet (no '(uk)' suffix), got: %s", result)
		}
	})
}

func TestStore_GetBase_DirectLookup(t *testing.T) {
	store, _ := newTestStore(t)

	if _, ok := store.GetBase("en", "layout"); !ok {
		t.Error("expected en/layout.html to be found directly")
	}
	if _, ok := store.GetBase("en", "does_not_exist"); ok {
		t.Error("expected missing base name to return false even with valid language")
	}
}

func TestStore_GetSnippet_DirectLookup(t *testing.T) {
	store, _ := newTestStore(t)

	if v, ok := store.GetSnippet("uk", "elements", "button"); !ok || !strings.Contains(v, "(uk)") {
		t.Errorf("expected uk-specific button snippet, got %q (ok=%v)", v, ok)
	}
	if _, ok := store.GetSnippet("uk", "elements", "nonexistent_key"); ok {
		t.Error("expected missing snippet key to return false, not fall through silently")
	}
}

func TestStore_LoadFromDir_Errors(t *testing.T) {
	t.Run("no language subdirectories at all", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "templates_test_empty")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		writeFile(t, filepath.Join(tmpDir, "layout.html"), "<html></html>")

		store := NewStore("en")
		if err = store.LoadFromDir(tmpDir); err == nil {
			t.Error("expected error when no language subdirectories are present")
		}
	})

	t.Run("default language present but empty of templates", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "templates_test_missing_default")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		writeFile(t, filepath.Join(tmpDir, "uk", "layout.html"), "<html></html>")

		store := NewStore("en")
		if err = store.LoadFromDir(tmpDir); err == nil {
			t.Error("expected error when default language has no loaded templates")
		}
	})

	t.Run("malformed JSON snippet file fails loudly", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "templates_test_bad_json")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		writeFile(t, filepath.Join(tmpDir, "en", "layout.html"), "<html></html>")
		writeFile(t, filepath.Join(tmpDir, "en", "broken.json"), `{"button": "<button>` /* truncated, invalid JSON */)

		store := NewStore("en")
		if err = store.LoadFromDir(tmpDir); err == nil {
			t.Error("expected error on malformed snippet JSON, not silent skip")
		}
	})
}
