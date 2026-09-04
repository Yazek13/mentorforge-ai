// Package progress contains the mentor assessment and the career roadmap.
// Lesson completion remains a separate browser-side fact.
package progress

const (
	TrackGo      = "go"
	TrackPython  = "python"
	TrackBackend = "backend"
	TrackAI      = "ai"
	TrackJob     = "job-readiness"
)

type ProgressTrack struct {
	ID          string
	Title       string
	Percent     int
	Description string
	NextGoal    string
	Evidence    []string
	Roadmap     []string
}

// AssessmentSnapshot is a dated mentor assessment. Keeping the date with the
// values allows a later version to store several snapshots without changing UI data.
type AssessmentSnapshot struct {
	Date   string
	Tracks []ProgressTrack
}

type CareerStage struct {
	Code        string
	Title       string
	MinPercent  int
	MaxPercent  int
	Description string
	Focus       string
}

type CurrentCareerStage struct {
	Stage        CareerStage
	NextTitle    string
	PointsToNext int
	HasNext      bool
}

type Objective struct {
	Title      string
	Goal       string
	LessonSlug string
}

type JobSearchStage struct {
	MinPercent int
	MaxPercent int
	RangeLabel string
	Title      string
	Current    bool
}

type JourneyStatus string

const (
	JourneyComplete   JourneyStatus = "COMPLETE"
	JourneyInProgress JourneyStatus = "IN_PROGRESS"
	JourneyLocked     JourneyStatus = "LOCKED"
)

type AIJourneyStage struct {
	Number      int
	Title       string
	Status      JourneyStatus
	Description string
}

func (s AIJourneyStage) StatusLabel() string {
	switch s.Status {
	case JourneyComplete:
		return "COMPLETE ✓"
	case JourneyInProgress:
		return "IN PROGRESS →"
	default:
		return "LOCKED"
	}
}

func (s AIJourneyStage) StatusClass() string {
	switch s.Status {
	case JourneyComplete:
		return "complete"
	case JourneyInProgress:
		return "in-progress"
	default:
		return "locked"
	}
}

type Plan struct {
	Snapshot          AssessmentSnapshot
	CurrentStage      CurrentCareerStage
	Objective         Objective
	JobSearchRoadmap  []JobSearchStage
	AIJourney         []AIJourneyStage
	AIJourneyComplete int
	AIJourneyTotal    int
	AIJourneyCurrent  string
	AIJourneyNext     string
}

// CurrentPlan is the single Go-side source for the v1 assessment and roadmaps.
func CurrentPlan() Plan {
	snapshot := currentSnapshot()
	jobReadiness := trackPercent(snapshot.Tracks, TrackJob)
	journey := currentAIJourney()

	return Plan{
		Snapshot:          snapshot,
		CurrentStage:      CurrentStageFor(jobReadiness),
		Objective:         currentObjective(),
		JobSearchRoadmap:  currentJobSearchRoadmap(jobReadiness),
		AIJourney:         journey,
		AIJourneyComplete: countJourneyStatus(journey, JourneyComplete),
		AIJourneyTotal:    len(journey),
		AIJourneyCurrent:  "Replaceable ModelProvider",
		AIJourneyNext:     "Self-hosted model",
	}
}

