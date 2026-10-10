package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/woodleighschool/zagreus/external/trello"
	"github.com/woodleighschool/zagreus/internal/config"
)

func (s *Service) init(ctx context.Context, write bool) error {
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

	s.logger.DebugContext(ctx, "checking current trello board state")
	if err := s.stateCheck(ctx, write); err != nil {
		s.logger.ErrorContext(ctx, "failed to check existing trello board state", "err", err)
		return err
	}

	s.logger.DebugContext(ctx, "setting destination board/lists")
	if err := s.setDestinations(ctx); err != nil {
		s.logger.ErrorContext(ctx, "failed to set destinations", "err", err)
		return err
	}

	return nil
}

func (s *Service) setDestinations(ctx context.Context) error {
	board, err := s.setBoard(ctx)
	if err != nil {
		return err
	}

	if err := s.setLabels(ctx, board); err != nil {
		return err
	}
	return nil
}

func (s *Service) stateCheck(ctx context.Context, write bool) error {
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
		board, err := s.createBoard(ctx, latestBoard.ID, name)
		if err != nil {
			return err
		}
		if err := s.trello.UpdateBoard(ctx, latestBoard.ID, trello.UpdateBoardRequest{Closed: new(true)}); err != nil {
			return err
		}
		latestBoard = board
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
	if len(missingLists) > 0 {
		s.logger.WarnContext(ctx, "missing lists on latest board", "lists", missingLists)
	}

	missingLabels := make(map[string]string)
	for name, color := range config.Labels {
		if strings.ToLower(latestBoard.LabelNames[color]) != name {
			missingLabels[name] = color
		}
	}
	if len(missingLabels) > 0 {
		s.logger.WarnContext(ctx, "missing labels on latest board", "labels", missingLabels)
	}

	if len(missingLists) == 0 && len(missingLabels) == 0 {
		return nil
	}
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

func (s *Service) setBoard(ctx context.Context) (trello.BoardResponse, error) {
	s.logger.DebugContext(ctx, "setting runtime board")
	s.logger.DebugContext(ctx, "retrieving most recent board matching prefix", "prefix", s.config.Trello.Settings.Prefix)
	board, err := s.trello.GetLatestBoard(ctx, &s.config.Trello.Settings.Prefix)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve board", "err", err)
		return trello.BoardResponse{}, err
	}
	s.logger.DebugContext(ctx, "setting runtime lists")
	s.logger.DebugContext(ctx, "retrieving lists for current board", "board", board.Name)
	lists, err := s.trello.ListBoardLists(ctx, board.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to retrieve lists for board", "board", board.Name, "err", err)
		return trello.BoardResponse{}, err
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
		return trello.BoardResponse{}, fmt.Errorf("could not find in and/or out lists")
	}
	return board, nil
}

func (s *Service) setLabels(ctx context.Context, board trello.BoardResponse) error {
	s.logger.DebugContext(ctx, "setting runtime labels")
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
