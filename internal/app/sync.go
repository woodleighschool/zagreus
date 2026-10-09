package app

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/woodleighschool/zagreus/external/nessus"
	"github.com/woodleighschool/zagreus/external/trello"
	"github.com/woodleighschool/zagreus/internal/config"
)

type RuntimeSettings struct {
	CurrentTerm      config.Term
	DestinationBoard string
	InList           string
	OutList          string
	Labels           map[Severity]string
}

func (s *Service) init(ctx context.Context, write bool) error {
	s.logger.DebugContext(ctx, "checking current trello board state")
	if err := s.StateCheck(ctx, write); err != nil {
		s.logger.ErrorContext(ctx, "failed to check existing trello board state", "err", err)
		return err
	}

	s.logger.DebugContext(ctx, "setting current term")
	var currentTerm *config.Term
	for _, term := range s.config.Settings.Terms {
		if term.Range.In(s.now()) {
			s.logger.DebugContext(ctx, "found term", "term", term)
			currentTerm = &term
		}
	}
	if currentTerm == nil {
		s.logger.ErrorContext(ctx, "could not find any configured term for current date")
		return fmt.Errorf("app startup: %s is not in any configured term", s.now().Format("2006-01-02"))
	}
	s.runtime.CurrentTerm = *currentTerm

	s.logger.DebugContext(ctx, "setting destination board/lists")
	if err := s.setDestinations(ctx); err != nil {
		s.logger.ErrorContext(ctx, "failed to set destinations", "err", err)
		return err
	}

	return nil
}

