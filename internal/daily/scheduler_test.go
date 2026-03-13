package daily

import (
	"testing"
	"time"
)

func TestEffectiveDate(t *testing.T) {
	tests := []struct {
		name      string
		resetHour int
		now       time.Time
		want      string
	}{
		{
			name:      "reset at 0, afternoon",
			resetHour: 0,
			now:       time.Date(2026, 3, 13, 15, 0, 0, 0, time.UTC),
			want:      "2026-03-13",
		},
		{
			name:      "reset at 6, before reset",
			resetHour: 6,
			now:       time.Date(2026, 3, 13, 3, 0, 0, 0, time.UTC),
			want:      "2026-03-12",
		},
		{
			name:      "reset at 6, after reset",
			resetHour: 6,
			now:       time.Date(2026, 3, 13, 8, 0, 0, 0, time.UTC),
			want:      "2026-03-13",
		},
		{
			name:      "reset at 6, exactly at reset",
			resetHour: 6,
			now:       time.Date(2026, 3, 13, 6, 0, 0, 0, time.UTC),
			want:      "2026-03-13",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{resetHour: tt.resetHour}
			got := s.EffectiveDate(tt.now)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
