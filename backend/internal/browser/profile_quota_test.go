package browser

import (
	"ant-chrome/backend/internal/config"
	"testing"
	"time"
)

func TestCreateIgnoresLegacyProfileLimit(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	mgr := NewManager(cfg, t.TempDir())
	mgr.Profiles["existing"] = &Profile{
		ProfileId:   "existing",
		ProfileName: "existing",
		UserDataDir: "existing",
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	created, err := mgr.Create(ProfileInput{ProfileName: "new-profile"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created == nil {
		t.Fatal("Create returned nil profile")
	}
	if len(mgr.Profiles) != 2 {
		t.Fatalf("expected 2 profiles after create, got %d", len(mgr.Profiles))
	}
}

func TestCopyIgnoresLegacyProfileLimit(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	now := time.Now().Format(time.RFC3339)
	mgr := NewManager(cfg, t.TempDir())
	mgr.Profiles["existing"] = &Profile{
		ProfileId:       "existing",
		ProfileName:     "existing",
		UserDataDir:     "existing",
		FingerprintArgs: []string{"--fingerprint-brand=Chrome"},
		LaunchArgs:      []string{"--disable-sync"},
		Tags:            []string{"tag-a"},
		Keywords:        []string{"keyword-a"},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	cloned, err := mgr.Copy("existing", "")
	if err != nil {
		t.Fatalf("Copy returned error: %v", err)
	}
	if cloned == nil {
		t.Fatal("Copy returned nil profile")
	}
	if len(mgr.Profiles) != 2 {
		t.Fatalf("expected 2 profiles after copy, got %d", len(mgr.Profiles))
	}
}
