package mcla

type ErrorDesc struct {
	Error     string         `json:"error"`
	Message   string         `json:"message"`
	Solutions []int          `json:"solutions"`
	Data      map[string]any `json:"data,omitempty"`
}

type SolutionDesc struct {
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	LinkTo      string   `json:"link_to"`
}

type ErrorDB interface {
	ForEachErrors(callback func(*ErrorDesc) error) (err error)
	GetSolution(id int) (sol *SolutionDesc, err error)
}
