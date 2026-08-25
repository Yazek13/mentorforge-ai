package httptransport

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"mentorforge-ai/backend-go/internal/domain"
	"mentorforge-ai/backend-go/internal/service/learning"
)

type topicCardView struct {
	ID            int
	Title         string
	Description   string
	LessonCount   int
	QuestionCount int
}

type topicQuestionView struct {
	ID       int
	Number   int
	Question string
	Level    string
}

type lessonSummaryView struct {
	ID            int
	Number        int
	Title         string
	Level         string
	QuestionCount int
	Questions     []topicQuestionView
}

// В данные экспорта входят только названия материала и собственные задания.
// Текст лекции и эталонные ответы здесь намеренно отсутствуют.
type exportQuestionView struct {
	ID       int    `json:"id"`
	LessonID int    `json:"lesson_id"`
	Level    string `json:"level"`
	Question string `json:"question"`
}

type exportPracticeView struct {
	ID           int      `json:"id"`
	LessonID     int      `json:"lesson_id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Requirements []string `json:"requirements"`
	Language     string   `json:"language"`
}

type exportLessonView struct {
	ID        int                  `json:"id"`
	TopicID   int                  `json:"topic_id"`
	Topic     string               `json:"topic"`
	TopicSlug string               `json:"topic_slug"`
	Slug      string               `json:"slug"`
	Title     string               `json:"title"`
	Level     string               `json:"level"`
	Questions []exportQuestionView `json:"questions"`
	Practice  exportPracticeView   `json:"practice"`
}

type exportDataView struct {
	Scope    string             `json:"scope"`
	FileName string             `json:"file_name"`
	Lessons  []exportLessonView `json:"lessons"`
}

type lessonRuntimeView struct {
	LessonID       int   `json:"lesson_id"`
	QuestionIDs    []int `json:"question_ids"`
	PracticeTaskID int   `json:"practice_task_id"`
	NextLessonID   int   `json:"next_lesson_id,omitempty"`
}

type homePageData struct {
	PageTitle     string
	TopicCount    int
	LessonCount   int
	QuestionCount int
	Topics        []topicCardView
	ExportData    exportDataView
}

type topicPageData struct {
	PageTitle     string
	Topic         domain.Topic
	Lessons       []lessonSummaryView
	LessonCount   int
	QuestionCount int
	HasLessons    bool
	FirstLessonID int
	ExportData    exportDataView
}

type lessonPageData struct {
	PageTitle        string
	Topic            domain.Topic
	Lesson           domain.Lesson
	Level            string
	LessonNumber     int
	TopicLessonCount int
	QuestionCount    int
	Stage            string
	RuntimeData      lessonRuntimeView
}

type lessonQuestionView struct {
	Number   int
	Question domain.Question
}

type lessonQuestionsPageData struct {
	PageTitle        string
	Topic            domain.Topic
	Lesson           domain.Lesson
	Level            string
	LessonNumber     int
	TopicLessonCount int
	Questions        []lessonQuestionView
	QuestionCount    int
	Stage            string
	RuntimeData      lessonRuntimeView
}

type lessonPracticePageData struct {
	PageTitle        string
	Topic            domain.Topic
	Lesson           domain.Lesson
	Practice         domain.PracticeTask
	Level            string
	LessonNumber     int
	TopicLessonCount int
	QuestionCount    int
	Stage            string
	CodeTask         bool
	RuntimeData      lessonRuntimeView
}

type lessonCompletePageData struct {
	PageTitle        string
	Topic            domain.Topic
	Lesson           domain.Lesson
	Practice         domain.PracticeTask
	Level            string
	LessonNumber     int
	TopicLessonCount int
	QuestionCount    int
	Stage            string
	HasNextLesson    bool
	NextLessonID     int
	NextLessonLabel  string
	RuntimeData      lessonRuntimeView
	ExportData       exportDataView
}

type errorPageData struct {
	PageTitle string
	Status    int
	Title     string
	Message   string
	BackURL   string
	BackLabel string
}

type lessonMaterial struct {
	Topic            domain.Topic
	Lesson           domain.Lesson
	Questions        []domain.Question
	Practice         domain.PracticeTask
	LessonNumber     int
	TopicLessonCount int
	NextLessonID     int
	NextLessonLabel  string
}

func (h *Handler) handleHomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.handleNotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		h.renderPageError(w, http.StatusMethodNotAllowed, "Метод не поддерживается", "Для этой страницы доступен только GET-запрос.")
		return
	}

	topics := h.service.ListTopics()
	lessons := h.lessonsInLearningOrder()
	questions := h.service.ListQuestions()
	lessonCounts := make(map[int]int)
	questionCounts := make(map[int]int)

	for _, lesson := range lessons {
		lessonCounts[lesson.TopicID]++
	}
	for _, question := range questions {
		questionCounts[question.TopicID]++
	}

	topicCards := make([]topicCardView, 0, len(topics))
	for _, topic := range topics {
		topicCards = append(topicCards, topicCardView{
			ID:            topic.ID,
			Title:         topic.Title,
			Description:   topic.Description,
			LessonCount:   lessonCounts[topic.ID],
			QuestionCount: questionCounts[topic.ID],
		})
	}

	h.renderHTML(w, http.StatusOK, "index", homePageData{
		PageTitle:     "Главная",
		TopicCount:    len(topics),
		LessonCount:   len(lessons),
		QuestionCount: len(questions),
		Topics:        topicCards,
		ExportData: exportDataView{
			Scope:    "all",
			FileName: "mentorforge_all_submissions.md",
			Lessons:  h.buildExportLessons(lessons),
		},
	})
}

func (h *Handler) handleTopicPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.renderPageError(w, http.StatusMethodNotAllowed, "Метод не поддерживается", "Для страницы темы доступен только GET-запрос.")
		return
	}

	topicID, ok := parsePageID(r.URL.Path, "/topics/")
	if !ok {
		h.renderPageError(w, http.StatusBadRequest, "Некорректный адрес темы", "Проверьте идентификатор темы в адресной строке.")
		return
	}

	topic, err := h.service.GetTopic(topicID)
	if err != nil {
		h.renderPageError(w, http.StatusNotFound, "Тема не найдена", "Запрошенной учебной темы пока нет.")
		return
	}

	lessons := h.service.ListLessonsByTopic(topic.ID)
	lessonViews := make([]lessonSummaryView, 0, len(lessons))
	totalQuestions := 0
	for lessonIndex, lesson := range lessons {
		questions := h.service.ListQuestionsByLesson(lesson.ID)
		questionViews := make([]topicQuestionView, 0, len(questions))
		for questionIndex, question := range questions {
			questionViews = append(questionViews, topicQuestionView{
				ID:       question.ID,
				Number:   questionIndex + 1,
				Question: question.Question,
				Level:    levelLabel(question.Level),
			})
		}

		totalQuestions += len(questions)
		lessonViews = append(lessonViews, lessonSummaryView{
			ID:            lesson.ID,
			Number:        lessonIndex + 1,
			Title:         lesson.Title,
			Level:         levelLabel(lesson.Level),
			QuestionCount: len(questions),
			Questions:     questionViews,
		})
	}

	data := topicPageData{
		PageTitle:     topic.Title,
		Topic:         topic,
		Lessons:       lessonViews,
		LessonCount:   len(lessons),
		QuestionCount: totalQuestions,
		HasLessons:    len(lessons) > 0,
		ExportData: exportDataView{
			Scope:    "topic",
			FileName: exportFileNameForTopic(topic),
			Lessons:  h.buildExportLessons(lessons),
		},
	}
	if data.HasLessons {
		data.FirstLessonID = lessons[0].ID
	}

	h.renderHTML(w, http.StatusOK, "topic", data)
}

func (h *Handler) handleLearnPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.renderPageError(w, http.StatusMethodNotAllowed, "Метод не поддерживается", "Для учебной страницы доступен только GET-запрос.")
		return
	}

	lessons := h.service.ListLessons()
	if len(lessons) == 0 {
		h.renderPageError(w, http.StatusNotFound, "Учебные уроки пока не добавлены", "Вернитесь позднее, чтобы открыть первый материал.")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/lessons/%d", lessons[0].ID), http.StatusSeeOther)
}

// Старые ссылки на отдельные вопросы сохраняются, но теперь всегда начинают урок с лекции.
func (h *Handler) handleLearnQuestionPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.renderPageError(w, http.StatusMethodNotAllowed, "Метод не поддерживается", "Для учебной страницы доступен только GET-запрос.")
		return
	}

	questionID, ok := parsePageID(r.URL.Path, "/learn/")
	if !ok {
		h.renderPageError(w, http.StatusBadRequest, "Некорректный адрес вопроса", "Проверьте идентификатор вопроса в адресной строке.")
		return
	}

	question, err := h.service.GetQuestion(questionID)
	if err != nil {
		h.renderPageError(w, http.StatusNotFound, "Вопрос не найден", "Запрошенного учебного вопроса пока нет.")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/lessons/%d", question.LessonID), http.StatusSeeOther)
}

func (h *Handler) handleLessonRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.renderPageError(w, http.StatusMethodNotAllowed, "Метод не поддерживается", "Для учебных страниц доступен только GET-запрос.")
		return
	}

	lessonID, stage, validID, validPath := parseLessonPath(r.URL.Path)
	if !validID {
		h.renderPageError(w, http.StatusBadRequest, "Некорректный адрес урока", "Проверьте идентификатор урока в адресной строке.")
		return
	}
	if !validPath {
		h.renderPageError(w, http.StatusNotFound, "Страница урока не найдена", "Запрошенного этапа урока не существует.")
		return
	}

	material, err := h.loadLessonMaterial(lessonID)
	if err != nil {
		if errors.Is(err, learning.ErrLessonNotFound) {
			h.renderPageError(w, http.StatusNotFound, "Урок не найден", "Запрошенного учебного урока пока нет.")
			return
		}

		h.renderPageError(w, http.StatusInternalServerError, "Внутренняя ошибка", "Не удалось подготовить связанные материалы урока.")
		return
	}

	switch stage {
	case "lecture":
		h.renderLessonPage(w, material)
	case "questions":
		h.renderLessonQuestionsPage(w, material)
	case "practice":
		h.renderLessonPracticePage(w, material)
	case "complete":
		h.renderLessonCompletePage(w, material)
	}
}

func (h *Handler) renderLessonPage(w http.ResponseWriter, material lessonMaterial) {
	h.renderHTML(w, http.StatusOK, "lesson", lessonPageData{
		PageTitle:        material.Lesson.Title,
		Topic:            material.Topic,
		Lesson:           material.Lesson,
		Level:            levelLabel(material.Lesson.Level),
		LessonNumber:     material.LessonNumber,
		TopicLessonCount: material.TopicLessonCount,
		QuestionCount:    len(material.Questions),
		Stage:            "lecture",
		RuntimeData:      buildRuntimeData(material),
	})
}

func (h *Handler) renderLessonQuestionsPage(w http.ResponseWriter, material lessonMaterial) {
	questions := make([]lessonQuestionView, 0, len(material.Questions))
	for index, question := range material.Questions {
		questions = append(questions, lessonQuestionView{Number: index + 1, Question: question})
	}

	h.renderHTML(w, http.StatusOK, "lesson-questions", lessonQuestionsPageData{
		PageTitle:        "Вопросы — " + material.Lesson.Title,
		Topic:            material.Topic,
		Lesson:           material.Lesson,
		Level:            levelLabel(material.Lesson.Level),
		LessonNumber:     material.LessonNumber,
		TopicLessonCount: material.TopicLessonCount,
		Questions:        questions,
		QuestionCount:    len(questions),
		Stage:            "questions",
		RuntimeData:      buildRuntimeData(material),
	})
}

func (h *Handler) renderLessonPracticePage(w http.ResponseWriter, material lessonMaterial) {
	h.renderHTML(w, http.StatusOK, "lesson-practice", lessonPracticePageData{
		PageTitle:        "Практика — " + material.Lesson.Title,
		Topic:            material.Topic,
		Lesson:           material.Lesson,
		Practice:         material.Practice,
		Level:            levelLabel(material.Lesson.Level),
		LessonNumber:     material.LessonNumber,
		TopicLessonCount: material.TopicLessonCount,
		QuestionCount:    len(material.Questions),
		Stage:            "practice",
		CodeTask:         material.Practice.Language != "text",
		RuntimeData:      buildRuntimeData(material),
	})
}

func (h *Handler) renderLessonCompletePage(w http.ResponseWriter, material lessonMaterial) {
	exportData := exportDataView{
		Scope:    "lesson",
		FileName: exportFileNameForLesson(material.Lesson),
		Lessons:  h.buildExportLessons([]domain.Lesson{material.Lesson}),
	}

	h.renderHTML(w, http.StatusOK, "lesson-complete", lessonCompletePageData{
		PageTitle:        "Завершение — " + material.Lesson.Title,
		Topic:            material.Topic,
		Lesson:           material.Lesson,
		Practice:         material.Practice,
		Level:            levelLabel(material.Lesson.Level),
		LessonNumber:     material.LessonNumber,
		TopicLessonCount: material.TopicLessonCount,
		QuestionCount:    len(material.Questions),
		Stage:            "complete",
		HasNextLesson:    material.NextLessonID > 0,
		NextLessonID:     material.NextLessonID,
		NextLessonLabel:  material.NextLessonLabel,
		RuntimeData:      buildRuntimeData(material),
		ExportData:       exportData,
	})
}

func (h *Handler) loadLessonMaterial(lessonID int) (lessonMaterial, error) {
	lesson, err := h.service.GetLesson(lessonID)
	if err != nil {
		return lessonMaterial{}, err
	}

	topic, err := h.service.GetTopic(lesson.TopicID)
	if err != nil {
		return lessonMaterial{}, err
	}

	practice, err := h.service.GetPracticeTask(lesson.PracticeTaskID)
	if err != nil {
		return lessonMaterial{}, err
	}

	topicLessons := h.service.ListLessonsByTopic(topic.ID)
	lessonNumber := 0
	for index, item := range topicLessons {
		if item.ID == lesson.ID {
			lessonNumber = index + 1
			break
		}
	}
	if lessonNumber == 0 {
		return lessonMaterial{}, learning.ErrLessonNotFound
	}

	nextLessonID := 0
	nextLessonLabel := ""
	orderedLessons := h.lessonsInLearningOrder()
	for index, item := range orderedLessons {
		if item.ID == lesson.ID && index+1 < len(orderedLessons) {
			nextLesson := orderedLessons[index+1]
			nextLessonID = nextLesson.ID
			if nextLesson.TopicID == lesson.TopicID {
				for nextIndex, topicLesson := range h.service.ListLessonsByTopic(nextLesson.TopicID) {
					if topicLesson.ID == nextLesson.ID {
						nextLessonLabel = fmt.Sprintf("Перейти к уроку %d", nextIndex+1)
						break
					}
				}
			} else {
				nextLessonLabel = "Перейти к следующей теме"
			}
			break
		}
	}

	return lessonMaterial{
		Topic:            topic,
		Lesson:           lesson,
		Questions:        h.service.ListQuestionsByLesson(lesson.ID),
		Practice:         practice,
		LessonNumber:     lessonNumber,
		TopicLessonCount: len(topicLessons),
		NextLessonID:     nextLessonID,
		NextLessonLabel:  nextLessonLabel,
	}, nil
}

// lessonsInLearningOrder сохраняет порядок тем и порядок уроков внутри каждой темы.
// Это позволяет добавлять урок с новым ID, не перенумеровывая существующие материалы.
func (h *Handler) lessonsInLearningOrder() []domain.Lesson {
	lessons := make([]domain.Lesson, 0, len(h.service.ListLessons()))
	for _, topic := range h.service.ListTopics() {
		lessons = append(lessons, h.service.ListLessonsByTopic(topic.ID)...)
	}

	return lessons
}

func buildRuntimeData(material lessonMaterial) lessonRuntimeView {
	questionIDs := make([]int, 0, len(material.Questions))
	for _, question := range material.Questions {
		questionIDs = append(questionIDs, question.ID)
	}

	return lessonRuntimeView{
		LessonID:       material.Lesson.ID,
		QuestionIDs:    questionIDs,
		PracticeTaskID: material.Practice.ID,
		NextLessonID:   material.NextLessonID,
	}
}

func (h *Handler) buildExportLessons(lessons []domain.Lesson) []exportLessonView {
	result := make([]exportLessonView, 0, len(lessons))
	for _, lesson := range lessons {
		topic, topicErr := h.service.GetTopic(lesson.TopicID)
		practice, practiceErr := h.service.GetPracticeTask(lesson.PracticeTaskID)
		if topicErr != nil || practiceErr != nil {
			continue
		}

		questions := h.service.ListQuestionsByLesson(lesson.ID)
		exportQuestions := make([]exportQuestionView, 0, len(questions))
		for _, question := range questions {
			exportQuestions = append(exportQuestions, exportQuestionView{
				ID:       question.ID,
				LessonID: lesson.ID,
				Level:    question.Level,
				Question: question.Question,
			})
		}

		result = append(result, exportLessonView{
			ID:        lesson.ID,
			TopicID:   topic.ID,
			Topic:     topic.Title,
			TopicSlug: topic.Slug,
			Slug:      lesson.Slug,
			Title:     lesson.Title,
			Level:     lesson.Level,
			Questions: exportQuestions,
			Practice: exportPracticeView{
				ID:           practice.ID,
				LessonID:     lesson.ID,
				Title:        practice.Title,
				Description:  practice.Description,
				Requirements: practice.Requirements,
				Language:     practice.Language,
			},
		})
	}

	return result
}

func (h *Handler) renderPageError(w http.ResponseWriter, status int, title, message string) {
	h.renderHTML(w, status, "not_found", errorPageData{
		PageTitle: title,
		Status:    status,
		Title:     title,
		Message:   message,
		BackURL:   "/",
		BackLabel: "Вернуться на главную",
	})
}

func (h *Handler) renderHTML(w http.ResponseWriter, status int, templateName string, data any) {
	var body bytes.Buffer
	if err := h.templates.ExecuteTemplate(&body, templateName, data); err != nil {
		log.Printf("cannot render template %s: %v", templateName, err)
		writeHTMLInternalError(w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if _, err := w.Write(body.Bytes()); err != nil {
		log.Printf("cannot write HTML response: %v", err)
	}
}

func writeHTMLInternalError(w http.ResponseWriter) {
	const page = `<!doctype html><html lang="ru"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Внутренняя ошибка — MentorForge AI</title></head><body><main><h1>Внутренняя ошибка</h1><p>Не удалось открыть страницу. Попробуйте ещё раз.</p><a href="/">Вернуться на главную</a></main></body></html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(page))
}

