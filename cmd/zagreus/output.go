package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/woodleighschool/zagreus/internal/app"
)

type syncReport struct {
	Cards   []app.Result  `json:"cards"`
	Summary reportSummary `json:"summary"`
	Error   string        `json:"error,omitempty"`
}

type reportSummary struct {
	Total    int                 `json:"total"`
	Actions  map[app.Action]int  `json:"actions"`
	Outcomes map[app.Outcome]int `json:"outcomes"`
}

func writeReport(writer io.Writer, jsonOutput, includeUnchanged, apply bool, results []app.Result, runErr error) error {
	if results == nil {
		return nil
	}
	report := syncReport{
		Cards:   make([]app.Result, 0, len(results)),
		Summary: reportSummary{Total: len(results), Actions: make(map[app.Action]int), Outcomes: make(map[app.Outcome]int)},
	}
	for _, result := range results {
		report.Summary.Actions[result.Action]++
		report.Summary.Outcomes[result.Outcome]++
		if jsonOutput || includeUnchanged || result.Outcome != app.OutcomeSkipped || result.Error != nil {
			report.Cards = append(report.Cards, result)
		}
	}
	if runErr != nil {
		report.Error = runErr.Error()
	}
	if jsonOutput {
		encoder := json.NewEncoder(writer)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(report)
	}

	style := newTextStyle(writer)
	var text strings.Builder
	heading := "Card sync plan"
	if apply {
		heading = "Card sync results"
	}
	fmt.Fprintf(&text, "%s\n\n", style.heading(heading))
	for _, result := range report.Cards {
		writeResult(&text, style, result, apply)
	}
	actions := report.Summary.Actions
	fmt.Fprintf(&text, "\n%s %d total, %d create, %d update, %d close, %d reopen, %d skip\n",
		style.paint("Cards:", color.Bold), report.Summary.Total, actions[app.ActionCreate], actions[app.ActionUpdate],
		actions[app.ActionClose], actions[app.ActionReopen], actions[app.ActionSkip])
	if apply {
		outcomes := report.Summary.Outcomes
		fmt.Fprintf(&text, "%s %d created, %d updated, %d closed, %d reopened, %d skipped, %d failed\n",
			style.paint("Cards:", color.Bold), outcomes[app.OutcomeCreated], outcomes[app.OutcomeUpdated],
			outcomes[app.OutcomeClosed], outcomes[app.OutcomeReopened], outcomes[app.OutcomeSkipped], outcomes[app.OutcomeFailed])
	} else {
		text.WriteString("No actions submitted. Only displaying plans.\n")
	}
	if runErr != nil {
		text.WriteString("Sync incomplete; reported outcomes remain valid\n")
	}
	_, err := io.WriteString(writer, text.String())
	return err
}

func writeResult(text *strings.Builder, style textStyle, result app.Result, apply bool) {
	var label, desc string
	if apply {
		switch result.Outcome {
		case app.OutcomeCreated:
			label, desc = "CREATED", "Created card on Trello board"
		case app.OutcomeUpdated:
			label, desc = "UPDATED", "Updated existing card on Trello board"
		case app.OutcomeClosed:
			label, desc = "CLOSED", "Closed card with all remediated hosts"
		case app.OutcomeReopened:
			label, desc = "REOPENED", "Reopened card with previously remediated hosts"
		case app.OutcomeSkipped:
			label, desc = "SKIPPED", "Skipped card with no changes to make"
		case app.OutcomeFailed:
			label, desc = "FAILED", fmt.Sprintf("Failed to enact change on card: %s", *result.Error)
		case app.OutcomePlan:
			panic(fmt.Errorf("impossible circumstance, apply = true and OutcomePlanned"))
		}
	} else {
		switch result.Action {
		case app.ActionCreate:
			label, desc = "CREATE", "Create card on Trello board"
		case app.ActionUpdate:
			label, desc = "UPDATE", "Update existing card on Trello board"
		case app.ActionClose:
			label, desc = "CLOSE", "Close card with all remediated hosts"
		case app.ActionReopen:
			label, desc = "REOPEN", "Reopen card with previously remediated hosts"
		case app.ActionSkip:
			label, desc = "SKIP", "Skip card with no changes to make"
		}
	}
	title := strconv.Quote(result.Title)
	fmt.Fprintf(text, "%s\n  %s: %s\n", style.heading(title), style.paint(strings.ToLower(label), titleColour(label)), style.paint(strings.ToLower(desc), color.Faint))
	if result.Vulnerability != nil {
		if result.Vulnerability.Labels != nil {
			labels := make([]string, 0, len(result.Vulnerability.Labels))
			for _, label := range result.Vulnerability.Labels {
				labels = append(labels, style.paint(strings.ToLower(label.String()), labelColour(label.String())))
			}
			fmt.Fprintf(text, "  Labels: [%s]\n", strings.Join(labels, ", "))
		}
		if result.Vulnerability.Hosts != nil {
			fmt.Fprintf(text, "  Hosts: %d\n", len(result.Vulnerability.Hosts))
		}
		if result.Vulnerability.CardID != nil {
			fmt.Fprintf(text, "  ID: %s\n", humanText(*result.Vulnerability.CardID))
		}
	}
	if result.Error != nil {
		fmt.Fprintf(text, "  %s %s\n", style.paint("✗", color.FgHiRed), humanText((*result.Error).Error()))
	}
	fmt.Fprintf(text, "\n")
}

func titleColour(title string) color.Attribute {
	switch title {
	case "CREATE", "CREATED", "UPDATE", "UPDATED", "CLOSE", "CLOSED":
		return color.FgHiGreen
	case "REOPEN", "REOPENED", "SKIP", "SKIPPED":
		return color.FgHiYellow
	case "FAILED":
		return color.FgHiRed
	}
	return color.Faint
}

func labelColour(label string) color.Attribute {
	switch label {
	case "info":
		return color.Faint
	case "low":
		return color.FgHiCyan
	case "medium":
		return color.FgHiYellow
	case "high":
		return color.FgHiRed
	case "critical":
		return color.FgHiMagenta
	case "resolved":
		return color.FgHiGreen
	case "waiting":
		return color.FgHiBlue
	}
	return color.Faint
}

func humanText(value string) string {
	var text strings.Builder
	for _, r := range value {
		if strconv.IsPrint(r) {
			text.WriteRune(r)
		} else {
			quoted := strconv.QuoteRune(r)
			text.WriteString(quoted[1 : len(quoted)-1])
		}
	}
	return text.String()
}
