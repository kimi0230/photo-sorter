//go:build darwin
// +build darwin

package tagger

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMacOSTagger_AddTag(t *testing.T) {
	if _, err := exec.LookPath("tag"); err != nil {
		t.Skipf("missing tag command: %v", err)
	}

	// Create a test file on disk.
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test_image.jpg")
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer func() {
		file.Close()
	}()

	// Write some test data.
	_, err = file.Write([]byte("test image content"))
	if err != nil {
		t.Fatalf("failed to write test data: %v", err)
	}

	tagger := NewMacOSTagger()

	// Add a tag.
	err = tagger.AddTag(testFile, "TestTag")
	if err != nil {
		t.Errorf("failed to add tag: %v", err)
	}

	// List tags.
	tags, err := tagger.ListTags(testFile)
	if err != nil {
		t.Errorf("failed to list tags: %v", err)
	}

	fmt.Printf("tags for %s: %v\n", testFile, tags)
}
