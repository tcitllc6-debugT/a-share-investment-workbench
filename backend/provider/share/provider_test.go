package share

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalProviderWritesMarkdown(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalProvider(dir)

	path, err := p.ShareText("hello", "AI助手")

	require.NoError(t, err)
	require.FileExists(t, path)
	require.Equal(t, filepath.Join(dir, "AI助手.md"), path)
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(b), "hello")
}
