package httptransport

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mentorforge-ai/backend-go/internal/service/learning"
	webassets "mentorforge-ai/backend-go/web"
)

// Handler объединяет HTTP-роуты и сервис.
type Handler struct {
	service     *learning.Service
	templates   *template.Template
	staticFiles http.Handler
}

// NewHandler создаёт HTTP-обработчик с готовым сервисом.
func NewHandler() *Handler {
	return &Handler{
		service:     learning.NewService(),
		templates:   template.Must(template.ParseFS(webassets.Files, "templates/*.html")),
		staticFiles: newStaticFileHandler(),
	}
}

// Routes возвращает http.Handler с зарегистрированными endpoint.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", h.staticFiles))
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/topics", h.handleTopics)
	mux.HandleFunc("/topics/", h.handleTopicPage)
	mux.HandleFunc("/questions", h.handleQuestions)
	mux.HandleFunc("/questions/", h.handleQuestionByID)
	mux.HandleFunc("/reviews", h.handleReviews)
	mux.HandleFunc("/lessons/", h.handleLessonRoute)
	mux.HandleFunc("/learn", h.handleLearnPage)
	mux.HandleFunc("/learn/", h.handleLearnQuestionPage)
	mux.HandleFunc("/", h.handleHomePage)
	return mux
}

func newStaticFileHandler() http.Handler {
	staticFiles, err := fs.Sub(webassets.Files, "static")
	if err != nil {
		// Ошибка означает, что статические файлы не попали в сборку приложения.
		panic("static files are not embedded")
	}

	return http.FileServer(http.FS(staticFiles))
}

func (h *Handler) handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "route not found")
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "mentorforge-ai",
	})
}

func (h *Handler) handleTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, h.service.ListTopics())
}

func (h *Handler) handleQuestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, h.service.ListQuestions())
}

func (h *Handler) handleQuestionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, ok := parseQuestionID(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid question id")
		return
	}

	question, err := h.service.GetQuestion(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "question not found")
		return
	}

	writeJSON(w, http.StatusOK, question)
}

func (h *Handler) handleReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var request struct {
		QuestionID int    `json:"question_id"`
		Grade      string `json:"grade"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if request.QuestionID <= 0 {
		writeError(w, http.StatusBadRequest, "question_id must be positive")
		return
	}

	review, err := h.service.ReviewQuestion(request.QuestionID, request.Grade)
	if err != nil {
		switch {
		case errors.Is(err, learning.ErrQuestionNotFound):
			writeError(w, http.StatusNotFound, "question not found")
		case errors.Is(err, learning.ErrInvalidGrade):
			writeError(w, http.StatusBadRequest, "invalid grade")
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"question_id":    review.QuestionID,
		"grade":          review.Grade,
		"next_review_at": review.NextReviewAt.Format(time.RFC3339),
	})
}

func parseQuestionID(path string) (int, bool) {
	idPart := strings.TrimPrefix(path, "/questions/")
	if idPart == "" || strings.Contains(idPart, "/") {
		return 0, false
	}

	id, err := strconv.Atoi(idPart)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}
