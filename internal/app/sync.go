package app

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

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

func (s *Service) Plan(ctx context.Context, onlyChanges bool) ([]Intent, error) {
	var result []Intent

	s.logger.DebugContext(ctx, "setting up runtime vars")
	if err := s.init(ctx, false); err != nil {
		s.logger.ErrorContext(ctx, "failed to setup runtime", "err", err)
		return nil, err
	}
	s.logger.DebugContext(ctx, "finished runtime setup")

	s.logger.DebugContext(ctx, "preparing intents")
	intents, err := s.prepare(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to prepare intents", "err", err)
		return nil, err
	}

	for _, intent := range intents {
		if intent.Action != ActionSkip || !onlyChanges {
			result = append(result, intent)
		}
	}
	return result, nil
}

func (s *Service) FullSync(ctx context.Context) ([]Result, error) {
	var result []Result

	s.logger.DebugContext(ctx, "setting up runtime vars")
	if err := s.init(ctx, true); err != nil {
		s.logger.ErrorContext(ctx, "failed to setup runtime", "err", err)
		return nil, err
	}
	s.logger.DebugContext(ctx, "finished runtime setup")

	s.logger.DebugContext(ctx, "preparing intents")
	intents, err := s.prepare(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to prepare intents", "err", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "exporting prepared intents to trello")
	if err := s.exportChanges(ctx, intents, &result); err != nil {
		s.logger.ErrorContext(ctx, "failed to export changes to trello", "err", err)
		return nil, err
	}

	return result, nil
}

func (s *Service) Poll(ctx context.Context) ([]Result, error) {
	// TODO: Implementation
	var _ []Result

	if err := s.stateCheck(ctx, true); err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *Service) prepare(ctx context.Context) ([]Intent, error) {
	s.logger.DebugContext(ctx, "retrieving information from nessus")
	vulnerabilities, err := s.importNessusDetails(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve information from nessus", "err", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "retrieving information from trello")
	existing, err := s.importTrelloDetails(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve information from trello", "err", err)
		return nil, err
	}

	// TODO: Do I want stupid verbose logging?
	var intents []Intent
	for name, details := range vulnerabilities {
		var intent Intent
		card, cardOk := existing[name]
		if cardOk {
			hosts, equal := reconcileLists(details.Hosts, card.Hosts)
			states := slices.Collect(maps.Values(hosts))
			switch {
			case equal:
				intent.Action = ActionSkip
			case slices.Contains(states, false):
				if card.Closed {
					intent.Action = ActionReopen
					details.Labels = slices.DeleteFunc(details.Labels, func(l Severity) bool { return l == ResolvedSeverity })
				} else {
					intent.Action = ActionUpdate
				}
			default:
				intent.Action = ActionClose
			}
			intent.CardID = card.CardID
			intent.ChecklistID = card.ChecklistID
			intent.Title = details.Title
			intent.Description = details.Description
			intent.Labels = details.Labels
			intent.Hosts = hosts
		} else {
			hosts := make(map[string]bool)
			for host := range details.Hosts {
				hosts[host] = false
			}
			intent = Intent{
				Action:      ActionCreate,
				Title:       details.Title,
				Description: details.Description,
				Labels:      details.Labels,
				Hosts:       hosts,
			}
		}
		intents = append(intents, intent)
	}

	for name, details := range existing {
		_, ok := vulnerabilities[name]
		if !ok {
			intents = append(intents, Intent{
				Action:      ActionClose,
				CardID:      details.CardID,
				ChecklistID: details.ChecklistID,
				Title:       details.Title,
				Description: details.Description,
				Labels:      details.Labels,
				Hosts:       nil,
			})
		}
	}

	return intents, nil
}

func (s *Service) importNessusDetails(ctx context.Context) (map[string]Vulnerability, error) {
	s.logger.DebugContext(ctx, "getting host details for nessus scan", "scan", s.config.Nessus.Settings.Scan)
	hosts, err := s.nessus.GetScanDetails(ctx, s.config.Nessus.Settings.Scan)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get nessus scan details", "err", err)
		return nil, fmt.Errorf("unable to get nessus scans: %w", err)
	}
	vulnerabilities := make(map[string]Vulnerability)
	for _, host := range hosts {
		if s.shouldIgnoreHost(host) {
			continue
		}
		for _, vuln := range host.Vulnerabilities {
			if s.shouldIgnoreVuln(host, vuln) {
				continue
			}
			value, ok := vulnerabilities[vuln.PluginName]
			if ok {
				id := getIdentifier(host)
				value.Hosts[id] = false
			} else {
				payload, err := s.createPayload(ctx, vuln)
				if err != nil {
					s.logger.ErrorContext(ctx, "failed to prepare vulnerability information", "err", err)
					return nil, err
				}
				id := getIdentifier(host)
				payload.Hosts[id] = false
				value = payload
			}
			vulnerabilities[vuln.PluginName] = value
		}
	}

	return vulnerabilities, nil
}

func (s *Service) importTrelloDetails(ctx context.Context) (map[string]Vulnerability, error) {
	s.logger.DebugContext(ctx, "getting most recent trello board matching prefix", "prefix", s.config.Trello.Settings.Prefix)
	board, err := s.trello.GetLatestBoard(ctx, &s.config.Trello.Settings.Prefix)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve trello board", "err", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "listing card from trello board", "board", board.Name, "id", board.ID)
	boardCards, err := s.trello.ListBoardCards(ctx, board.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list trello board cards", "err", err)
		return nil, err
	}

	regex := regexp.MustCompile(".*zagreus_import.*")

	cards := make(map[string]Vulnerability)
	for _, resp := range boardCards {
		if !regex.MatchString(resp.Description) {
			s.logger.DebugContext(ctx, "skipping non-zagreus card", "name", resp.Name, "id", resp.ID)
			continue
		}
		if resp.Closed && s.config.Settings.Mode == "create" {
			s.logger.DebugContext(ctx, "skipping already closed card while in create mode", "name", resp.Name, "id", resp.ID)
			continue
		}
		card := Vulnerability{
			Title:       resp.Name,
			Description: resp.Description,
			CardID:      &resp.ID,
			Closed:      resp.Closed,
			Hosts:       make(map[string]bool),
		}
		if len(resp.ChecklistIDs) != 0 {
			s.logger.DebugContext(ctx, "getting checklists for trello card", "board", board.Name, "card", resp.ID)
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
	for _, intent := range intents {
		result := Result{
			Title:  intent.Title,
			Action: &intent.Action,
		}
		switch intent.Action {
		case ActionCreate:
			s.logger.DebugContext(ctx, "creating new card", "name", intent.Title)
			reason, err := s.createCard(ctx, intent)
			result.Reason = reason
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to create new card", "err", err)
				result.Outcome = OutcomeFailed
				result.Error = new(fmt.Errorf("%w", err))
			} else {
				result.Outcome = OutcomeCreated
			}
		case ActionUpdate:
			s.logger.DebugContext(ctx, "updating existing card", "name", intent.Title, "id", *intent.CardID)
			reason, err := s.updateCard(ctx, intent)
			result.Reason = reason
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to update card", "id", *intent.CardID, "err", err)
				result.Outcome = OutcomeFailed
				result.Error = new(fmt.Errorf("%w", err))
			} else {
				result.Outcome = OutcomeUpdated
			}
		case ActionReopen:
			s.logger.DebugContext(ctx, "reopening existing closed card", "name", intent.Title, "id", *intent.CardID)
			reason, err := s.reopenCard(ctx, intent)
			result.Reason = reason
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to reopen card", "id", *intent.CardID, "err", err)
				result.Outcome = OutcomeFailed
				result.Error = new(fmt.Errorf("%w", err))
			} else {
				result.Outcome = OutcomeReopened
			}
		case ActionClose:
			s.logger.DebugContext(ctx, "closing existing card with no active hosts", "name", intent.Title, "id", *intent.CardID)
			reason, err := s.closeCard(ctx, intent)
			result.Reason = reason
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to close card", "id", *intent.CardID, "err", err)
				result.Outcome = OutcomeFailed
				result.Error = new(fmt.Errorf("%w", err))
			} else {
				result.Outcome = OutcomeClosed
			}
		case ActionSkip:
			s.logger.DebugContext(ctx, "skipping card as no changes to make", "name", intent.Title, "id", *intent.CardID)
			result.Outcome = OutcomeSkipped
			result.Reason = "No changes to make"
		default:
			s.logger.ErrorContext(ctx, "unknown intent action", "action", intent.Action)
			return fmt.Errorf("unknown intent action: %s", intent.Action)
		}
		*results = append(*results, result)
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
