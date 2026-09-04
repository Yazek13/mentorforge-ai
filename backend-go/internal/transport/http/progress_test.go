package httptransport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
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
	NewHandler().Routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/progress", nil))
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
