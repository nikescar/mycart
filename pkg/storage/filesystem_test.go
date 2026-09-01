package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilesystem_New(t *testing.T) {
	t.Parallel()

	basePath := t.TempDir()

	fs, err := NewFilesystem(basePath)
	require.NoError(t, err)
	require.NotNil(t, fs)
}

func TestFilesystem_Put(t *testing.T) {
	t.Parallel()

	basePath := t.TempDir()
	fs, err := NewFilesystem(basePath)
	require.NoError(t, err)

	ctx := context.Background()
	data := bytes.NewBufferString("test content")

	err = fs.Put(ctx, "test.txt", data)
	require.NoError(t, err)
}

func TestFilesystem_Get(t *testing.T) {
	t.Parallel()

	basePath := t.TempDir()
	fs, err := NewFilesystem(basePath)
	require.NoError(t, err)

	ctx := context.Background()

	// Put file first
	testContent := "test content"
	err = fs.Put(ctx, "test.txt", strings.NewReader(testContent))
	require.NoError(t, err)

	// Get file
	reader, err := fs.Get(ctx, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, reader)
	defer reader.Close()

	// Verify content
	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, testContent, string(content))
}

func TestFilesystem_Get_NotFound(t *testing.T) {
	t.Parallel()

	basePath := t.TempDir()
	fs, err := NewFilesystem(basePath)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = fs.Get(ctx, "nonexistent.txt")
	require.Error(t, err)
}

func TestFilesystem_Delete(t *testing.T) {
	t.Parallel()

	basePath := t.TempDir()
	fs, err := NewFilesystem(basePath)
	require.NoError(t, err)

	ctx := context.Background()

	// Put file first
	err = fs.Put(ctx, "test.txt", strings.NewReader("test"))
	require.NoError(t, err)

	// Delete file
	err = fs.Delete(ctx, "test.txt")
	require.NoError(t, err)

	// Verify it's gone
	exists, err := fs.Exists(ctx, "test.txt")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestFilesystem_Exists(t *testing.T) {
	t.Parallel()

	basePath := t.TempDir()
	fs, err := NewFilesystem(basePath)
	require.NoError(t, err)

	ctx := context.Background()

	// File doesn't exist yet
	exists, err := fs.Exists(ctx, "test.txt")
	require.NoError(t, err)
	require.False(t, exists)

	// Put file
	err = fs.Put(ctx, "test.txt", strings.NewReader("test"))
	require.NoError(t, err)

	// File now exists
	exists, err = fs.Exists(ctx, "test.txt")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestFilesystem_List(t *testing.T) {
	t.Parallel()

	basePath := t.TempDir()
	fs, err := NewFilesystem(basePath)
	require.NoError(t, err)

	ctx := context.Background()

	// Put multiple files
	files := []string{
		"uploads/image1.jpg",
		"uploads/image2.jpg",
		"digitals/file1.pdf",
		"digitals/file2.pdf",
	}

	for _, file := range files {
		err = fs.Put(ctx, file, strings.NewReader("content"))
		require.NoError(t, err)
	}

	// List uploads
	uploadFiles, err := fs.List(ctx, "uploads/")
	require.NoError(t, err)
	require.Len(t, uploadFiles, 2)
	require.Contains(t, uploadFiles, "uploads/image1.jpg")
	require.Contains(t, uploadFiles, "uploads/image2.jpg")

	// List digitals
	digitalFiles, err := fs.List(ctx, "digitals/")
	require.NoError(t, err)
	require.Len(t, digitalFiles, 2)
	require.Contains(t, digitalFiles, "digitals/file1.pdf")
	require.Contains(t, digitalFiles, "digitals/file2.pdf")

	// List all
	allFiles, err := fs.List(ctx, "")
	require.NoError(t, err)
	require.Len(t, allFiles, 4)
}
