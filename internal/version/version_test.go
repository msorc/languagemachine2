package version

import "testing"

func TestStamped(t *testing.T) {
	old, oldDate := version, date
	t.Cleanup(func() { version, date = old, oldDate })
	if Version() == "" || Date() == "" {
		t.Fatalf("unstamped: version %q, date %q", Version(), Date())
	}
	version, date = "1.2.3", "20261008"
	if Version() != "1.2.3" || Date() != "20261008" {
		t.Errorf("stamped: version %q, date %q", Version(), Date())
	}
}
