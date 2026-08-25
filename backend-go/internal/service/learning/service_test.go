package learning

import (
	"errors"
	"strings"
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
	if got := len(questions); got != 39 {
		t.Fatalf("количество вопросов = %d, ожидалось 39", got)
	}

	for _, question := range questions {
		if len(question.FollowUpQuestions) != 2 {
			t.Errorf("у вопроса %d должно быть два дополнительных вопроса", question.ID)
		}
		if question.SimpleAnswer == "" || question.InterviewAnswer == "" || question.RealUsage == "" || question.PracticeTask == "" {
			t.Errorf("у вопроса %d заполнены не все учебные поля", question.ID)
		}
		if question.LessonID == 0 || len(question.RequiredPoints) == 0 {
			t.Errorf("вопрос %d не связан с уроком или не содержит обязательных пунктов", question.ID)
		}
	}
}

func TestFirstLessonHasLectureQuestionsAndPractice(t *testing.T) {
	service := NewService()

	lessons := service.ListLessons()
	if got := len(lessons); got != 7 {
		t.Fatalf("количество уроков = %d, ожидалось 7", got)
	}

	lesson, err := service.GetLesson(1)
	if err != nil {
		t.Fatalf("первый урок не найден: %v", err)
	}
	if lesson.Title != "Что такое программа и язык Go" {
		t.Fatalf("название первого урока = %q", lesson.Title)
	}
	if len(lesson.TheorySections) < 5 || len(lesson.Terms) == 0 || len(lesson.CodeExamples) == 0 {
		t.Fatal("первая лекция заполнена не полностью")
	}
	if len(lesson.QuestionIDs) != 5 || lesson.PracticeTaskID != 1 {
		t.Fatalf("неверные связи первого урока: вопросы=%v, практика=%d", lesson.QuestionIDs, lesson.PracticeTaskID)
	}

	questions := service.ListQuestionsByLesson(lesson.ID)
	expectedQuestions := []string{
		"Что такое программа?",
		"Что такое язык программирования?",
		"Что такое Go?",
		"Что означает компиляция?",
		"Для чего нужна функция main?",
	}
	if len(questions) != len(expectedQuestions) {
		t.Fatalf("вопросов первого урока = %d, ожидалось %d", len(questions), len(expectedQuestions))
	}
	for index, expected := range expectedQuestions {
		if questions[index].Question != expected || questions[index].LessonID != lesson.ID {
			t.Errorf("вопрос %d = %q, lesson_id=%d", index+1, questions[index].Question, questions[index].LessonID)
		}
	}

	practice, err := service.GetPracticeTask(lesson.PracticeTaskID)
	if err != nil {
		t.Fatalf("практика первого урока не найдена: %v", err)
	}
	if practice.LessonID != lesson.ID || practice.Language != "go" || len(practice.Requirements) == 0 {
		t.Fatalf("практика первого урока заполнена неверно: %#v", practice)
	}
	if practice.StarterCode == "" || practice.ExpectedResult == "" {
		t.Fatal("в практике отсутствует стартовый код или ожидаемый результат")
	}
}

func TestGoCollectionsLessonHasCompleteMaterial(t *testing.T) {
	service := NewService()

	lesson, err := service.GetLesson(7)
	if err != nil {
		t.Fatalf("третий урок Go не найден: %v", err)
	}
	if lesson.TopicID != 1 || lesson.Slug != "go-lesson-03" || lesson.Title != "Условия, циклы и коллекции Go" {
		t.Fatalf("неверные данные урока 3: %#v", lesson)
	}
	if len(lesson.TheorySections) != 10 {
		t.Fatalf("разделов лекции = %d, ожидалось 10", len(lesson.TheorySections))
	}
	if len(lesson.Terms) != 18 {
		t.Fatalf("терминов = %d, ожидалось 18", len(lesson.Terms))
	}
	if len(lesson.CodeExamples) < 5 || len(lesson.KeyPoints) < 8 || len(lesson.CommonMistakes) < 8 {
		t.Fatal("в лекции недостаточно примеров, ключевых мыслей или частых ошибок")
	}
	if len(lesson.QuestionIDs) != 9 || lesson.PracticeTaskID != 7 {
		t.Fatalf("неверные связи урока: вопросы=%v, практика=%d", lesson.QuestionIDs, lesson.PracticeTaskID)
	}

	questions := service.ListQuestionsByLesson(lesson.ID)
	if len(questions) != 9 {
		t.Fatalf("вопросов урока = %d, ожидалось 9", len(questions))
	}
	for index, question := range questions {
		if question.ID != 31+index || question.TopicID != 1 || question.LessonID != lesson.ID {
			t.Errorf("неверная связь вопроса: %#v", question)
		}
		if question.SimpleAnswer == "" || question.InterviewAnswer == "" || len(question.RequiredPoints) == 0 || len(question.FollowUpQuestions) != 2 {
			t.Errorf("вопрос %d заполнен не полностью", question.ID)
		}
	}
	if questions[7].Question != "Как проверить, существует ли ключ в map?" || !strings.Contains(questions[7].InterviewAnswer, "ok") {
		t.Fatal("вопрос о безопасном поиске не объясняет переменную ok")
	}
	if questions[8].Question != "Гарантирован ли порядок обхода map?" || !strings.Contains(questions[8].SimpleAnswer, "не гарантируется") {
		t.Fatal("вопрос о порядке map сформулирован недостаточно явно")
	}

	practice, err := service.GetPracticeTask(lesson.PracticeTaskID)
	if err != nil {
		t.Fatalf("практика урока 3 не найдена: %v", err)
	}
	if practice.LessonID != lesson.ID || practice.Title != "Каталог учебных карточек" || practice.Language != "go" {
		t.Fatalf("неверные данные практики: %#v", practice)
	}
	if len(practice.Requirements) < 15 || len(practice.Hints) != 6 {
		t.Fatalf("практика заполнена не полностью: требований=%d, подсказок=%d", len(practice.Requirements), len(practice.Hints))
	}
	if strings.Contains(practice.StarterCode, "func filterByMinLevel(cards []StudyCard, minLevel int) ([]StudyCard, error) {") ||
		strings.Contains(practice.StarterCode, "func findCardByTitle(cards map[string]StudyCard, title string) (StudyCard, error) {") {
		t.Fatal("стартовый материал не должен содержать реализации функций")
	}
}
