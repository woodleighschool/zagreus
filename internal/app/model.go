package app

type Result struct {
	Title  string `json:"title"`
	Action string `json:"action"`
	Reason string `json:"reason"`
	Error  *error `json:"error"`
}
