package entity

import "testing"

func TestStageStatusTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current StageStatus
		next    StageStatus
		want    bool
	}{
		{name: "active to paused", current: StageStatusActive, next: StageStatusPaused, want: true},
		{name: "active to archived", current: StageStatusActive, next: StageStatusArchived, want: true},
		{name: "paused to active", current: StageStatusPaused, next: StageStatusActive, want: true},
		{name: "archived to active", current: StageStatusArchived, next: StageStatusActive, want: true},
		{name: "archived to paused", current: StageStatusArchived, next: StageStatusPaused, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := canTransitionStageStatus(tt.current, tt.next); got != tt.want {
				t.Fatalf("canTransitionStageStatus(%q, %q) = %v, want %v", tt.current, tt.next, got, tt.want)
			}
		})
	}
}
