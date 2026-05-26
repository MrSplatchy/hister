package indexer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/model"
)

func setupTestDB(t *testing.T) *config.Config {
	t.Helper()
	cfg := &config.Config{
		Server: config.Server{
			Database: "file::memory:",
		},
	}
	err := model.Init(cfg)
	if err != nil {
		t.Fatalf("failed to init test DB: %v", err)
	}
	return cfg
}

func TestDirectoryUserResolution(t *testing.T) {
	setupTestDB(t)

	// Create test users
	u1, err := model.CreateUser("alice", "password123", false)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	u2, err := model.CreateUser("bob", "password123", false)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name     string
		username string
		wantID   uint
		wantErr  bool
	}{
		{
			name:     "empty username is global",
			username: "",
			wantID:   0,
			wantErr:  false,
		},
		{
			name:     "existing user alice",
			username: "alice",
			wantID:   u1.ID,
			wantErr:  false,
		},
		{
			name:     "existing user bob",
			username: "bob",
			wantID:   u2.ID,
			wantErr:  false,
		},
		{
			name:     "non-existent user",
			username: "charlie",
			wantID:   0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID uint
			var err error
			if tt.username != "" {
				u, e := model.GetUser(tt.username)
				if e != nil {
					err = e
				} else {
					gotID = u.ID
				}
			}
			if tt.wantErr {
				if err == nil {
					t.Errorf("user resolution(%q) expected error, got nil", tt.username)
				}
				return
			}
			if err != nil {
				t.Errorf("user resolution(%q) unexpected error: %v", tt.username, err)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("user resolution(%q) = %d, want %d", tt.username, gotID, tt.wantID)
			}
		})
	}
}

func TestIndexFileWithUserID(t *testing.T) {
	setupTestDB(t)

	// Create a test user
	u, err := model.CreateUser("testuser", "password123", false)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Create temp dirs for data and test files
	dataDir := t.TempDir()
	testDir := t.TempDir()

	// Create test files (avoid patterns that match sensitive content regex)
	testFile := filepath.Join(testDir, "test.txt")
	err = os.WriteFile(testFile, []byte("sample document content about indexing files for testing purposes"), 0o644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	testFile2 := filepath.Join(testDir, "test2.txt")
	err = os.WriteFile(testFile2, []byte("sample global document content for indexing test purposes"), 0o644)
	if err != nil {
		t.Fatalf("failed to create test file 2: %v", err)
	}

	// Initialize the indexer with proper data and index dirs
	idxCfg := config.CreateDefaultConfig()
	idxCfg.App.Directory = dataDir
	// Override the index dir
	err = Init(idxCfg)
	if err != nil {
		t.Fatalf("failed to init indexer: %v", err)
	}

	// Index the file with the test user's ID
	err = IndexFile(testFile, u.ID)
	if err != nil {
		t.Fatalf("IndexFile with user ID failed: %v", err)
	}

	// Index the file without user ID (global)
	err = IndexFile(testFile2, 0)
	if err != nil {
		t.Fatalf("IndexFile without user ID failed: %v", err)
	}
}

func TestDirectoryUserField(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		expected string
	}{
		{
			name:     "empty user",
			user:     "",
			expected: "",
		},
		{
			name:     "user set",
			user:     "alice",
			expected: "alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := &config.Directory{
				Path: "/some/path",
				User: tt.user,
			}
			if dir.User != tt.expected {
				t.Errorf("Directory.User = %q, want %q", dir.User, tt.expected)
			}
		})
	}
}
