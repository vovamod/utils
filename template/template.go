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
	GetBase(name string) (string, bool)
	GetSnippet(objectName, key string) (string, bool)
}

// Trequest creates a template request (base name, additional objects to include, data fields to replace)
type Trequest struct {
	BaseName string              `json:"base_name"`
	Objects  []string            `json:"objects"` // e.g., "button.primary"
	Data     []map[string]string `json:"data"`
}

type Store struct {
	mu          sync.RWMutex
	baseHTML    map[string]string
	htmlObjects map[string]map[string]string
}

func NewStore() *Store {
	return &Store{
		baseHTML:    make(map[string]string),
		htmlObjects: make(map[string]map[string]string),
	}
}

func (s *Store) GetBase(name string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.baseHTML[name]
	return val, ok
}

func (s *Store) GetSnippet(objName, key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if objMap, ok := s.htmlObjects[objName]; ok {
		var snippet string
		snippet, ok = objMap[key]
		return snippet, ok
	}
	return "", false
}

// LoadFromDir scans a directory for .html (bases) and .json (objects).
func (s *Store) LoadFromDir(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fullPath := filepath.Join(dirPath, file.Name())
		ext := strings.ToLower(filepath.Ext(file.Name()))
		name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		var data []byte
		data, err = os.ReadFile(fullPath)
		if err != nil {
			return err
		}

		s.mu.Lock()
		switch ext {
		case ".html":
			s.baseHTML[name] = string(data)
		case ".json":
			var obj map[string]string
			if err = json.Unmarshal(data, &obj); err == nil {
				s.htmlObjects[name] = obj
			}
		}
		s.mu.Unlock()
	}
	return nil
}

type Templator struct {
	provider Provider
}

func NewTemplator(p Provider) *Templator {
	r := &Templator{
		provider: p,
	}
	return r
}

func (r *Templator) Format(req Trequest) (string, error) {
	base, ok := r.provider.GetBase(req.BaseName)
	if !ok {
		return "", fmt.Errorf("base template '%s' not found", req.BaseName)
	}

	// If objects present we firstly need to modify them and paste em inside our final body
	if len(req.Objects) > 0 {
		var sb strings.Builder
		for _, objKey := range req.Objects {
			parts := strings.SplitN(objKey, ".", 2)
			if len(parts) == 2 {
				if s, ok := r.provider.GetSnippet(parts[0], parts[1]); ok {
					sb.WriteString(s)
				}
			}
		}
		base = strings.ReplaceAll(base, "{body}", sb.String())
	} else {
		base = strings.ReplaceAll(base, "{body}", "")
	}

	// Prepare for req.Data replacement
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
