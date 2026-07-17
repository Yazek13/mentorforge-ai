package domain

import "time"

// Review хранит результат повторения вопроса.
type Review struct {
	QuestionID   int       `json:"question_id"`
	Grade        string    `json:"grade"`
	ReviewedAt   time.Time `json:"reviewed_at"`
	NextReviewAt time.Time `json:"next_review_at"`
}