func currentSnapshot() AssessmentSnapshot {
	return AssessmentSnapshot{
		Date: "2026-09-04",
		Tracks: []ProgressTrack{
			newTrack(
				TrackGo,
				"Go Engineer",
				18,
				"Сформирован начальный Go foundation: от запуска программы до условий, циклов и коллекций.",
				"Packages / Modules / структура Go-проекта",
				[]string{
					"программа", "компиляция", "variables", "types", "struct", "functions", "error",
					"if / else", "for", "slice", "append", "range", "map", "value, ok",
					"простые успешные и ошибочные сценарии",
				},
				[]string{
					"Packages / Modules", "Methods", "Pointers", "Interfaces", "net/http", "JSON / API",
					"Testing", "Context", "Concurrency", "Database", "Production Backend",
				},
			),
			newTrack(
				TrackPython,
				"Python Engineer",
				5,
				"Есть вводное понимание роли Python в автоматизации и AI/ML; системный track ещё впереди.",
				"Начать системный Python track",
				nil,
				[]string{
					"Python Core", "Functions / Collections", "Exceptions", "OOP / dataclasses", "Typing",
					"Iterators / generators", "AsyncIO", "pytest", "Backend", "AI/ML ecosystem",
				},
			),
			newTrack(
				TrackBackend,
				"Backend / Architecture",
				12,
				"Есть начальное понимание серверного проекта и разделения ответственности; границы компонентов нужно закрепить практикой.",
				"Понять package/module boundaries и структуру MentorForge",
				nil,
				[]string{
					"Project structure", "HTTP", "REST / JSON", "Service boundaries", "Storage", "PostgreSQL",
					"Testing", "Logging", "Configuration", "Deployment", "Scaling", "System Design",
				},
			),
			newTrack(
				TrackAI,
				"AI Engineer",
				6,
				"Освоены базовые понятия model, inference, prompt, embeddings и RAG; собственная AI-система ещё не построена.",
				"Разобрать self-hosted inference и open-weight model",
				nil,
				[]string{
					"Model basics", "Self-hosted inference", "Open-weight model", "vLLM", "Embeddings", "RAG",
					"Memory", "Tools", "Agent", "Evals", "LoRA / SFT", "Fine-tuning", "MentorForge Model",
				},
			),
			newTrack(
				TrackJob,
				"Готовность к работе",
				14,
				"Фундамент формируется, но для системных откликов пока не хватает устойчивой инженерной практики и проектов.",
				"Укрепить Go/Python/backend foundation и довести практические работы до проверяемого результата",
				nil,
				nil,
			),
		},
	}
}

func newTrack(id, title string, percent int, description, nextGoal string, evidence, roadmap []string) ProgressTrack {
	return ProgressTrack{
		ID:          id,
		Title:       title,
		Percent:     clampPercent(percent),
		Description: description,
		NextGoal:    nextGoal,
		Evidence:    evidence,
		Roadmap:     roadmap,
	}
}

func careerStages() []CareerStage {
	return []CareerStage{
		{Code: "Foundation", Title: "Фундамент", MinPercent: 0, MaxPercent: 19, Description: "Пока рано искать работу.", Focus: "Сейчас цель — получить устойчивый Go/Python/backend foundation."},
		{Code: "Apprentice Engineer", Title: "Ученик-инженер", MinPercent: 20, MaxPercent: 34, Description: "Базовые знания складываются в систему.", Focus: "Связывайте язык, проектную структуру и небольшие законченные задачи."},
		{Code: "Junior Track", Title: "Junior Track", MinPercent: 35, MaxPercent: 49, Description: "Начинаются полноценные инженерные задачи.", Focus: "Нужны проверяемые backend-задачи и уверенное объяснение решений."},
		{Code: "Job Search Preparation", Title: "Предстарт поиска", MinPercent: 50, MaxPercent: 59, Description: "Готовим проекты, резюме и разбираем вакансии.", Focus: "Закройте ключевые пробелы и подготовьте доказательства практических навыков."},
		{Code: "Ready for Applications", Title: "Готов к первым откликам", MinPercent: 60, MaxPercent: 74, Description: "Можно системно искать Junior / Junior+ позиции.", Focus: "Продолжайте учиться параллельно с первыми целевыми откликами."},
		{Code: "Working Engineer", Title: "Рабочий инженер", MinPercent: 75, MaxPercent: 89, Description: "Есть практический backend foundation.", Focus: "Углубляйте надёжность, эксплуатацию и архитектурные решения."},
		{Code: "AI / Backend Engineer", Title: "AI / Backend Engineer", MinPercent: 90, MaxPercent: 96, Description: "Можно проектировать более сложные системы.", Focus: "Развивайте system design и полный жизненный цикл AI/backend решений."},
		{Code: "AI Architect", Title: "AI Architect", MinPercent: 97, MaxPercent: 100, Description: "Строит собственную AI-систему.", Focus: "Соединяйте модели, данные, инструменты, evals и production infrastructure."},
	}
}

// CareerStageFor returns a stage for any value after constraining it to 0..100.
func CareerStageFor(percent int) CareerStage {
	percent = clampPercent(percent)
	stages := careerStages()
	for _, stage := range stages {
		if percent >= stage.MinPercent && percent <= stage.MaxPercent {
			return stage
		}
	}
	return stages[len(stages)-1]
}

