package runtime

import "testing"

type fakeLocalPortChecker struct {
	unavailable map[int]bool
}

func (f fakeLocalPortChecker) Available(port int) bool {
	return !f.unavailable[port]
}

func TestNextAvailableLocalPort(t *testing.T) {
	tests := []struct {
		name      string
		preferred int
		reserved  map[int]struct{}
		checker   fakeLocalPortChecker
		want      int
		wantErr   bool
	}{
		{
			name:      "returns preferred port when available",
			preferred: 3000,
			reserved:  map[int]struct{}{},
			checker:   fakeLocalPortChecker{},
			want:      3000,
		},
		{
			name:      "skips reserved and host unavailable ports",
			preferred: 3000,
			reserved:  map[int]struct{}{3000: {}, 3001: {}},
			checker:   fakeLocalPortChecker{unavailable: map[int]bool{3002: true}},
			want:      3003,
		},
		{
			name:      "returns error when range is exhausted",
			preferred: 65535,
			reserved:  map[int]struct{}{65535: {}},
			checker:   fakeLocalPortChecker{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextAvailableLocalPort(tt.preferred, tt.reserved, tt.checker)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got port %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected port %d, got %d", tt.want, got)
			}
		})
	}
}
