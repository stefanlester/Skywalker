package skywalker

import (
	"testing"

	"github.com/stefanlester/skywalker/filesystems"
)

// clearFSGateEnv unsets every gate/config env var that createFileSystems reads,
// so a test starts from a known-empty configuration. t.Setenv registers a
// cleanup that restores the original values after the test, keeping this
// hermetic and safe to run alongside other tests.
func clearFSGateEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"MINIO_SECRET", "MINIO_ENDPOINT", "MINIO_KEY", "MINIO_USESSL", "MINIO_REGION", "MINIO_BUCKET",
		"S3_SECRET", "S3_KEY", "S3_REGION", "S3_ENDPOINT", "S3_BUCKET",
		"SFTP_HOST", "SFTP_USER", "SFTP_PASS", "SFTP_PORT",
		"WEBDAV_HOST", "WEBDAV_USER", "WEBDAV_PASS",
	} {
		t.Setenv(key, "")
	}
}

// TestSkywalker_createFileSystems_none verifies that with no gate vars set, the
// returned map is empty. Construction is offline (no client is opened), so this
// runs without any external service.
func TestSkywalker_createFileSystems_none(t *testing.T) {
	clearFSGateEnv(t)

	var s Skywalker
	fs := s.createFileSystems()

	if len(fs) != 0 {
		t.Errorf("expected empty file systems map, got %d entries: %v", len(fs), keysOf(fs))
	}
}

// TestSkywalker_createFileSystems_all verifies that when all four gate vars are
// set, the map contains exactly those four keys and every value type-asserts to
// filesystems.FS. The interface assertion is the core regression guard: each
// backend must be stored as a pointer, because its FS methods use pointer
// receivers. Storing a value would compile here but fail the assertion the
// engine relies on (FileSystems[fsType].(filesystems.FS)).
func TestSkywalker_createFileSystems_all(t *testing.T) {
	clearFSGateEnv(t)

	// Only the gate vars matter for wiring; the rest may stay empty. No client
	// is opened during construction, so no server is contacted.
	t.Setenv("MINIO_SECRET", "test-minio-secret")
	t.Setenv("S3_SECRET", "test-s3-secret")
	t.Setenv("SFTP_HOST", "sftp.example.test")
	t.Setenv("WEBDAV_HOST", "https://webdav.example.test")

	var s Skywalker
	fs := s.createFileSystems()

	wantKeys := []string{"MINIO", "S3", "SFTP", "WEBDAV"}
	if len(fs) != len(wantKeys) {
		t.Fatalf("expected %d file systems, got %d: %v", len(wantKeys), len(fs), keysOf(fs))
	}

	for _, key := range wantKeys {
		val, ok := fs[key]
		if !ok {
			t.Errorf("expected file systems map to contain key %q", key)
			continue
		}
		if _, ok := val.(filesystems.FS); !ok {
			t.Errorf("value for %q (%T) does not satisfy filesystems.FS; backend must be stored as a pointer", key, val)
		}
	}
}

// TestSkywalker_createFileSystems_perBackend table-drives one gate var at a
// time and asserts that exactly the matching backend is wired, keyed correctly,
// and satisfies filesystems.FS. This localizes any single-backend wiring
// regression to the offending row.
func TestSkywalker_createFileSystems_perBackend(t *testing.T) {
	tests := []struct {
		name    string
		gateKey string
		wantKey string
	}{
		{name: "minio", gateKey: "MINIO_SECRET", wantKey: "MINIO"},
		{name: "s3", gateKey: "S3_SECRET", wantKey: "S3"},
		{name: "sftp", gateKey: "SFTP_HOST", wantKey: "SFTP"},
		{name: "webdav", gateKey: "WEBDAV_HOST", wantKey: "WEBDAV"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearFSGateEnv(t)
			t.Setenv(tt.gateKey, "gate-value")

			var s Skywalker
			fs := s.createFileSystems()

			if len(fs) != 1 {
				t.Fatalf("expected exactly 1 file system for %s, got %d: %v", tt.name, len(fs), keysOf(fs))
			}

			val, ok := fs[tt.wantKey]
			if !ok {
				t.Fatalf("expected key %q for %s, got keys %v", tt.wantKey, tt.name, keysOf(fs))
			}
			if _, ok := val.(filesystems.FS); !ok {
				t.Errorf("value for %q (%T) does not satisfy filesystems.FS", tt.wantKey, val)
			}
		})
	}
}

// keysOf returns the keys of a file systems map for readable failure messages.
func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
