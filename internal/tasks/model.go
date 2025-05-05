package tasks

type Task struct {
	Title              string   `json:"title"`
	Body               string   `json:"body"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Labels             []string `json:"labels"`
}
