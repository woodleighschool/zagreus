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

func (s Severity) String() string {
	return string(s)
}

type Action string

const (
	ActionCreate Action = `create`
	ActionUpdate Action = `update`
	ActionClose  Action = `close`
	ActionSkip   Action = `skip`
	ActionReopen Action = `reopen`
)

type Outcome string

const (
	OutcomeCreated  Outcome = `created`
	OutcomeUpdated  Outcome = `updated`
	OutcomeReopened Outcome = `reopened`
	OutcomeClosed   Outcome = `closed`
	OutcomeSkipped  Outcome = `skipped`
	OutcomeFailed   Outcome = `failed`
	OutcomePlan     Outcome = `plan`
)

type Result struct {
	Intent

	Outcome Outcome `json:"outcome"`
	Error   *error  `json:"error"`
}

type Intent struct {
	Action Action `json:"action"`

	Title       string     `json:"title"`
	Description string     `json:"description"`
	Labels      []Severity `json:"label"`
	Reason      string     `json:"reason"`

	Vulnerability *Vulnerability `json:"vulnerability"`

	CardID      *string `json:"cardId,omitempty"`
	ChecklistID *string `json:"checklistID,omitempty"`
}

type Vulnerability struct {
	Title       string          `json:"title"`
	Description string          `json:"desc"`
	Hosts       map[string]bool `json:"hosts"`
	Labels      []Severity      `json:"labels"`
	Closed      bool            `json:"closed"`
	CardID      *string         `json:"cardID"`
	ChecklistID *string         `json:"checklistID"`
}
