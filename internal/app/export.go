package app

import (
	"context"
	"fmt"

	"github.com/woodleighschool/zagreus/external/trello"
)

func (s *Service) createBoard(ctx context.Context, boardID string, name string) (trello.BoardResponse, error) {
	board, err := s.trello.CopyBoard(ctx, trello.NewBoardRequest{
		Name:           name,
		SourceBoardID:  &boardID,
		KeepFromSource: new("cards"),
	})
	if err != nil {
		return trello.BoardResponse{}, err
	}
	lists, err := s.trello.ListBoardLists(ctx, board.ID)
	if err != nil {
		return trello.BoardResponse{}, err
	}
	var out string
	for _, list := range lists {
		if list.Name == s.config.Trello.Settings.Out {
			out = list.ID
		}
	}
	if out == "" {
		return trello.BoardResponse{}, fmt.Errorf("unable to find out list for cleanup")
	}
	if err := s.trello.ArchiveCardsInList(ctx, out); err != nil {
		return trello.BoardResponse{}, err
	}
	return board, nil
}

func (s *Service) createCard(ctx context.Context, intent Intent) error {
	labelIDs := make([]string, len(intent.Labels))
	for index, label := range intent.Labels {
		labelIDs[index] = s.runtime.Labels[label]
	}
	payload := trello.NewCardRequest{
		Name:        &intent.Title,
		Description: &intent.Description,
		LabelIDs:    labelIDs,
		Position:    new(trello.KeywordTop),
	}
	card, err := s.trello.CreateCard(ctx, s.runtime.InList, payload)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create card", "err", err)
		return fmt.Errorf("failed to create card: %w", err)
	}
	list, err := s.trello.CreateChecklist(ctx, card.ID, trello.NewChecklistRequest{
		Name: new("Hosts"),
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create checklist for card", "card", card.ID, "err", err)
		return fmt.Errorf("failed to create checklist on card: %w", err)
	}
	for host := range intent.Vulnerability.Hosts {
		err := s.trello.CreateCheckitem(ctx, list.ID, trello.NewCheckitemRequest{
			Name:    host,
			Checked: new(false),
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to create checkitem for card", "card", card.ID, "item", host, "err", err)
			return fmt.Errorf("failed to create checklist item for %s: %w", host, err)
		}
	}
	return nil
}

func (s *Service) updateCard(ctx context.Context, intent Intent) error {
	if intent.CardID == nil {
		return fmt.Errorf("missing card id for intent: %v", intent)
	}
	payload := trello.UpdateCardRequest{
		Name:        &intent.Title,
		Description: &intent.Description,
		Position:    new(trello.KeywordTop),
	}
	err := s.trello.UpdateCard(ctx, *intent.CardID, payload)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to update card", "err", err)
		return fmt.Errorf("failed to update card: %w", err)
	}
	if err := s.updateChecklist(ctx, intent); err != nil {
		return fmt.Errorf("failed to update card checklist: %w", err)
	}
	return nil
}

func (s *Service) reopenCard(ctx context.Context, intent Intent) error {
	if intent.CardID == nil {
		return fmt.Errorf("missing card id for intent: %v", intent)
	}
	labelIDs := make([]string, len(intent.Labels))
	for index, label := range intent.Labels {
		labelIDs[index] = s.runtime.Labels[label]
	}
	payload := trello.UpdateCardRequest{
		Name:        &intent.Title,
		Description: &intent.Description,
		Position:    new(trello.KeywordTop),
		Due:         &trello.DateTime{},
		DueComplete: new(false),
		ListID:      &s.runtime.InList,
		LabelIDs:    labelIDs,
	}
	err := s.trello.UpdateCard(ctx, *intent.CardID, payload)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to reopen card", "err", err)
		return fmt.Errorf("failed to reopen card: %w", err)
	}
	if err := s.updateChecklist(ctx, intent); err != nil {
		return fmt.Errorf("failed to update card checklist: %w", err)
	}
	return nil
}

func (s *Service) closeCard(ctx context.Context, intent Intent) error {
	if intent.ChecklistID == nil {
		return fmt.Errorf("missing checklist id for card to close: %v", intent)
	}
	if intent.CardID == nil {
		return fmt.Errorf("missing card ID to close: %v", intent)
	}
	items, err := s.trello.GetChecklistItems(ctx, *intent.ChecklistID)
	if err != nil {
		return fmt.Errorf("failed to get existing checklist items to close: %w", err)
	}
	for _, item := range items {
		if err := s.trello.UpdateCheckitem(ctx, *intent.CardID, item.ID, trello.UpdateCheckitemRequest{
			State: new(trello.CheckItemState(true)),
		}); err != nil {
			return fmt.Errorf("failed to close checkitem: %w", err)
		}
	}
	remediatedTime := trello.DateTime(s.now())
	labelIDs := make([]string, len(intent.Labels)+1)
	for index, label := range intent.Labels {
		labelIDs[index] = s.runtime.Labels[label]
	}
	labelIDs[len(labelIDs)-1] = s.runtime.Labels[ResolvedSeverity]
	payload := trello.UpdateCardRequest{
		Name:        &intent.Title,
		Description: &intent.Description,
		Position:    new(trello.KeywordTop),
		Due:         &remediatedTime,
		DueComplete: new(true),
		LabelIDs:    labelIDs,
		ListID:      &s.runtime.OutList,
	}
	err = s.trello.UpdateCard(ctx, *intent.CardID, payload)
	if err != nil {
		return fmt.Errorf("failed to close card: %w", err)
	}
	return nil
}

func (s *Service) updateChecklist(ctx context.Context, intent Intent) error {
	var listID string
	if intent.ChecklistID == nil {
		s.logger.WarnContext(ctx, "card is missing checklist", "card", *intent.CardID)
		list, err := s.trello.CreateChecklist(ctx, *intent.CardID, trello.NewChecklistRequest{
			Name: new("Hosts"),
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to create missing checklist on existing card", "card", *intent.CardID, "err", err)
			return fmt.Errorf("failed to create missing checklist on existing card: %w", err)
		}
		listID = list.ID
	} else {
		listID = *intent.ChecklistID
	}
	checklist := make(map[string]trello.CheckitemResponse)
	currentChecklist, err := s.trello.GetChecklistItems(ctx, listID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve existing checklist items", "card", *intent.CardID, "err", err)
		return fmt.Errorf("failed to retrieve existing checklist items: %w", err)
	}
	for _, item := range currentChecklist {
		checklist[item.Name] = item
	}
	for host, status := range intent.Vulnerability.Hosts {
		val, ok := checklist[host]
		switch {
		case !ok:
			err := s.trello.CreateCheckitem(ctx, listID, trello.NewCheckitemRequest{
				Name: host,
			})
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to create checklist item", "card", *intent.CardID, "item", host, "state", status, "err", err)
				return fmt.Errorf("failed to create checklist item: %w", err)
			}
		case ok && val.State.Value() != status:
			err := s.trello.UpdateCheckitem(ctx, *intent.CardID, val.ID, trello.UpdateCheckitemRequest{
				State: new(trello.CheckItemState(status)),
			})
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to update checklist item", "card", *intent.CardID, "item", host, "state", status, "err", err)
				return fmt.Errorf("failed to update checklist item: %w", err)
			}
		case ok && val.State.Value() == status:
			continue
		default:
			s.logger.ErrorContext(ctx, "unknown switch state", "incoming_state", status, "existing_state", val.State)
			return fmt.Errorf("unknown states - incoming state: %v, existing state: %v", status, val.State)
		}
	}
	return nil
}