func (s *Service) Plan(ctx context.Context) ([]Intent, error) {
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
		if intent.Action != ActionSkip {
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

	if err := s.StateCheck(ctx, true); err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *Service) StateCheck(ctx context.Context, write bool) error {
	config := s.config.Trello.Settings
	s.logger.DebugContext(ctx, "retrieving most recent board from with trello matching prefix", "prefix", config.Prefix)
	latestBoard, err := s.trello.GetLatestBoard(ctx, &config.Prefix)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve board", "err", err)
		return err
	}

	name := fmt.Sprintf("%s - %s - %d", config.Prefix, s.runtime.CurrentTerm.Name, s.now().Year())
	if latestBoard.Name != name {
		s.logger.DebugContext(ctx, "latest board is not for current term, copying to new board", "board", latestBoard.Name)
		return s.trello.CopyBoard(ctx, latestBoard.ID)
	}

	lists, err := s.trello.ListBoardLists(ctx, latestBoard.ID)
	missingLists := make(map[int]string)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve lists from latest board", "err", err)
		return err
	}
	for index, name := range config.Lists {
		if !slices.ContainsFunc(lists, func(list trello.BoardListResponse) bool {
			return list.Name == name
		}) {
			missingLists[index] = name
		}
	}
	s.logger.WarnContext(ctx, "missing lists on latest board", "lists", missingLists)

	missingLabels := make(map[string]string)
	for name, color := range config.Labels {
		if latestBoard.LabelNames[color] != name {
			missingLabels[name] = color
		}
	}
	s.logger.WarnContext(ctx, "missing labels on latest board", "labels", missingLabels)

	if write {
		s.logger.InfoContext(ctx, "creating missing lists", "lists", missingLists)
		var errs []error
		for index, name := range missingLists {
			pos := trello.KeywordFloat(index)
			if _, err := s.trello.CreateBoardList(ctx, latestBoard.ID, trello.NewBoardListRequest{Name: name, Position: &pos}); err != nil {
				s.logger.ErrorContext(ctx, "failed to create missing list", "list", name, "err", err)
				errs = append(errs, err)
			}
		}
		s.logger.InfoContext(ctx, "creating missing labels", "labels", missingLabels)
		for name, color := range missingLabels {
			if err := s.trello.CreateBoardLabel(ctx, latestBoard.ID, trello.NewBoardLabelRequest{Name: name, Color: color}); err != nil {
				s.logger.ErrorContext(ctx, "failed to create missing label", "label", name, "color", color, "err", err)
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}
	s.logger.WarnContext(ctx, "not in write mode, not actioning missing lists/labels")
	return fmt.Errorf("missing lists: %v; missing labels: %v", missingLists, missingLabels)
}

func (s *Service) setDestinations(ctx context.Context) error {
	s.logger.DebugContext(ctx, "setting runtime board")
	s.logger.DebugContext(ctx, "retrieving most recent board matching prefix", "prefix", s.config.Trello.Settings.Prefix)
	board, err := s.trello.GetLatestBoard(ctx, &s.config.Trello.Settings.Prefix)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve board", "err", err)
		return err
	}

	s.logger.DebugContext(ctx, "setting runtime lists")
	s.logger.DebugContext(ctx, "retrieving lists for current board", "board", board.Name)
	lists, err := s.trello.ListBoardLists(ctx, board.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve lists for board", "board", board.Name, "err", err)
		return err
	}
	for _, list := range lists {
		switch list.Name {
		case s.config.Trello.Settings.In:
			s.logger.DebugContext(ctx, "setting in list", "list", list.Name, "id", list.ID)
			s.runtime.InList = list.ID
		case s.config.Trello.Settings.Out:
			s.logger.DebugContext(ctx, "setting out list", "list", list.Name, "id", list.ID)
			s.runtime.OutList = list.ID
		default:
			continue
		}
	}
	if s.runtime.InList == "" || s.runtime.OutList == "" {
		s.logger.ErrorContext(ctx, "missing in and/or out list", "in_list", s.runtime.InList, "out_list", s.runtime.OutList)
		return fmt.Errorf("could not find in and/or out lists")
	}

	s.logger.DebugContext(ctx, "setting runtime labels")
	s.logger.DebugContext(ctx, "retrieving labels for current board", "board", board.Name)
	labels, err := s.trello.ListBoardLabels(ctx, board.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve labels for board", "board", board.Name, "err", err)
	}
	s.runtime.Labels = make(map[Severity]string)
	for _, label := range labels {
		switch label.Name {
		case "Low":
			s.logger.DebugContext(ctx, "setting low severity label", "id", label.ID, "color", label.Color)
			s.runtime.Labels[LowSeverity] = label.ID
		case "Medium":
			s.logger.DebugContext(ctx, "setting medium severity label", "id", label.ID, "color", label.Color)
			s.runtime.Labels[MediumSeverity] = label.ID
		case "High":
			s.logger.DebugContext(ctx, "setting high severity label", "id", label.ID, "color", label.Color)
			s.runtime.Labels[HighSeverity] = label.ID
		case "Critical":
			s.logger.DebugContext(ctx, "setting critical severity label", "id", label.ID, "color", label.Color)
			s.runtime.Labels[CriticalSeverity] = label.ID
		case "Resolved":
			s.logger.DebugContext(ctx, "setting resolved severity label", "id", label.ID, "color", label.Color)
			s.runtime.Labels[ResolvedSeverity] = label.ID
		case "Waiting":
			s.logger.DebugContext(ctx, "setting waiting severity label", "id", label.ID, "color", label.Color)
			s.runtime.Labels[WaitingSeverity] = label.ID
		default:
			continue
		}
	}

	return nil
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
				intent.Action = ActionUpdate
			default:
				intent.Action = ActionClose
			}
			intent.CardID = card.CardID
			intent.ChecklistID = card.ChecklistID
			intent.Title = details.Title
			intent.Description = details.Description
			intent.Label = *details.Label
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
				Label:       *details.Label,
				Hosts:       hosts,
			}
		}
		intents = append(intents, intent)
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
			if vuln.PluginID == 42873 {
				s.logger.DebugContext(ctx, "party time")
			}
			value, ok := vulnerabilities[vuln.PluginName]
			if ok {
				if host.Info.FQDN == "" {
					value.Hosts[host.Info.IP] = false
				} else {
					value.Hosts[host.Info.FQDN] = false
				}
			} else {
				payload, err := s.createPayload(ctx, vuln)
				if err != nil {
					s.logger.ErrorContext(ctx, "failed to prepare vulnerability information", "err", err)
					return nil, err
				}
				if host.Info.FQDN == "" {
					payload.Hosts[host.Info.IP] = false
				} else {
					payload.Hosts[host.Info.FQDN] = false
				}
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

	cards := make(map[string]Vulnerability)
	for _, resp := range boardCards {
		card := Vulnerability{
			Title:       resp.Name,
			Description: resp.Description,
			CardID:      &resp.ID,
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
		result.Label = new(LowSeverity)
	case nessus.MediumSeverity:
		result.Label = new(MediumSeverity)
	case nessus.HighSeverity:
		result.Label = new(HighSeverity)
	case nessus.CriticalSeverity:
		result.Label = new(CriticalSeverity)
	case nessus.InfoSeverity:
		result.Label = new(InfoSeverity)
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
	
	**External:** [Report details](%s/scans/reports/%d/vulnerabilities/%d "")`,
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

func (s *Service) createCard(ctx context.Context, intent Intent) (string, error) {
	payload := trello.NewCardRequest{
		Name:        &intent.Title,
		Description: &intent.Description,
		LabelIDs:    []string{s.runtime.Labels[intent.Label]},
		Position:    new(trello.KeywordTop),
	}
	card, err := s.trello.CreateCard(ctx, s.runtime.InList, payload)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create card", "err", err)
		return "Failed to create card", err
	}
	list, err := s.trello.CreateChecklist(ctx, card.ID, trello.NewChecklistRequest{
		Name: new("Hosts"),
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create checklist for card", "card", card.ID, "err", err)
		return "Unable to create checklist", err
	}
	for host := range intent.Hosts {
		err := s.trello.CreateCheckitem(ctx, list.ID, trello.NewCheckitemRequest{
			Name:    host,
			Checked: new(false),
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to create checkitem for card", "card", card.ID, "item", host, "err", err)
			return fmt.Sprintf("Failed to create checklist item for %s", host), err
		}
	}
	return "Vulnerability was not on board", nil
}

func (s *Service) updateCard(ctx context.Context, intent Intent) (string, error) {
	if intent.CardID == nil {
		return "Missing card ID?", fmt.Errorf("missing card id for intent: %v", intent)
	}
	payload := trello.UpdateCardRequest{
		Name:        &intent.Title,
		Description: &intent.Description,
		Position:    new(trello.KeywordTop),
	}
	err := s.trello.UpdateCard(ctx, *intent.CardID, payload)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to update card", "err", err)
		return "Failed to update card", err
	}
	var listID string
	if intent.ChecklistID == nil {
		s.logger.WarnContext(ctx, "card is missing checklist", "card", *intent.CardID)
		list, err := s.trello.CreateChecklist(ctx, *intent.CardID, trello.NewChecklistRequest{
			Name: new("Hosts"),
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to create missing checklist on existing card", "card", *intent.CardID, "err", err)
			return "Unable to create missing checklist on existing card", err
		}
		listID = list.ID
	} else {
		listID = *intent.ChecklistID
	}
	checklist := make(map[string]trello.CheckitemResponse)
	currentChecklist, err := s.trello.GetChecklistItems(ctx, listID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve existing checklist items", "card", *intent.CardID, "err", err)
		return "Unable to retrieve existing checklist items", err
	}
	for _, item := range currentChecklist {
		checklist[item.Name] = item
	}
	for host, status := range intent.Hosts {
		val, ok := checklist[host]
		switch {
		case !ok:
			err := s.trello.CreateCheckitem(ctx, listID, trello.NewCheckitemRequest{
				Name: host,
			})
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to create checklist item", "card", *intent.CardID, "item", host, "state", status, "err", err)
				return "Unable to create checklist item", err
			}
		case ok && val.State.Value() != status:
			err := s.trello.UpdateCheckitem(ctx, *intent.CardID, val.ID, trello.UpdateCheckitemRequest{
				State: new(trello.CheckItemState(status)),
			})
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to update checklist item", "card", *intent.CardID, "item", host, "state", status, "err", err)
				return "Unable to update checklist item", err
			}
		case ok && val.State.Value() == status:
			continue
		default:
			s.logger.ErrorContext(ctx, "unknown switch state", "incoming_state", status, "existing_state", val.State)
			return "Checkitem switch state unknown", fmt.Errorf("incoming state: %v, existing state: %v", status, val.State)
		}
	}
	return "Details updated", nil
}

func (s *Service) closeCard(ctx context.Context, intent Intent) (string, error) {
	if intent.ChecklistID == nil {
		return "Missing checklist ID?", fmt.Errorf("missing checklist id for card to close: %v", intent)
	}
	if intent.CardID == nil {
		return "Missing card ID?", fmt.Errorf("missing card ID to close: %v", intent)
	}
	items, err := s.trello.GetChecklistItems(ctx, *intent.ChecklistID)
	if err != nil {
		return "Failed to get existing checklist items to close", err
	}
	for _, item := range items {
		if err := s.trello.UpdateCheckitem(ctx, *intent.CardID, item.ID, trello.UpdateCheckitemRequest{
			State: new(trello.CheckItemState(false)),
		}); err != nil {
			return "Unable to close checkitem", err
		}
	}
	payload := trello.UpdateCardRequest{
		Name:        &intent.Title,
		Description: &intent.Description,
		Position:    new(trello.KeywordTop),
		Closed:      new(true),
	}
	err = s.trello.UpdateCard(ctx, *intent.CardID, payload)
	if err != nil {
		return "Failed to close card", err
	}
	return "All hosts remediated", nil
}

func (s *Service) shouldIgnoreHost(host nessus.HostDetails) bool {
	// Check config for a host entry
	hostConfig, ok := s.config.Nessus.Settings.Ignore.Hosts[host.Info.FQDN]
	if ok {
		if hostConfig.Full != nil && *hostConfig.Full {
			return true
		}
	}
	return false
}

func (s *Service) shouldIgnoreVuln(host nessus.HostDetails, vuln nessus.HostVulnerability) bool {
	if slices.Contains(s.config.Nessus.Settings.Ignore.Severity, vuln.Severity.String()) || slices.Contains(s.config.Nessus.Settings.Ignore.Plugins, vuln.PluginID) {
		return true
	}

	hostConfig, ok := s.config.Nessus.Settings.Ignore.Hosts[host.Info.FQDN]
	if !ok {
		return false
	}
	if slices.Contains(hostConfig.Plugins, vuln.PluginID) || slices.Contains(hostConfig.Severity, vuln.Severity.String()) {
		return true
	}
	return false
}

func reconcileLists(nessusList map[string]bool, trelloList map[string]bool) (map[string]bool, bool) {
	result := make(map[string]bool)
	maps.Copy(result, trelloList)

	// Import new items from Nessus
	for host := range nessusList {
		_, ok := result[host]
		if !ok {
			result[host] = false
		}
	}

	// Mark any items in Trello not in Nessus as resolved
	for host := range trelloList {
		_, ok := nessusList[host]
		if !ok {
			result[host] = true
		}
	}

	// Check for equality with Trello
	equal := true
	if len(result) != len(trelloList) {
		equal = false
	}
	for host, value := range result {
		val, ok := trelloList[host]
		if !ok || val != value {
			equal = false
		}
	}

	return result, equal
}
