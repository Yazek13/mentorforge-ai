package domain

// Question описывает учебный вопрос и материалы для повторения.
type Question struct {
	ID                int      `json:"id"`
	TopicID           int      `json:"topic_id"`
	LessonID          int      `json:"lesson_id"`
	Level             string   `json:"level"`
	Question          string   `json:"question"`
	SimpleAnswer      string   `json:"simple_answer"`
	InterviewAnswer   string   `json:"interview_answer"`
	RealUsage         string   `json:"real_usage"`
	FollowUpQuestions []string `json:"follow_up_questions"`
	PracticeTask      string   `json:"practice_task"`
	RequiredPoints    []string `json:"required_points"`
}
