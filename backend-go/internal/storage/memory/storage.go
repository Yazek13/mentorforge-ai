package memory

import (
	"sort"
	"sync"

	"mentorforge-ai/backend-go/internal/domain"
)

// Storage хранит данные в памяти.
type Storage struct {
	mu        sync.RWMutex
	topics    map[int]domain.Topic
	questions map[int]domain.Question
	reviews   map[int]domain.Review
}

// NewStorage создаёт in-memory storage и копирует стартовые данные.
func NewStorage(topics []domain.Topic, questions []domain.Question) *Storage {
	s := &Storage{
		topics:    make(map[int]domain.Topic, len(topics)),
		questions: make(map[int]domain.Question, len(questions)),
		reviews:   make(map[int]domain.Review),
	}

	for _, topic := range topics {
		s.topics[topic.ID] = topic
	}

	for _, question := range questions {
		s.questions[question.ID] = question
	}

	return s
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

// SaveReview сохраняет результат повторения.
func (s *Storage) SaveReview(review domain.Review) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reviews[review.QuestionID] = review
}
