package memory

import (
	"sort"
	"sync"

	"mentorforge-ai/backend-go/internal/domain"
)

// Storage хранит данные в памяти.
type Storage struct {
	mu            sync.RWMutex
	topics        map[int]domain.Topic
	lessons       map[int]domain.Lesson
	questions     map[int]domain.Question
	practiceTasks map[int]domain.PracticeTask
	reviews       map[int]domain.Review
}

// NewStorage создаёт in-memory storage и копирует стартовые данные.
func NewStorage(topics []domain.Topic, lessons []domain.Lesson, questions []domain.Question, practiceTasks []domain.PracticeTask) *Storage {
	s := &Storage{
		topics:        make(map[int]domain.Topic, len(topics)),
		lessons:       make(map[int]domain.Lesson, len(lessons)),
		questions:     make(map[int]domain.Question, len(questions)),
		practiceTasks: make(map[int]domain.PracticeTask, len(practiceTasks)),
		reviews:       make(map[int]domain.Review),
	}

	for _, topic := range topics {
		s.topics[topic.ID] = topic
	}

	for _, question := range questions {
		s.questions[question.ID] = question
	}

	for _, lesson := range lessons {
		s.lessons[lesson.ID] = lesson
	}

	for _, practiceTask := range practiceTasks {
		s.practiceTasks[practiceTask.ID] = practiceTask
	}

	return s
}

// ListLessons возвращает уроки в порядке их ID.
func (s *Storage) ListLessons() []domain.Lesson {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]domain.Lesson, 0, len(s.lessons))
	for _, lesson := range s.lessons {
		items = append(items, lesson)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items
}

// GetLesson возвращает урок по ID.
func (s *Storage) GetLesson(id int) (domain.Lesson, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lesson, ok := s.lessons[id]
	return lesson, ok
}

// ListTopics возвращает темы в порядке их ID.
func (s *Storage) ListTopics() []domain.Topic {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]domain.Topic, 0, len(s.topics))
	for _, topic := range s.topics {
		items = append(items, topic)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items
}

// GetTopic возвращает тему по ID.
func (s *Storage) GetTopic(id int) (domain.Topic, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	topic, ok := s.topics[id]
	return topic, ok
}

// ListQuestions возвращает вопросы в порядке их ID.
func (s *Storage) ListQuestions() []domain.Question {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]domain.Question, 0, len(s.questions))
	for _, question := range s.questions {
		items = append(items, question)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items
}

// GetQuestion возвращает вопрос по ID.
func (s *Storage) GetQuestion(id int) (domain.Question, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	question, ok := s.questions[id]
	return question, ok
}

// GetPracticeTask возвращает практическое задание по ID.
func (s *Storage) GetPracticeTask(id int) (domain.PracticeTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	practiceTask, ok := s.practiceTasks[id]
	return practiceTask, ok
}

// SaveReview сохраняет результат повторения.
func (s *Storage) SaveReview(review domain.Review) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reviews[review.QuestionID] = review
}
