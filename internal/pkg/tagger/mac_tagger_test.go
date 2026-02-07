//go:build darwin
// +build darwin

package tagger

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMacOSTagger_AddTag(t *testing.T) {
	// Create a test file on disk.
	testDir := "testdata"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	testFile := filepath.Join(testDir, "test_image.jpg")
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer func() {
		file.Close()
		os.RemoveAll(testDir)
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
