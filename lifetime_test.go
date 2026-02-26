package oak

import "testing"

func TestLifetime_String(t *testing.T) {
	tests := []struct {
		l    Lifetime
		want string
	}{
		{Singleton, "singleton"},
		{Transient, "transient"},
		{Lifetime(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.l.String(); got != tt.want {
			t.Errorf("Lifetime(%d).String() = %q, want %q", tt.l, got, tt.want)
		}
	}
}
