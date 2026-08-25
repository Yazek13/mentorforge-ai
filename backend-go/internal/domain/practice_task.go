package domain

// PracticeTask описывает отдельное практическое задание урока.
type PracticeTask struct {
	ID             int      `json:"id"`
	LessonID       int      `json:"lesson_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Requirements   []string `json:"requirements"`
	ExpectedResult string   `json:"expected_result"`
	Hints          []string `json:"hints"`
	StarterCode    string   `json:"starter_code"`
	Language       string   `json:"language"`
}
