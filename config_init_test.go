package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunConfigInit_NewDirectory(t *testing.T) {
	// Network-dependent — tested in BATS integration tests.
	t.Skip("Network-dependent test - tested in BATS integration tests")
}

func TestRunConfigInit_ExistingFileForce(t *testing.T) {
	// Network-dependent — tested in BATS integration tests.
	t.Skip("Network-dependent test - tested in BATS integration tests")
}

func TestInitArgs_Structure(t *testing.T) {
	args := InitArgs{
		Force: true,
		URL:   "https://example.com/test.yml",
	}

	if !args.Force {
		t.Error("Force field not working")
	}
	if args.URL != "https://example.com/test.yml" {
		t.Error("URL field not working")
	}
}

func TestDownloadFile_InvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yml")

	err := downloadFile("http://invalid-url-that-does-not-exist.local", testFile)
	if err == nil {
		t.Error("Expected error when downloading from invalid URL")
	}
	if fileExists(testFile) {
		t.Error("File should not be created when download fails")
	}
}

func TestDownloadFile_404Response(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yml")

	err := downloadFile("https://raw.githubusercontent.com/ahoy-cli/ahoy/master/non-existent-file.yml", testFile)
	if err == nil {
		t.Error("Expected error when downloading 404 URL")
	}
	if err != nil && err.Error() == "" {
		t.Error("Error should have a descriptive message")
	}
	if fileExists(testFile) {
		t.Error("File should not be created when download returns 404")
	}
}

func TestFileExists_Helper(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if fileExists(testFile) {
		t.Error("fileExists should return false for non-existent file")
	}

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !fileExists(testFile) {
		t.Error("fileExists should return true for existing file")
	}

	if fileExists(tmpDir) {
		t.Error("fileExists should return false for directories")
	}
}
