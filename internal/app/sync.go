package app

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/woodleighschool/zagreus/external/nessus"
	"github.com/woodleighschool/zagreus/internal/config"
)

type RuntimeSettings struct {
	CurrentTerm      config.Term
	DestinationBoard string
	InList           string
	OutList          string
	Labels           map[Severity]string
}

func (s *Service) Plan(ctx context.Context, includeUnchanged bool) ([]Result, error) {
	var result []Result
	var err error
	step := s.stage(ctx, "Setting up runtime variables")
	defer func() { step(err) }()
	if err := s.init(ctx, false); err != nil {
		step(fmt.Errorf("failed to setup runtime: %w", err))
		return nil, err
	}
	step(nil, "detail", "runtime setup successfully")

	step = s.stage(ctx, "Getting Nessus details")
	vulns, err := s.importNessusDetails(ctx)
	if err != nil {
		step(fmt.Errorf("failed to import nessus details: %w", err))
		return nil, err
	}
	step(nil, "detail", fmt.Sprintf("%d vulnerabilities", len(vulns)))

	step = s.stage(ctx, "Getting existing Trello cards")
	cards, err := s.importTrelloDetails(ctx)
	if err != nil {
		step(fmt.Errorf("failed to import details from trello: %w", err))
		return nil, err
	}
	step(nil, "detail", fmt.Sprintf("%d cards", len(cards)))

	step = s.stage(ctx, "Preparing intents")
	intents := s.prepare(ctx, vulns, cards)
	if err != nil {
		step(fmt.Errorf("failed to prepare intents: %w", err))
		return nil, err
	}

	for _, intent := range intents {
		if intent.Action != ActionSkip || includeUnchanged {
			result = append(result, Result{
				Intent:  intent,
				Outcome: OutcomePlan,
			})
		}
	}
	step(nil, "detail", fmt.Sprintf("%d intents", len(intents)))
	return result, nil
}

func (s *Service) Sync(ctx context.Context) ([]Result, error) {
	var result []Result
	var err error
	step := s.stage(ctx, "Setting up runtime variables")
	defer func() { step(err) }()
	if err := s.init(ctx, false); err != nil {
		step(fmt.Errorf("failed to setup runtime: %w", err))
		return nil, err
	}
	step(nil, "detail", "runtime setup successfully")

	step = s.stage(ctx, "Getting Nessus details")
	vulns, err := s.importNessusDetails(ctx)
	if err != nil {
		step(fmt.Errorf("failed to import nessus details: %w", err))
		return nil, err
	}
	step(nil, "detail", fmt.Sprintf("%d vulnerabilities", len(vulns)))

	step = s.stage(ctx, "Getting existing Trello cards")
	cards, err := s.importTrelloDetails(ctx)
	if err != nil {
		step(fmt.Errorf("failed to import details from trello: %w", err))
		return nil, err
	}
	step(nil, "detail", fmt.Sprintf("%d cards", len(cards)))

	step = s.stage(ctx, "Preparing intents")
	intents := s.prepare(ctx, vulns, cards)
	if err != nil {
		step(fmt.Errorf("failed to prepare intents: %w", err))
		return nil, err
	}
	step(nil, "detail", fmt.Sprintf("%d intents", len(intents)))

	step = s.stage(ctx, "Exporting prepared intents to Trello")
	if err := s.exportChanges(ctx, intents, &result); err != nil {
		step(fmt.Errorf("failed to export changes to trello: %w", err))
		return nil, err
	}
	step(nil, "detail", "changes exported successfully")
	return result, nil
}

func (s *Service) Validate(ctx context.Context) error {
	return s.init(ctx, false)
}

