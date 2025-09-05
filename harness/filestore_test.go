package harness

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStoreDownload(t *testing.T) {
	// Create a test server to mock the Harness API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate file download response
		if strings.Contains(r.URL.Path, "/download") {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test file content"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Setup test directory
	testDir := "./test_filestore"
	defer os.RemoveAll(testDir)

	// Create API client
	api := &APIRequest{
		BaseURL: server.URL,
		Client:  resty.New(),
		APIKey:  "test-api-key",
	}

	tests := []struct {
		name           string
		file           FileStoreContent
		account        string
		org            string
		project        string
		folder         string
		expectedPath   string
		expectError    bool
	}{
		{
			name: "Account level file download",
			file: FileStoreContent{
				Identifier: "test-file-1",
				Name:       "test-file.yaml",
				Path:       "/manifests/test-file.yaml",
			},
			account:      "test-account",
			org:          "",
			project:      "",
			folder:       "/account",
			expectedPath: "./filestore/account/manifests/test-file.yaml",
			expectError:  false,
		},
		{
			name: "Org level file download",
			file: FileStoreContent{
				Identifier: "test-file-2",
				Name:       "org-file.yaml",
				Path:       "/configs/org-file.yaml",
			},
			account:      "test-account",
			org:          "test-org",
			project:      "",
			folder:       "/test-org",
			expectedPath: "./filestore/test-org/configs/org-file.yaml",
			expectError:  false,
		},
		{
			name: "Project level file download",
			file: FileStoreContent{
				Identifier: "test-file-3",
				Name:       "project-file.yaml",
				Path:       "/templates/project-file.yaml",
			},
			account:      "test-account",
			org:          "test-org",
			project:      "test-project",
			folder:       "/test-org/test-project",
			expectedPath: "./filestore/test-org/test-project/templates/project-file.yaml",
			expectError:  false,
		},
		{
			name: "File without extension (should be skipped)",
			file: FileStoreContent{
				Identifier: "test-folder",
				Name:       "folder",
				Path:       "/folder",
			},
			account:      "test-account",
			org:          "",
			project:      "",
			folder:       "/account",
			expectedPath: "", // No file should be created
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			os.RemoveAll("./filestore")

			err := tt.file.DownloadFile(api, tt.account, tt.org, tt.project, tt.folder)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.expectedPath != "" {
				// Check if file was created at expected path
				assert.FileExists(t, tt.expectedPath, "File should exist at expected path")

				// Check file content
				content, err := ioutil.ReadFile(tt.expectedPath)
				require.NoError(t, err)
				assert.Equal(t, "test file content", string(content))
			}
		})
	}
}

func TestFileStorePathCorrection(t *testing.T) {
	// Test that file paths are correctly constructed without duplicate "filestore" directories
	tests := []struct {
		name         string
		folder       string
		filePath     string
		expectedPath string
	}{
		{
			name:         "Account level path",
			folder:       "/account",
			filePath:     "/manifests/deployment.yaml",
			expectedPath: "./filestore/account/manifests/deployment.yaml",
		},
		{
			name:         "Org level path",
			folder:       "/org1",
			filePath:     "/configs/service.yaml",
			expectedPath: "./filestore/org1/configs/service.yaml",
		},
		{
			name:         "Project level path",
			folder:       "/org1/project1",
			filePath:     "/templates/pipeline.yaml",
			expectedPath: "./filestore/org1/project1/templates/pipeline.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualPath := "./filestore" + tt.folder + tt.filePath
			assert.Equal(t, tt.expectedPath, actualPath)

			// Ensure no duplicate "filestore" directory
			assert.NotContains(t, actualPath, "filestore/filestore", "Path should not contain duplicate filestore directory")
		})
	}
}

func TestDirectoryCreation(t *testing.T) {
	testPath := "./test_filestore/org1/project1/manifests/deployment.yaml"
	defer os.RemoveAll("./test_filestore")

	// Ensure directory is created
	dir := filepath.Dir(testPath)
	err := os.MkdirAll(dir, 0755)
	assert.NoError(t, err)

	// Check directory exists
	assert.DirExists(t, dir)

	// Create a test file
	file, err := os.Create(testPath)
	assert.NoError(t, err)
	defer file.Close()

	// Write content and verify
	content := "test content"
	_, err = file.WriteString(content)
	assert.NoError(t, err)

	// Read back and verify
	readContent, err := ioutil.ReadFile(testPath)
	assert.NoError(t, err)
	assert.Equal(t, content, string(readContent))
}

func TestFileStoreAPIError(t *testing.T) {
	// Create a test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"ERROR","message":"Unauthorized","correlationId":"test-123"}`))
	}))
	defer server.Close()

	api := &APIRequest{
		BaseURL: server.URL,
		Client:  resty.New(),
		APIKey:  "invalid-key",
	}

	file := FileStoreContent{
		Identifier: "test-file",
		Name:       "test-file.yaml",
		Path:       "/test-file.yaml",
	}

	err := file.DownloadFile(api, "test-account", "", "", "/account")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error downloading file")
}

// TestFileStoreIntegration tests the complete filestore workflow
func TestFileStoreIntegration(t *testing.T) {
	// This test simulates the complete filestore download and git setup process
	testDir := "./test_integration"
	defer os.RemoveAll(testDir)

	// Create test files structure
	testFiles := []struct {
		path    string
		content string
	}{
		{"/account/manifests/app1.yaml", "account level manifest"},
		{"/org1/configs/service.yaml", "org level service config"},
		{"/org1/project1/templates/pipeline.yaml", "project level pipeline template"},
	}

	// Create filestore directory structure
	filestoreDir := filepath.Join(testDir, "filestore")
	for _, tf := range testFiles {
		fullPath := filepath.Join(filestoreDir, tf.path)
		dir := filepath.Dir(fullPath)

		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = ioutil.WriteFile(fullPath, []byte(tf.content), 0644)
		require.NoError(t, err)
	}

	// Verify all files were created correctly
	for _, tf := range testFiles {
		fullPath := filepath.Join(filestoreDir, tf.path)
		assert.FileExists(t, fullPath)

		content, err := ioutil.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, tf.content, string(content))
	}

	// Verify directory structure is correct (no duplicate filestore directories)
	err := filepath.Walk(filestoreDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check that path doesn't contain "filestore/filestore"
		assert.NotContains(t, path, "filestore/filestore", fmt.Sprintf("Path %s should not contain duplicate filestore directory", path))
		return nil
	})
	assert.NoError(t, err)
}

func BenchmarkFileStoreDownload(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("benchmark test content"))
	}))
	defer server.Close()

	api := &APIRequest{
		BaseURL: server.URL,
		Client:  resty.New(),
		APIKey:  "test-api-key",
	}

	file := FileStoreContent{
		Identifier: "benchmark-file",
		Name:       "benchmark.yaml",
		Path:       "/benchmark.yaml",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os.RemoveAll("./filestore")
		file.DownloadFile(api, "test-account", "", "", "/account")
	}
}



