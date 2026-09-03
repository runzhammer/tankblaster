package computerplayers

import "testing"

func TestStrongerIDIncreasesWithoutWrapping(t *testing.T) {
	tests := []struct {
		id   ID
		want ID
	}{
		{DoedelID, FrederikID},
		{FrederikID, MisterXID},
		{MisterXID, DrNukeID},
		{DrNukeID, HaraldID},
		{HaraldID, HaraldID},
		{ID(99), DoedelID},
	}

	for _, tt := range tests {
		if got := StrongerID(tt.id); got != tt.want {
			t.Fatalf("StrongerID(%v) = %v, want %v", tt.id, got, tt.want)
		}
	}
}
