package domain

// Topic описывает учебную тему.
type Topic struct {
	ID          int    `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
