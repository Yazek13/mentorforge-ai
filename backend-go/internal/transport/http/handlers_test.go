package httptransport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mentorforge-ai/backend-go/internal/domain"
)

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	NewHandler().Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("Content-Type = %q, ожидался JSON", contentType)
	}

	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("ответ не является JSON: %v", err)
	}
	if body["status"] != "ok" || body["service"] != "mentorforge-ai" {
		t.Fatalf("неожиданный ответ: %#v", body)
	}
}

func TestCreateReview(t *testing.T) {
	requestBody := strings.NewReader(`{"question_id":1,"grade":"good"}`)
	request := httptest.NewRequest(http.MethodPost, "/reviews", requestBody)
	response := httptest.NewRecorder()

	NewHandler().Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался %d; ответ: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var body struct {
		QuestionID   int    `json:"question_id"`
		Grade        string `json:"grade"`
		NextReviewAt string `json:"next_review_at"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("ответ не является JSON: %v", err)
	}
	if body.QuestionID != 1 || body.Grade != "good" || body.NextReviewAt == "" {
		t.Fatalf("неожиданный ответ: %#v", body)
	}
}

func TestUnknownRouteReturnsJSONError(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	response := httptest.NewRecorder()

	NewHandler().Routes().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("статус = %d, ожидался %d", response.Code, http.StatusNotFound)
	}

	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("ошибка API не является JSON: %v", err)
	}
	if body["error"] == "" {
		t.Fatal("в JSON-ошибке отсутствует поле error")
	}
}

func TestHTMLLessonPages(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedCode int
		expectedText string
	}{
		{name: "главная", path: "/", expectedCode: http.StatusOK, expectedText: "Доступные темы"},
		{name: "страница темы", path: "/topics/1", expectedCode: http.StatusOK, expectedText: "Уроки темы"},
		{name: "лекция", path: "/lessons/1", expectedCode: http.StatusOK, expectedText: "Что такое программа и язык Go"},
		{name: "вопросы", path: "/lessons/1/questions", expectedCode: http.StatusOK, expectedText: "Сначала прочитайте лекцию"},
		{name: "практика", path: "/lessons/1/practice", expectedCode: http.StatusOK, expectedText: "Сначала завершите предыдущие этапы"},
		{name: "завершение", path: "/lessons/1/complete", expectedCode: http.StatusOK, expectedText: "Результаты для наставника"},
		{name: "лекция Go 3", path: "/lessons/7", expectedCode: http.StatusOK, expectedText: "Условия, циклы и коллекции Go"},
		{name: "вопросы Go 3", path: "/lessons/7/questions", expectedCode: http.StatusOK, expectedText: "Гарантирован ли порядок обхода map?"},
		{name: "практика Go 3", path: "/lessons/7/practice", expectedCode: http.StatusOK, expectedText: "Каталог учебных карточек"},
		{name: "урок не найден", path: "/lessons/999", expectedCode: http.StatusNotFound, expectedText: "Урок не найден"},
		{name: "тема не найдена", path: "/topics/999", expectedCode: http.StatusNotFound, expectedText: "Тема не найдена"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()

			NewHandler().Routes().ServeHTTP(response, request)

			if response.Code != test.expectedCode {
				t.Fatalf("статус = %d, ожидался %d", response.Code, test.expectedCode)
			}
			if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
				t.Fatalf("Content-Type = %q, ожидался HTML", contentType)
			}
			if !strings.Contains(response.Body.String(), test.expectedText) {
				t.Fatalf("ответ не содержит текст %q", test.expectedText)
			}
		})
	}
}

func TestLegacyLearnRoutesOpenLectureFirst(t *testing.T) {
	tests := []struct {
		path     string
		location string
	}{
		{path: "/learn", location: "/lessons/1"},
		{path: "/learn/2", location: "/lessons/1"},
		{path: "/learn/11", location: "/lessons/3"},
		{path: "/learn/31", location: "/lessons/7"},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			NewHandler().Routes().ServeHTTP(response, request)

			if response.Code != http.StatusSeeOther {
				t.Fatalf("статус = %d, ожидался %d", response.Code, http.StatusSeeOther)
			}
			if location := response.Header().Get("Location"); location != test.location {
				t.Fatalf("Location = %q, ожидался %q", location, test.location)
			}
		})
	}
}

func TestReferenceAnswersAreHiddenUntilSubmission(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/lessons/1/questions", nil)
	response := httptest.NewRecorder()

	NewHandler().Routes().ServeHTTP(response, request)
	body := response.Body.String()

	if !strings.Contains(body, `data-answer-comparison hidden`) {
		t.Fatal("блок сравнения должен быть скрыт при открытии страницы")
	}
	if !strings.Contains(body, "Эталон откроется только после фиксации ответа") {
		t.Fatal("на странице отсутствует пояснение о фиксации ответа")
	}
	if !strings.Contains(body, "Программа — это набор инструкций") {
		t.Fatal("страница не содержит эталон из учебного хранилища")
	}
}

func TestExistingJSONEndpointsRemainJSON(t *testing.T) {
	paths := []string{"/topics", "/questions", "/questions/1", "/questions/39"}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()

			NewHandler().Routes().ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("статус = %d, ожидался %d", response.Code, http.StatusOK)
			}
			if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
				t.Fatalf("Content-Type = %q, ожидался JSON", contentType)
			}
			if !json.Valid(response.Body.Bytes()) {
				t.Fatal("ответ endpoint не является корректным JSON")
			}
		})
	}
}

func TestExportDataContainsNoReferenceMaterials(t *testing.T) {
	tests := []struct {
		path          string
		expectedScope string
		expectedFile  string
		expectedCount int
	}{
		{path: "/", expectedScope: "all", expectedFile: "mentorforge_all_submissions.md", expectedCount: 7},
		{path: "/topics/1", expectedScope: "topic", expectedFile: "mentorforge_submission_go_basics.md", expectedCount: 3},
		{path: "/lessons/1/complete", expectedScope: "lesson", expectedFile: "mentorforge_submission_go_lesson_01.md", expectedCount: 1},
		{path: "/lessons/7/complete", expectedScope: "lesson", expectedFile: "mentorforge_submission_go_lesson_03.md", expectedCount: 1},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			NewHandler().Routes().ServeHTTP(response, request)

			payloadJSON := extractExportJSON(t, response.Body.String())
			var payload struct {
				Scope    string                       `json:"scope"`
				FileName string                       `json:"file_name"`
				Lessons  []map[string]json.RawMessage `json:"lessons"`
			}
			if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
				t.Fatalf("JSON-блок экспорта некорректен: %v", err)
			}

			if payload.Scope != test.expectedScope || payload.FileName != test.expectedFile {
				t.Fatalf("неожиданные параметры экспорта: scope=%q, file=%q", payload.Scope, payload.FileName)
			}
			if len(payload.Lessons) != test.expectedCount {
				t.Fatalf("количество уроков для экспорта = %d, ожидалось %d", len(payload.Lessons), test.expectedCount)
			}

			for _, forbidden := range []string{"simple_answer", "interview_answer", "required_points", "theory_sections", "key_points", "hints", "starter_code", "expected_result", "expected_output"} {
				if strings.Contains(payloadJSON, forbidden) {
					t.Fatalf("в экспортные данные попало запрещённое поле %q", forbidden)
				}
			}
		})
	}
}

func TestExportFileNamesUseSafeSlugs(t *testing.T) {
	topics := []struct {
		topic    domain.Topic
		expected string
	}{
		{topic: domain.Topic{ID: 1, Slug: "go-basics"}, expected: "mentorforge_submission_go_basics.md"},
		{topic: domain.Topic{ID: 3, Slug: "ai-ml-basics"}, expected: "mentorforge_submission_ai_ml_basics.md"},
		{topic: domain.Topic{ID: 8, Slug: ""}, expected: "mentorforge_submission_topic_8.md"},
	}

	for _, test := range topics {
		if got := exportFileNameForTopic(test.topic); got != test.expected {
			t.Errorf("имя файла = %q, ожидалось %q", got, test.expected)
		}
	}

	lesson := domain.Lesson{ID: 1, Slug: "go-lesson-01"}
	if got := exportFileNameForLesson(lesson); got != "mentorforge_submission_go_lesson_01.md" {
		t.Errorf("имя файла урока = %q", got)
	}
	thirdGoLesson := domain.Lesson{ID: 7, Slug: "go-lesson-03"}
	if got := exportFileNameForLesson(thirdGoLesson); got != "mentorforge_submission_go_lesson_03.md" {
		t.Errorf("имя файла третьего урока Go = %q", got)
	}
}

func TestWorkflowControlsAreRendered(t *testing.T) {
	pages := []struct {
		path     string
		expected []string
	}{
		{path: "/", expected: []string{"Скачать все результаты", "mentorforge_all_submissions.md"}},
		{path: "/topics/1", expected: []string{"Скачать результаты темы", "mentorforge_submission_go_basics.md"}},
		{path: "/lessons/1", expected: []string{"Я прочитал лекцию — перейти к вопросам", "go run main.go"}},
		{path: "/lessons/1/questions", expected: []string{"data-question-editor", "Сохранить черновик", "Зафиксировать ответ", "Перейти к практике"}},
		{path: "/lessons/1/practice", expected: []string{"Моё решение", "Как работает моё решение", "Результат запуска", "Что было сложно", "Завершить практику"}},
		{path: "/lessons/1/complete", expected: []string{"Скачать результаты урока", "Пройти урок заново"}},
		{path: "/lessons/7", expected: []string{"Зачем нужны условия", "Порядок обхода map", "Ожидаемый вывод"}},
		{path: "/lessons/7/practice", expected: []string{"Каталог учебных карточек", "Проверил успешный и ошибочный сценарии", "Вставляйте полный вывод программы"}},
		{path: "/lessons/7/complete", expected: []string{"mentorforge_submission_go_lesson_03.md", "Скачать результаты урока"}},
	}

	for _, page := range pages {
		t.Run(page.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, page.path, nil)
			response := httptest.NewRecorder()
			NewHandler().Routes().ServeHTTP(response, request)

			body := response.Body.String()
			for _, expected := range page.expected {
				if !strings.Contains(body, expected) {
					t.Fatalf("страница не содержит %q", expected)
				}
			}
		})
	}
}

func TestLessonTwoNavigatesToThirdGoLesson(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/lessons/2/complete", nil)
	response := httptest.NewRecorder()
	NewHandler().Routes().ServeHTTP(response, request)

	body := response.Body.String()
	if !strings.Contains(body, `href="/lessons/7"`) {
		t.Fatal("завершение урока 2 не ведёт к новому уроку Go")
	}
	if !strings.Contains(body, ">Перейти к уроку 3</a>") {
		t.Fatal("кнопка перехода к уроку 3 имеет неверную подпись")
	}
}

func TestGoTopicShowsThirdLessonAfterSecond(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/topics/1", nil)
	response := httptest.NewRecorder()
	NewHandler().Routes().ServeHTTP(response, request)

	body := response.Body.String()
	secondPosition := strings.Index(body, "Базовые элементы кода Go")
	thirdPosition := strings.Index(body, "Условия, циклы и коллекции Go")
	if secondPosition == -1 || thirdPosition == -1 || thirdPosition <= secondPosition {
		t.Fatal("урок 3 должен отображаться после урока 2 в теме Go Basics")
	}
}

func extractExportJSON(t *testing.T, body string) string {
	t.Helper()

	const startTag = `<script type="application/json" id="mentorforge-export-data">`
	const endTag = `</script>`

	start := strings.Index(body, startTag)
	if start == -1 {
		t.Fatal("JSON-блок экспорта не найден")
	}
	start += len(startTag)

	end := strings.Index(body[start:], endTag)
	if end == -1 {
		t.Fatal("JSON-блок экспорта не закрыт")
	}

	return body[start : start+end]
}
