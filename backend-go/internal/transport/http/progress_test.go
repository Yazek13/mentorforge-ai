package httptransport

import (
	"encoding/json"
	"html"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"mentorforge-ai/backend-go/internal/service/progress"
)

func TestProgressPageReturnsHTML(t *testing.T) {
	response := httptest.NewRecorder()
	NewHandler().Routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/progress", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, expected %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("Content-Type = %q, expected HTML", contentType)
	}

	body := response.Body.String()
	for _, expected := range []string{
		"Мой прогресс",
		"Foundation",
		"Фундамент",
		"Go Engineer",
		"18%",
		"Python Engineer",
		"5%",
		"Backend / Architecture",
		"12%",
		"AI Engineer",
		"6%",
		"Готовность к работе",
		"14%",
		"6 процентных пунктов",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("progress page does not contain %q", expected)
		}
	}
}

func TestProgressObjectiveUsesExistingLessonWhenAvailable(t *testing.T) {
	handler := NewHandler()
	response := httptest.NewRecorder()
	handler.Routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/progress", nil))
	body := response.Body.String()

	lessonAvailable := false
	for _, lesson := range handler.service.ListLessons() {
		if lesson.Slug == "go-lesson-04" {
			lessonAvailable = true
			if !strings.Contains(body, `href="/lessons/`+strconv.Itoa(lesson.ID)+`"`) {
				t.Fatalf("existing objective lesson %d is not linked", lesson.ID)
			}
		}
	}
	if !lessonAvailable && !strings.Contains(body, "Следующий урок ещё не добавлен.") {
		t.Fatal("missing objective lesson must not create a broken link")
	}
}

func TestProgressPageDoesNotExposeSensitiveOrReferenceData(t *testing.T) {
	t.Setenv("MENTORFORGE_AI_PROVIDER", "openai")
	t.Setenv("OPENAI_API_KEY", "progress-test-secret-key")
	response := httptest.NewRecorder()
	handler := NewHandler()
	handler.Routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/progress", nil))
	body := response.Body.String()

	for _, forbidden := range []string{
		"progress-test-secret-key",
		"OPENAI_API_KEY",
		"simple_answer",
		"interview_answer",
		"required_points",
		"Программа — это набор инструкций, которые компьютер выполняет для решения задачи.",
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("progress page exposes forbidden value %q", forbidden)
		}
	}
	for _, question := range handler.service.ListQuestions() {
		for _, answer := range []string{question.SimpleAnswer, question.InterviewAnswer} {
			if answer != "" && (strings.Contains(body, answer) || strings.Contains(body, html.EscapeString(answer))) {
				t.Errorf("reference answer for question %d leaked", question.ID)
			}
		}
	}
}

func TestProgressLearningMetadataUsesExistingLessonsOnly(t *testing.T) {
	handler := NewHandler()
	response := httptest.NewRecorder()
	handler.Routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/progress", nil))
	body := response.Body.String()
	const marker = `<script type="application/json" id="mentorforge-progress-lessons">`
	_, payload, found := strings.Cut(body, marker)
	if !found {
		t.Fatal("missing progress lesson metadata")
	}
	payload, _, found = strings.Cut(payload, "</script>")
	if !found {
		t.Fatal("unterminated metadata")
	}
	var lessons []progressLessonView
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&lessons); err != nil {
		t.Fatal(err)
	}
	existing := handler.lessonsInLearningOrder()
	if len(lessons) != len(existing) {
		t.Fatalf("metadata contains %d lessons, repository has %d", len(lessons), len(existing))
	}
	for index, lesson := range lessons {
		want := existing[index]
		if lesson.ID != want.ID || lesson.Title != want.Title || lesson.PracticeTaskID != want.PracticeTaskID || lesson.URL != "/lessons/"+strconv.Itoa(want.ID) {
			t.Errorf("metadata differs from learning order: %#v", lesson)
		}
		questions := handler.service.ListQuestionsByLesson(want.ID)
		if len(lesson.QuestionIDs) != len(questions) {
			t.Fatalf("incorrect questions for lesson %d", want.ID)
		}
		for i, question := range questions {
			if lesson.QuestionIDs[i] != question.ID {
				t.Errorf("incorrect question ID for lesson %d", want.ID)
			}
		}
	}
	if strings.Contains(body, "mentorforge-export-data") {
		t.Fatal("career progress must not be added to learning export")
	}
}

func TestProgressMethodsAndStaticAssets(t *testing.T) {
	handler := NewHandler().Routes()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/progress", nil))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /progress = %d", response.Code)
	}
	for _, path := range []string{"/static/progress.js", "/static/progress.css"} {
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK || response.Body.Len() == 0 {
			t.Errorf("asset %s is unavailable", path)
		}
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(response.Body.String(), `href="/progress"`) {
		t.Fatal("home page must link directly to progress")
	}
}

func TestProgressTemplateUsesUpdatedAssessment(t *testing.T) {
	plan := progress.CurrentPlan()
	plan.Snapshot.Tracks[0].Percent = 22
	plan.JobReadiness = 24
	plan.CurrentStage = progress.CurrentStageFor(plan.JobReadiness)
	response := httptest.NewRecorder()
	NewHandler().renderHTML(response, http.StatusOK, "progress", progressPageData{PageTitle: "Test", Plan: plan})
	for _, value := range []string{`value="22"`, "Job Readiness: 24%", "Ученик-инженер", "11 процентных пунктов"} {
		if !strings.Contains(response.Body.String(), value) {
			t.Errorf("updated Go data is not rendered: %s", value)
		}
	}
}

func TestExistingHTMLRoutesStillWorkWithProgressPage(t *testing.T) {
	handler := NewHandler().Routes()
	paths := []string{
		"/",
		"/topics/1",
		"/lessons/1",
		"/lessons/1/questions",
		"/lessons/1/practice",
		"/lessons/1/complete",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("%s returned %d", path, response.Code)
			}
			if !strings.HasPrefix(response.Header().Get("Content-Type"), "text/html") {
				t.Fatalf("%s is not HTML", path)
			}
		})
	}
}

func TestExistingJSONAPIRoutesStillWorkWithProgressPage(t *testing.T) {
	handler := NewHandler().Routes()
	paths := []string{"/health", "/topics", "/questions", "/questions/1"}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("%s returned %d", path, response.Code)
			}
			if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/json") || !json.Valid(response.Body.Bytes()) {
				t.Fatalf("%s is not valid JSON", path)
			}
		})
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/reviews", strings.NewReader(`{"question_id":1,"grade":"good"}`)))
	if response.Code != http.StatusOK || !json.Valid(response.Body.Bytes()) {
		t.Fatalf("POST /reviews compatibility failed: status=%d body=%s", response.Code, response.Body.String())
	}
}
