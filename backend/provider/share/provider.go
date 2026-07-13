package share

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Provider interface {
	ShareText(text string, title string) (string, error)
}

type LocalProvider struct {
	dir string
}

func NewLocalProvider(dir string) *LocalProvider {
	return &LocalProvider{dir: dir}
}

func (p *LocalProvider) ShareText(text string, title string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("share text is empty")
	}
	title = sanitizeTitle(title)
	if title == "" {
		title = "analysis"
	}
	if err := os.MkdirAll(p.dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(p.dir, title+".md")
	return path, os.WriteFile(path, []byte(text), 0644)
}

func sanitizeTitle(title string) string {
	title = strings.TrimSpace(title)
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	title = re.ReplaceAllString(title, "_")
	return strings.Trim(title, ". ")
}
