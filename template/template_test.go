package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderer_Render(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "templates_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(path string) {
		err = os.RemoveAll(path)
		if err != nil {
			t.Errorf(err.Error())
		}
	}(tmpDir)

	baseHTML := `<html><body><h1>{title}</h1><div>{body}</div><p>Footer: {footer}</p></body></html>`
	if err := os.WriteFile(filepath.Join(tmpDir, "layout.html"), []byte(baseHTML), 0644); err != nil {
		t.Fatal(err)
	}

	snippetsJSON := `{
		"button": "<button>{btn_text}</button>",
		"input": "<input type='text' placeholder='{hint}'>"
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "elements.json"), []byte(snippetsJSON), 0644); err != nil {
		t.Fatal(err)
	}

	store := NewStore()
	if err := store.LoadFromDir(tmpDir); err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

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

		result, err := renderer.Format(req)
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
			Objects:  nil, // No objects provided
			Data: []map[string]string{
				{"{title}": "Empty Page"},
				{"{footer}": "No Body Here"},
			},
		}

		result, err := renderer.Format(req)
		if err != nil {
			t.Fatal(err)
		}

		if contains(result, "<button>") {
			t.Error("Result should not contain snippets")
		}
	})

	t.Run("Missing Base Error", func(t *testing.T) {
		req := Trequest{BaseName: "non_existent"}
		_, err = renderer.Format(req)
		if err == nil {
			t.Error("Expected error for missing base template, got nil")
		}
	})
}

func contains(str, substr string) bool {
	result, _ := filepath.Match("*"+substr+"*", str) // Simple check (actually I could make simpler, but nope.)
	return result
}
