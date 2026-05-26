package files

import (
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

func TestFindDirUser(t *testing.T) {
	setupTestDB(t)

	// Create test users
	u1, err := model.CreateUser("alice", "password123", false)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	dirs := []*config.Directory{
		{Path: "/home/alice/docs", User: "alice"},
		{Path: "/home/charlie/docs", User: "charlie"},
		{Path: "/shared/docs", User: ""},
		{Path: "/tmp", User: ""},
	}

	tests := []struct {
		name    string
		path    string
		wantID  uint
		wantErr bool
	}{
		{
			name:    "file in alice's directory",
			path:    "/home/alice/docs/file.txt",
			wantID:  u1.ID,
			wantErr: false,
		},
		{
			name:    "file in nested alice directory",
			path:    "/home/alice/docs/subdir/file.txt",
			wantID:  u1.ID,
			wantErr: false,
		},
		{
			name:    "file in shared global directory",
			path:    "/shared/docs/file.txt",
			wantID:  0,
			wantErr: false,
		},
		{
			name:    "file in tmp global directory",
			path:    "/tmp/file.txt",
			wantID:  0,
			wantErr: false,
		},
		{
			name:    "file not in any directory",
			path:    "/other/path/file.txt",
			wantID:  0,
			wantErr: false,
		},
		{
			name:    "non-existent user in config",
			path:    "/home/charlie/docs/file.txt",
			wantID:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := FindDirUser(dirs, tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("FindDirUser(%q) expected error, got nil", tt.path)
				}
				return
			}
			if err != nil {
				t.Errorf("FindDirUser(%q) unexpected error: %v", tt.path, err)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("FindDirUser(%q) = %d, want %d", tt.path, gotID, tt.wantID)
			}
		})
	}
}

func TestFindDirUserWithHomeExpansion(t *testing.T) {
	setupTestDB(t)

	home := ExpandHome("~/")
	if home == "~/ " {
		t.Skip("could not expand home directory")
	}

	// Create a test user
	_, err := model.CreateUser("homeuser", "password123", false)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	dirs := []*config.Directory{
		{Path: "~/project", User: "homeuser"},
	}

	tests := []struct {
		name        string
		path        string
		wantNonZero bool
	}{
		{
			name:        "tilde path matches",
			path:        filepath.Join(home, "project", "file.txt"),
			wantNonZero: true,
		},
		{
			name:        "tilde path does not match",
			path:        filepath.Join(home, "other", "file.txt"),
			wantNonZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := FindDirUser(dirs, tt.path)
			if err != nil {
				t.Errorf("FindDirUser(%q) unexpected error: %v", tt.path, err)
				return
			}
			if tt.wantNonZero && gotID == 0 {
				t.Errorf("FindDirUser(%q) = 0, expected non-zero user ID", tt.path)
			}
			if !tt.wantNonZero && gotID != 0 {
				t.Errorf("FindDirUser(%q) = %d, expected 0", tt.path, gotID)
			}
		})
	}
}
