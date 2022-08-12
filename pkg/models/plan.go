package models

type Period struct {
	Begin     string `json:"start_timestamp"`
	End       string `json:"end_timestamp"`
	Helligdag bool   `json:"helligdag"`
}

type Plan struct {
	Ident    string              `json:"id"`
	Schedule map[string][]Period `json:"schedule"`
}
