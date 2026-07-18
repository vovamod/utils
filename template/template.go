package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Provider interface {
	GetBase(lang, name string) (string, bool)
	GetSnippet(lang, objName, key string) (string, bool)
}

// Trequest creates a template request (base name, additional objects to include, data fields to replace)
type Trequest struct {
	BaseName string              `json:"base_name"`
	Objects  []string            `json:"objects"` // e.g., "button.primary"
	Data     []map[string]string `json:"data"`
}

type Store struct {
	mu          sync.RWMutex
	defaultLang string
	baseHTML    map[string]map[string]string            // lang/name/html
	htmlObjects map[string]map[string]map[string]string // lang/objName/key.value
}

func NewStore(defaultLang string) *Store {
	return &Store{
		defaultLang: defaultLang,
		baseHTML:    make(map[string]map[string]string),
		htmlObjects: make(map[string]map[string]map[string]string),
	}
}

func (s *Store) GetBase(lang, name string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if v, ok := s.lookupBase(lang, name); ok {
		return v, true
	}
	if lang != s.defaultLang {
		return s.lookupBase(s.defaultLang, name)
	}
	return "", false
}

func (s *Store) lookupBase(lang, name string) (string, bool) {
	m, ok := s.baseHTML[lang]
	if !ok {
		return "", false
	}
	v, ok := m[name]
	return v, ok
}

func (s *Store) GetSnippet(lang, objName, key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if v, ok := s.lookupSnippet(lang, objName, key); ok {
		return v, true
	}
	if lang != s.defaultLang {
		return s.lookupSnippet(s.defaultLang, objName, key)
	}
	return "", false
}

func (s *Store) lookupSnippet(lang, objName, key string) (string, bool) {
	objs, ok := s.htmlObjects[lang]
	if !ok {
		return "", false
	}
	obj, ok := objs[objName]
	if !ok {
		return "", false
	}
	v, ok := obj[key]
	return v, ok
}

// LoadFromDir scans rootPath for one subdirectory per language (e.g. "en/",
// "uk/"), and within each, .html files as bases and .json files as
// snippet/translation objects. Rework for rc-0.8.0
func (s *Store) LoadFromDir(rootPath string) error {
	langDirs, err := os.ReadDir(rootPath)
	if err != nil {
		return fmt.Errorf("failed to read template root directory: %w", err)
	}

	loadedAny := false

	for _, langEntry := range langDirs {
		if !langEntry.IsDir() {
			continue
		}
		lang := langEntry.Name()
		langPath := filepath.Join(rootPath, lang)

		files, err := os.ReadDir(langPath)
		if err != nil {
			return fmt.Errorf("failed to read language directory %q: %w", lang, err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			fullPath := filepath.Join(langPath, file.Name())
			ext := strings.ToLower(filepath.Ext(file.Name()))
			name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

			data, err := os.ReadFile(fullPath)
			if err != nil {
				return fmt.Errorf("failed to read %q: %w", fullPath, err)
			}

			switch ext {
			case ".html":
				s.mu.Lock()
				if s.baseHTML[lang] == nil {
					s.baseHTML[lang] = make(map[string]string)
				}
				s.baseHTML[lang][name] = string(data)
				s.mu.Unlock()
				loadedAny = true

			case ".json":
				var obj map[string]string
				if err = json.Unmarshal(data, &obj); err != nil {
					return fmt.Errorf("failed to parse %q: %w", fullPath, err)
				}
				s.mu.Lock()
				if s.htmlObjects[lang] == nil {
					s.htmlObjects[lang] = make(map[string]map[string]string)
				}
				s.htmlObjects[lang][name] = obj
				s.mu.Unlock()
				loadedAny = true
			}
		}
	}

	if !loadedAny {
		return fmt.Errorf("no templates loaded from %q -- expected language subdirectories (e.g. %q)", rootPath, s.defaultLang)
	}

	s.mu.RLock()
	_, hasDefaultBase := s.baseHTML[s.defaultLang]
	_, hasDefaultObjs := s.htmlObjects[s.defaultLang]
	s.mu.RUnlock()
	if !hasDefaultBase && !hasDefaultObjs {
		return fmt.Errorf("default language %q has no loaded templates -- fallback lookups would always fail", s.defaultLang)
	}

	return nil
}

type Templator struct {
	provider Provider
}

func NewTemplator(p Provider) *Templator {
	return &Templator{provider: p}
}

func (r *Templator) Format(lang string, req Trequest) (string, error) {
	base, ok := r.provider.GetBase(lang, req.BaseName)
	if !ok {
		return "", fmt.Errorf("base template %q not found for lang %q", req.BaseName, lang)
	}

	if len(req.Objects) > 0 {
		var sb strings.Builder
		for _, objKey := range req.Objects {
			parts := strings.SplitN(objKey, ".", 2)
			if len(parts) == 2 {
				if s, ok := r.provider.GetSnippet(lang, parts[0], parts[1]); ok {
					sb.WriteString(s)
				}
			}
		}
		base = strings.ReplaceAll(base, "{body}", sb.String())
	} else {
		base = strings.ReplaceAll(base, "{body}", "")
	}

	dataEntryCount := 0
	for _, m := range req.Data {
		dataEntryCount += len(m)
	}
	if dataEntryCount == 0 {
		return base, nil
	}

	pairs := make([]string, 0, dataEntryCount*2)
	for _, dataMap := range req.Data {
		for k, v := range dataMap {
			pairs = append(pairs, k, v)
		}
	}

	return strings.NewReplacer(pairs...).Replace(base), nil
}
