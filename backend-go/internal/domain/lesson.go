package domain

// Lesson описывает один последовательный учебный урок.
type Lesson struct {
	ID             int             `json:"id"`
	TopicID        int             `json:"topic_id"`
	Slug           string          `json:"slug"`
	Title          string          `json:"title"`
	Level          string          `json:"level"`
	Goal           string          `json:"goal"`
	Introduction   string          `json:"introduction"`
	TheorySections []TheorySection `json:"theory_sections"`
	Terms          []Term          `json:"terms"`
	CodeExamples   []CodeExample   `json:"code_examples"`
	RealUsage      []string        `json:"real_usage"`
	KeyPoints      []string        `json:"key_points"`
	CommonMistakes []string        `json:"common_mistakes"`
	Summary        string          `json:"summary"`
	QuestionIDs    []int           `json:"question_ids"`
	PracticeTaskID int             `json:"practice_task_id"`
}

// TheorySection содержит один логический раздел лекции.
type TheorySection struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Term объясняет новый термин до его дальнейшего использования.
type Term struct {
	Name        string `json:"name"`
	Explanation string `json:"explanation"`
}

// CodeExample содержит безопасный учебный пример и его разбор.
type CodeExample struct {
	Title       string `json:"title"`
	Language    string `json:"language"`
	Code        string `json:"code"`
	Explanation string `json:"explanation"`
	// ExpectedOutput остаётся пустым, если пример ничего не выводит.
	ExpectedOutput string `json:"expected_output,omitempty"`
}
