package model

type Argument struct {
	Name    string `json:"name"`
	Example string `json:"example"`
	Meaning string `json:"meaning"`
}

type Candidate struct {
	Command     string     `json:"command"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Args        []Argument `json:"args"`
}
