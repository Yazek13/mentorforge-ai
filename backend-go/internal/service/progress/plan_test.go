package progress

import (
	"fmt"
	"testing"

	"mentorforge-ai/backend-go/internal/service/learning"
)

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
		{0, "Foundation"},
		{-1, "Foundation"},
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
		{101, "AI Architect"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprint(test.percent), func(t *testing.T) {
			if got := CareerStageFor(test.percent).Code; got != test.code {
				t.Fatalf("CareerStageFor(%d) = %q, expected %q", test.percent, got, test.code)
			}
			current := CurrentStageFor(test.percent)
			if current.Stage.Code != test.code {
				t.Fatalf("CurrentStageFor(%d) = %#v", test.percent, current)
			}
			if current.HasNext {
				if current.PointsToNext != current.Stage.MaxPercent+1-clampPercent(test.percent) {
					t.Errorf("incorrect next-stage distance: %#v", current)
				}
			} else if current.PointsToNext != 0 || current.NextTitle != "" {
				t.Errorf("terminal stage must not promise a next stage: %#v", current)
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
	plan := CurrentPlan()
	stages := plan.AIJourney
	if len(stages) != 11 {
		t.Fatalf("AI Journey stages = %d, expected 11", len(stages))
	}

	for index, stage := range stages {
		want := JourneyLocked
		if index == 0 {
			want = JourneyComplete
		}
		if stage.Status != want {
			t.Errorf("stage %d %q = %q, expected %q", stage.Number, stage.Title, stage.Status, want)
		}
	}
	if plan.AIJourneyComplete != 1 || plan.AIJourneyTotal != 11 || plan.AIJourneyCurrent != "Learning Core" || plan.AIJourneyNext != "AI Tutor architecture" {
		t.Fatalf("journey summary does not match repository capabilities: %#v", plan)
	}
	// Learning Core is backed by working lesson/question/practice relationships.
	service := learning.NewService()
	if len(service.ListLessons()) == 0 {
		t.Fatal("Learning Core cannot be complete without lessons")
	}
	for _, lesson := range service.ListLessons() {
		if len(lesson.TheorySections) == 0 || len(service.ListQuestionsByLesson(lesson.ID)) == 0 {
			t.Errorf("lesson %d lacks learning material", lesson.ID)
		}
		if task, err := service.GetPracticeTask(lesson.PracticeTaskID); err != nil || task.LessonID != lesson.ID {
			t.Errorf("lesson %d lacks its practice task", lesson.ID)
		}
	}
}

func TestJourneyFrontier(t *testing.T) {
	for _, tc := range []struct {
		name          string
		stages        []AIJourneyStage
		current, next string
	}{
		{name: "empty"},
		{name: "locked", stages: []AIJourneyStage{{Title: "Core", Status: JourneyLocked}}, next: "Core"},
		{name: "in progress", stages: []AIJourneyStage{{Title: "Core", Status: JourneyComplete}, {Title: "Tutor", Status: JourneyInProgress}}, current: "Core", next: "Tutor"},
		{name: "complete", stages: []AIJourneyStage{{Title: "Core", Status: JourneyComplete}}, current: "Core"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			current, next := journeyFrontier(tc.stages)
			if current != tc.current || next != tc.next {
				t.Fatalf("frontier = (%q, %q), want (%q, %q)", current, next, tc.current, tc.next)
			}
		})
	}
}

func TestJobSearchRoadmapAndSnapshotAgree(t *testing.T) {
	plan := CurrentPlan()
	if plan.JobReadiness != trackPercent(plan.Snapshot.Tracks, TrackJob) || plan.FirstApplications != 60 {
		t.Fatal("job search values differ from assessment or application threshold")
	}
	for percent := -1; percent <= 101; percent++ {
		current := 0
		for _, stage := range currentJobSearchRoadmap(percent) {
			if stage.Current {
				current++
				if clampPercent(percent) < stage.MinPercent || clampPercent(percent) > stage.MaxPercent {
					t.Fatalf("wrong job stage for %d: %#v", percent, stage)
				}
			}
		}
		if current != 1 {
			t.Fatalf("%d current job stages for %d", current, percent)
		}
	}
}
