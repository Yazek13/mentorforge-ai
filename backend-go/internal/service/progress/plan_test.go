package progress

import "testing"

func TestCurrentSnapshotPercentages(t *testing.T) {
	plan := CurrentPlan()
	want := map[string]int{
		TrackGo:      18,
		TrackPython:  5,
		TrackBackend: 12,
		TrackAI:      6,
		TrackJob:     14,
	}

	if plan.Snapshot.Date != "2026-09-04" {
		t.Fatalf("snapshot date = %q", plan.Snapshot.Date)
	}
	if len(plan.Snapshot.Tracks) != len(want) {
		t.Fatalf("tracks = %d, expected %d", len(plan.Snapshot.Tracks), len(want))
	}
	for _, track := range plan.Snapshot.Tracks {
		if track.Percent != want[track.ID] {
			t.Errorf("track %q = %d, expected %d", track.ID, track.Percent, want[track.ID])
		}
		delete(want, track.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing tracks: %#v", want)
	}
}

func TestCurrentStageForJobReadinessFourteen(t *testing.T) {
	current := CurrentStageFor(14)
	if current.Stage.Code != "Foundation" || current.Stage.Title != "Фундамент" {
		t.Fatalf("stage = %#v, expected Foundation", current.Stage)
	}
	if !current.HasNext || current.PointsToNext != 6 {
		t.Fatalf("next stage calculation = %#v", current)
	}
}

func TestCareerStageBoundaries(t *testing.T) {
	tests := []struct {
		percent int
		code    string
	}{
		{19, "Foundation"},
		{20, "Apprentice Engineer"},
		{34, "Apprentice Engineer"},
		{35, "Junior Track"},
		{49, "Junior Track"},
		{50, "Job Search Preparation"},
		{59, "Job Search Preparation"},
		{60, "Ready for Applications"},
		{74, "Ready for Applications"},
		{75, "Working Engineer"},
		{89, "Working Engineer"},
		{90, "AI / Backend Engineer"},
		{96, "AI / Backend Engineer"},
		{97, "AI Architect"},
		{100, "AI Architect"},
	}

	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			if got := CareerStageFor(test.percent).Code; got != test.code {
				t.Fatalf("CareerStageFor(%d) = %q, expected %q", test.percent, got, test.code)
			}
		})
	}
}

func TestPercentIsClampedToDisplayRange(t *testing.T) {
	tests := []struct {
		value int
		want  int
	}{
		{-100, 0},
		{-1, 0},
		{0, 0},
		{50, 50},
		{100, 100},
		{101, 100},
		{900, 100},
	}

	for _, test := range tests {
		track := newTrack("test", "Test", test.value, "", "", nil, nil)
		if track.Percent != test.want {
			t.Errorf("percent %d displayed as %d, expected %d", test.value, track.Percent, test.want)
		}
	}
}

func TestAIJourneyMatchesImplementedCapabilities(t *testing.T) {
	stages := CurrentPlan().AIJourney
	if len(stages) != 11 {
		t.Fatalf("AI Journey stages = %d, expected 11", len(stages))
	}

	for index, stage := range stages {
		want := JourneyLocked
		if index < 3 {
			want = JourneyComplete
		}
		if stage.Status != want {
			t.Errorf("stage %d %q = %q, expected %q", stage.Number, stage.Title, stage.Status, want)
		}
	}
}