func (s *Service) Poll(ctx context.Context) ([]Result, error) {
	// TODO: Implementation
	var _ []Result

	if err := s.stateCheck(ctx, true); err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *Service) prepare(ctx context.Context, vulns map[string]Vulnerability, cards map[string]Vulnerability) []Intent {
	var intents []Intent
	progress := 0
	for name, details := range vulns {
		var intent Intent
		card, cardOk := cards[name]
		if cardOk {
			hosts, equal := reconcileLists(details.Hosts, card.Hosts)
			states := slices.Collect(maps.Values(hosts))
			switch {
			case equal:
				intent.Action = ActionSkip
				intent.Reason = "No changes required"
			case card.Closed:
				if slices.Contains(states, false) {
					intent.Action = ActionReopen
					intent.Reason = "Hosts still reporting from Nessus"
					details.Labels = reconcileLabels(details.Labels, card.Labels, ResolvedSeverity)
				} else {
					intent.Action = ActionSkip
					intent.Reason = "No changes required"
				}
			case !slices.Contains(states, false):
				intent.Action = ActionClose
				intent.Reason = "No more hosts reporting from Nessus"
			default:
				intent.Action = ActionUpdate
			}
			intent.CardID = card.CardID
			intent.ChecklistID = card.ChecklistID
			intent.Title = details.Title
			intent.Description = details.Description
			intent.Labels = reconcileLabels(details.Labels, card.Labels)
			intent.Vulnerability = &details
			delete(cards, name)
		} else {
			hosts := make(map[string]bool)
			for host := range details.Hosts {
				hosts[host] = false
			}
			intent = Intent{
				Action:        ActionCreate,
				Title:         details.Title,
				Description:   details.Description,
				Labels:        details.Labels,
				Vulnerability: &details,
				Reason:        "Card does not currently exist on board",
			}
		}
		intents = append(intents, intent)

		progress++
		s.logger.InfoContext(ctx, "Parsing vulnerabilities from Nessus", "progress", true, "current", progress, "total", len(vulns), "unit", "vulns", "progress_final", progress == len(vulns))
	}

	progress = 0
	for name, details := range cards {
		_, ok := vulns[name]
		if !ok && !details.Closed {
			intents = append(intents, Intent{
				Action:      ActionClose,
				Title:       details.Title,
				Description: details.Description,
				Labels:      details.Labels,
				Reason:      "No matching vulnerability from Nessus",
				CardID:      details.CardID,
				ChecklistID: details.ChecklistID,
			})
		} else {
			intents = append(intents, Intent{
				Action:      ActionSkip,
				Title:       details.Title,
				Description: details.Description,
				Labels:      details.Labels,
				Reason:      "Nothing to action on card",
				CardID:      details.CardID,
				ChecklistID: details.ChecklistID,
			})
		}

		progress++
		s.logger.InfoContext(ctx, "Parsing orphaned vulnerabilities on Trello", "progress", true, "current", progress, "total", len(cards), "unit", "vulns", "progress_final", progress == len(cards))
	}

	return intents
}

func (s *Service) importNessusDetails(ctx context.Context) (map[string]Vulnerability, error) {
	result := make(map[string]Vulnerability)
	s.logger.DebugContext(ctx, "getting host details for nessus scan", "scan", s.config.Nessus.Settings.Scan)
	scan, err := s.nessus.GetScan(ctx, s.config.Nessus.Settings.Scan)
	if err != nil {
		return nil, fmt.Errorf("unable to get nessus scan: %w", err)
	}
	progress := 0
	for _, host := range scan.Hosts {
		details, err := s.nessus.GetScanHostDetails(ctx, scan.Info.ObjectID, host.ID)
		if err != nil {
			return nil, fmt.Errorf("unable to get scan details for host %d: %w", host.ID, err)
		}
		if s.shouldIgnoreHost(details) {
			progress++
			s.logger.InfoContext(ctx, "Getting Nessus details", "progress", true, "current", progress, "total", len(scan.Hosts), "unit", "hosts", "progress_final", progress == len(scan.Hosts))
		}
		for _, vuln := range details.Vulnerabilities {
			if s.shouldIgnoreVuln(details, vuln) {
				continue
			}
			val, ok := result[vuln.PluginName]
			if ok {
				id := getIdentifier(details)
				val.Hosts[id] = false
			} else {
				payload, err := s.createPayload(ctx, vuln)
				if err != nil {
					s.logger.ErrorContext(ctx, "failed to prepare vulnerability information", "err", err)
					return nil, err
				}
				id := getIdentifier(details)
				payload.Hosts[id] = false
				val = payload
			}
			result[vuln.PluginName] = val
		}

		progress++
		s.logger.InfoContext(ctx, "Getting Nessus details", "progress", true, "current", progress, "total", len(scan.Hosts), "unit", "hosts", "progress_final", progress == len(scan.Hosts))
	}

	return result, nil
}

