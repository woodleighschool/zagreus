package app

type Severity string

const (
	InfoSeverity     Severity = `info`
	LowSeverity      Severity = `low`
	MediumSeverity   Severity = `medium`
	HighSeverity     Severity = `high`
	CriticalSeverity Severity = `critical`
	ResolvedSeverity Severity = `resolved`
	WaitingSeverity  Severity = `waiting`
)

type Action string

const (
	ActionCreate Action = `create`
	ActionUpdate Action = `update`
	ActionClose  Action = `close`
	ActionSkip   Action = `skip`
)

type Outcome string

const (
	OutcomeCreated Outcome = `created`
	OutcomeUpdated Outcome = `updated`
	OutcomeClosed  Outcome = `closed`
	OutcomeSkipped Outcome = `skipped`
	OutcomeFailed  Outcome = `failed`
)

type Result struct {
	Title         string         `json:"title"`
	Action        *Action        `json:"action"`
	Outcome       Outcome        `json:"outcome"`
	Reason        string         `json:"reason"`
	Vulnerability *Vulnerability `json:"vulnerability"`
	Error         *error         `json:"error"`
}

type Vulnerability struct {
	Title       string
	Description string
	Hosts       map[string]bool
	Label       *Severity
	CardID      *string
	ChecklistID *string
}

type Intent struct {
	Action      Action  `json:"action"`
	CardID      *string `json:"cardId,omitempty"`
	ChecklistID *string `json:"checklistID,omitempty"`

	Title       string          `json:"title"`
	Description string          `json:"description"`
	Label       Severity        `json:"label"`
	Hosts       map[string]bool `json:"hosts"`
}
