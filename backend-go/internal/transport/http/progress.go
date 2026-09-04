package httptransport

import (
	"fmt"
	"net/http"

	"mentorforge-ai/backend-go/internal/service/progress"
)

type progressLessonView struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	QuestionIDs    []int  `json:"question_ids"`
	PracticeTaskID int    `json:"practice_task_id"`
}

type progressObjectiveView struct {
	Title           string
	Goal            string
	LessonAvailable bool
	LessonURL       string
	LessonTitle     string
}

type progressPageData struct {
	PageTitle       string
	Plan            progress.Plan
	Objective       progressObjectiveView
	VerifiedLessons []progressLessonView
}

func (h *Handler) handleProgressPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.renderPageError(w, http.StatusMethodNotAllowed, "Метод не поддерживается", "Для страницы прогресса доступен только GET-запрос.")
		return
	}

	plan := progress.CurrentPlan()
	objective := progressObjectiveView{
		Title: plan.Objective.Title,
		Goal:  plan.Objective.Goal,
	}
	lessons := h.lessonsInLearningOrder()
	verifiedLessons := make([]progressLessonView, 0, len(lessons))

	for _, lesson := range lessons {
		questionIDs := make([]int, 0, len(lesson.QuestionIDs))
		for _, question := range h.service.ListQuestionsByLesson(lesson.ID) {
			questionIDs = append(questionIDs, question.ID)
		}

		lessonURL := fmt.Sprintf("/lessons/%d", lesson.ID)
		verifiedLessons = append(verifiedLessons, progressLessonView{
			ID:             lesson.ID,
			Title:          lesson.Title,
			URL:            lessonURL,
			QuestionIDs:    questionIDs,
			PracticeTaskID: lesson.PracticeTaskID,
		})

		if lesson.Slug == plan.Objective.LessonSlug {
			objective.LessonAvailable = true
			objective.LessonURL = lessonURL
			objective.LessonTitle = lesson.Title
		}
	}

	h.renderHTML(w, http.StatusOK, "progress", progressPageData{
		PageTitle:       "Мой прогресс",
		Plan:            plan,
		Objective:       objective,
		VerifiedLessons: verifiedLessons,
	})
}