func (s *Service) importTrelloDetails(ctx context.Context) (map[string]Vulnerability, error) {
	board, err := s.trello.GetLatestBoard(ctx, &s.config.Trello.Settings.Prefix)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve trello board", "err", err)
		return nil, err
	}

	boardCards, err := s.trello.ListBoardCards(ctx, board.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list trello board cards", "err", err)
		return nil, err
	}

	regex := regexp.MustCompile(".*zagreus_import.*")

	cards := make(map[string]Vulnerability)
	for progress := 0; progress < len(boardCards); s.updateProgress(ctx, "Getting existing Trello cards", &progress, len(boardCards), "cards") {
		resp := boardCards[progress]
		if !regex.MatchString(resp.Description) {
			s.logger.DebugContext(ctx, "skipping non-zagreus card", "name", resp.Name, "id", resp.ID, "verbose", true)
			continue
		}
		if resp.Closed && s.config.Settings.Mode == "create" {
			s.logger.DebugContext(ctx, "skipping already closed card while in create mode", "name", resp.Name, "id", resp.ID, "verbose", true)
			continue
		}
		card := Vulnerability{
			Title:       resp.Name,
			Description: resp.Description,
			CardID:      &resp.ID,
			Closed:      resp.DueComplete,
			Hosts:       make(map[string]bool),
		}
		if len(resp.ChecklistIDs) != 0 {
			s.logger.DebugContext(ctx, "getting checklists for trello card", "board", board.Name, "card", resp.ID, "verbose", true)
			checklists, err := s.trello.GetChecklists(ctx, resp.ChecklistIDs)
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to get checklists for trello card", "card", resp.Name, "id", resp.ID)
				return nil, err
			}
			for _, checklist := range checklists {
				if checklist.Name != "Hosts" {
					continue
				}
				card.ChecklistID = &checklist.ID
				for _, item := range checklist.Items {
					card.Hosts[item.Name] = item.State.Value()
				}
			}
		}
		for name, id := range s.runtime.Labels {
			if slices.Contains(resp.LabelIDs, id) {
				card.Labels = append(card.Labels, name)
			}
		}
		cards[card.Title] = card
	}
	return cards, nil
}

