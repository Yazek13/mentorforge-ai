package httptransport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