func CurrentStageFor(percent int) CurrentCareerStage {
	percent = clampPercent(percent)
	stages := careerStages()
	for index, stage := range stages {
		if percent < stage.MinPercent || percent > stage.MaxPercent {
			continue
		}

		current := CurrentCareerStage{Stage: stage}
		if index+1 < len(stages) {
			current.HasNext = true
			current.NextTitle = stages[index+1].Title
			current.PointsToNext = stages[index+1].MinPercent - percent
		}
		return current
	}
	return CurrentCareerStage{Stage: stages[len(stages)-1]}
}

func currentObjective() Objective {
	return Objective{
		Title:      "Go: Packages, Modules и структура проекта",
		Goal:       "Понять, как большой Go-проект разбивается на packages, как работает go.mod и как компоненты зависят друг от друга.",
		LessonSlug: "go-lesson-04",
	}
}

func currentJobSearchRoadmap(percent int) []JobSearchStage {
	percent = clampPercent(percent)
	stages := []JobSearchStage{
		{MinPercent: 0, MaxPercent: 19, RangeLabel: "0–19", Title: "Фундамент"},
		{MinPercent: 20, MaxPercent: 34, RangeLabel: "20–34", Title: "Ученик-инженер"},
		{MinPercent: 35, MaxPercent: 49, RangeLabel: "35–49", Title: "Junior Track"},
		{MinPercent: 50, MaxPercent: 59, RangeLabel: "50–59", Title: "Предстарт поиска"},
		{MinPercent: 60, MaxPercent: 74, RangeLabel: "60–74", Title: "Первые отклики"},
		{MinPercent: 75, MaxPercent: 89, RangeLabel: "75–89", Title: "Активный поиск / инженерная работа"},
		{MinPercent: 90, MaxPercent: 100, RangeLabel: "90–100", Title: "Сильный инженерный профиль"},
	}
	for index := range stages {
		stages[index].Current = percent >= stages[index].MinPercent && percent <= stages[index].MaxPercent
	}
	return stages
}

func currentAIJourney() []AIJourneyStage {
	return []AIJourneyStage{
		{Number: 1, Title: "Learning Core", Status: JourneyComplete, Description: "Learning service, Lessons, Questions и PracticeTask существуют."},
		{Number: 2, Title: "AI Tutor architecture", Status: JourneyComplete, Description: "TutorService и явный tutor workflow реализованы."},
		{Number: 3, Title: "Replaceable ModelProvider", Status: JourneyComplete, Description: "TutorService зависит от интерфейса ModelProvider; облачный adapter заменяем."},
		{Number: 4, Title: "Self-hosted model", Status: JourneyLocked, Description: "Local/self-hosted provider в repository пока отсутствует."},
		{Number: 5, Title: "RAG / Knowledge", Status: JourneyLocked, Description: "Runtime knowledge retrieval пока не реализован."},
		{Number: 6, Title: "Memory", Status: JourneyLocked, Description: "Память AI Agent пока не реализована."},
		{Number: 7, Title: "Agent Tools", Status: JourneyLocked, Description: "Инструменты и agent loop пока не реализованы."},
		{Number: 8, Title: "Evals", Status: JourneyLocked, Description: "Набор AI evals пока не реализован."},
		{Number: 9, Title: "Fine-tuning", Status: JourneyLocked, Description: "LoRA / SFT pipeline пока не реализован."},
		{Number: 10, Title: "MentorForge Model v1", Status: JourneyLocked, Description: "Собственная модель пока не создана."},
		{Number: 11, Title: "Production Server", Status: JourneyLocked, Description: "Production deployment AI-системы пока не реализован."},
	}
}

func clampPercent(percent int) int {
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func trackPercent(tracks []ProgressTrack, id string) int {
	for _, track := range tracks {
		if track.ID == id {
			return clampPercent(track.Percent)
		}
	}
	return 0
}

func countJourneyStatus(stages []AIJourneyStage, status JourneyStatus) int {
	count := 0
	for _, stage := range stages {
		if stage.Status == status {
			count++
		}
	}
	return count
}