func (s *Service) exportChanges(ctx context.Context, intents []Intent, results *[]Result) error {
	progress := 0
	for _, intent := range intents {
		result := Result{
			Intent: intent,
		}
		switch intent.Action {
		case ActionCreate:
			s.logger.DebugContext(ctx, "creating new card", "name", intent.Title, "verbose", true)
			if err := s.createCard(ctx, intent); err != nil {
				s.logger.ErrorContext(ctx, "failed to create new card", "err", err)
				result.Outcome = OutcomeFailed
				result.Error = new(fmt.Errorf("%w", err))
			} else {
				result.Outcome = OutcomeCreated
			}
		case ActionUpdate:
			s.logger.DebugContext(ctx, "updating existing card", "name", intent.Title, "id", *intent.CardID, "verbose", true)
			if err := s.updateCard(ctx, intent); err != nil {
				s.logger.ErrorContext(ctx, "failed to update card", "id", *intent.CardID, "err", err)
				result.Outcome = OutcomeFailed
				result.Error = new(fmt.Errorf("%w", err))
			} else {
				result.Outcome = OutcomeUpdated
			}
		case ActionReopen:
			s.logger.DebugContext(ctx, "reopening existing closed card", "name", intent.Title, "id", *intent.CardID, "verbose", true)
			if err := s.reopenCard(ctx, intent); err != nil {
				s.logger.ErrorContext(ctx, "failed to reopen card", "id", *intent.CardID, "err", err)
				result.Outcome = OutcomeFailed
				result.Error = new(fmt.Errorf("%w", err))
			} else {
				result.Outcome = OutcomeReopened
			}
		case ActionClose:
			s.logger.DebugContext(ctx, "closing existing card with no active hosts", "name", intent.Title, "id", *intent.CardID, "verbose", true)
			if err := s.closeCard(ctx, intent); err != nil {
				s.logger.ErrorContext(ctx, "failed to close card", "id", *intent.CardID, "err", err)
				result.Outcome = OutcomeFailed
				result.Error = new(fmt.Errorf("%w", err))
			} else {
				result.Outcome = OutcomeClosed
			}
		case ActionSkip:
			s.logger.DebugContext(ctx, "skipping card as no changes to make", "name", intent.Title, "id", *intent.CardID, "verbose", true)
			result.Outcome = OutcomeSkipped
		default:
			s.logger.ErrorContext(ctx, "unknown intent action", "action", intent.Action)
			return fmt.Errorf("unknown intent action: %s", intent.Action)
		}
		*results = append(*results, result)

		progress++
		s.logger.InfoContext(ctx, "Exporting prepared intents to Trello", "progress", true, "current", progress, "total", len(intents), "unit", "intents", "progress_final", progress == len(intents))
	}
	return nil
}

func (s *Service) createPayload(ctx context.Context, vuln nessus.HostVulnerability) (Vulnerability, error) {
	var result Vulnerability
	vulnDetails, err := s.nessus.GetPluginDetails(ctx, vuln.PluginID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get plugin details", "err", err)
		return Vulnerability{}, err
	}
	result.Title = vulnDetails.Name

	cvssScore := "unknown"
	cvssVector := "unknown"
	vulnAge := "unknown"

	switch vuln.Severity {
	case nessus.LowSeverity:
		result.Labels = []Severity{LowSeverity}
	case nessus.MediumSeverity:
		result.Labels = []Severity{MediumSeverity}
	case nessus.HighSeverity:
		result.Labels = []Severity{HighSeverity}
	case nessus.CriticalSeverity:
		result.Labels = []Severity{CriticalSeverity}
	case nessus.InfoSeverity:
		result.Labels = []Severity{InfoSeverity}
	}

	if vulnDetails.CVSS3BaseScore != nil {
		cvssScore = *vulnDetails.CVSS3BaseScore
	}

	if vulnDetails.CVSS3Vector != nil {
		cvssVector = *vulnDetails.CVSS3Vector
	}

	if vulnDetails.VulnAge != nil {
		vulnAge = *vulnDetails.VulnAge
	}

	description := fmt.Sprintf(`**CVE Score:** %s
	**CVE Vector:** %s
	**Age:** %s
	
	**External:** [Report details](%s/scans/reports/%d/vulnerabilities/%d "")

	_zagreus_import_`,
		cvssScore,
		cvssVector,
		vulnAge,
		s.config.Nessus.Connection.Host,
		s.config.Nessus.Settings.Scan,
		vuln.PluginID)

	result.Description = strings.ReplaceAll(description, "\t", "")

	result.Hosts = make(map[string]bool)

	return result, nil
}

func (s *Service) updateProgress(ctx context.Context, message string, progress *int, total int, unit string) {
	*progress++
	s.logger.InfoContext(ctx, message, "progress", true, "current", *progress, "total", total, "unit", unit, "progress_final", *progress == total)
}

func (s *Service) stage(ctx context.Context, message string) func(error, ...any) {
	started := time.Now()
	s.logger.InfoContext(ctx, message, "stage", true)
	var once sync.Once
	return func(err error, details ...any) {
		once.Do(func() {
			result := append([]any{"stage_result", true, "elapsed", time.Since(started).Round(time.Millisecond)}, details...)
			if err != nil {
				result = append(result, "error", err)
			}
			s.logger.InfoContext(ctx, message, result...)
		})
	}
}