func parsePageID(path, prefix string) (int, bool) {
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}

	idPart := strings.TrimPrefix(path, prefix)
	if idPart == "" || strings.Contains(idPart, "/") {
		return 0, false
	}

	id, err := strconv.Atoi(idPart)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}

func parseLessonPath(path string) (lessonID int, stage string, validID, validPath bool) {
	if !strings.HasPrefix(path, "/lessons/") {
		return 0, "", false, false
	}

	parts := strings.Split(strings.TrimPrefix(path, "/lessons/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, "", false, false
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		return 0, "", false, false
	}

	if len(parts) == 1 {
		return id, "lecture", true, true
	}
	if len(parts) != 2 {
		return id, "", true, false
	}

	switch parts[1] {
	case "questions", "practice", "complete":
		return id, parts[1], true, true
	default:
		return id, "", true, false
	}
}

func levelLabel(level string) string {
	switch level {
	case "beginner":
		return "Начальный"
	case "intermediate":
		return "Средний"
	case "advanced":
		return "Продвинутый"
	default:
		return level
	}
}

func safeSlug(value string, fallback string) string {
	var result strings.Builder
	previousUnderscore := false

	for _, symbol := range strings.ToLower(value) {
		isLetter := symbol >= 'a' && symbol <= 'z'
		isDigit := symbol >= '0' && symbol <= '9'
		if isLetter || isDigit {
			result.WriteRune(symbol)
			previousUnderscore = false
			continue
		}

		if !previousUnderscore && result.Len() > 0 {
			result.WriteByte('_')
			previousUnderscore = true
		}
	}

	slug := strings.Trim(result.String(), "_")
	if slug == "" {
		return fallback
	}

	return slug
}

func exportFileNameForTopic(topic domain.Topic) string {
	slug := safeSlug(topic.Slug, fmt.Sprintf("topic_%d", topic.ID))
	return "mentorforge_submission_" + slug + ".md"
}

func exportFileNameForLesson(lesson domain.Lesson) string {
	slug := safeSlug(lesson.Slug, fmt.Sprintf("lesson_%d", lesson.ID))
	return "mentorforge_submission_" + slug + ".md"
}
