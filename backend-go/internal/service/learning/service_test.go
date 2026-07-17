package learning

import (
	"errors"
	"testing"
	"time"
)

func TestReviewQuestionSchedulesNextReview(t *testing.T) {
	fixedTime := time.Date(2026, time.January, 10, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		grade string
		days  int
	}{
		{grade: "again", days: 0},
		{grade: "hard", days: 1},
		{grade: "good", days: 3},
		{grade: "easy", days: 7},
	}

	for _, test := range tests {
		t.Run(test.grade, func(t *testing.T) {
			service := NewService()
			service.now = func() time.Time { return fixedTime }

			review, err := service.ReviewQuestion(1, test.grade)
			if err != nil {
				t.Fatalf("ReviewQuestion() вернул ошибку: %v", err)
			}

			expected := fixedTime.AddDate(0, 0, test.days)
			if !review.NextReviewAt.Equal(expected) {
				t.Fatalf("NextReviewAt = %v, ожидалось %v", review.NextReviewAt, expected)
			}
		})
	}
}

func TestReviewQuestionValidatesInput(t *testing.T) {
	service := NewService()

	tests := []struct {
		name       string
		questionID int
		grade      string
		expected   error
	}{
		{name: "неизвестная оценка", questionID: 1, grade: "unknown", expected: ErrInvalidGrade},
		{name: "неизвестный вопрос", questionID: 999, grade: "good", expected: ErrQuestionNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.ReviewQuestion(test.questionID, test.grade)
			if !errors.Is(err, test.expected) {
				t.Fatalf("получена ошибка %v, ожидалась %v", err, test.expected)
			}
		})
	}
}

func TestStarterContent(t *testing.T) {
	service := NewService()

	if got := len(service.ListTopics()); got != 4 {
		t.Fatalf("количество тем = %d, ожидалось 4", got)
	}

	questions := service.ListQuestions()
	if got := len(questions); got != 30 {
		t.Fatalf("количество вопросов = %d, ожидалось 30", got)
	}

	for _, question := range questions {
		if len(question.FollowUpQuestions) != 2 {
			t.Errorf("у вопроса %d должно быть два дополнительных вопроса", question.ID)
		}
		if question.SimpleAnswer == "" || question.InterviewAnswer == "" || question.RealUsage == "" || question.PracticeTask == "" {
			t.Errorf("у вопроса %d заполнены не все учебные поля", question.ID)
		}
	}
}
