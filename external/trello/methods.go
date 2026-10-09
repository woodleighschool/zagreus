package trello

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) ListBoards(ctx context.Context) ([]BoardResponse, error) {
	var result []BoardResponse
	body, err := c.request(ctx, http.MethodGet, "/members/me/boards", nil, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetLatestBoard(ctx context.Context, prefix *string) (BoardResponse, error) {
	var result BoardResponse
	body, err := c.request(ctx, http.MethodGet, "/members/me/boards", nil, nil, http.StatusOK)
	if err != nil {
		return BoardResponse{}, err
	}
	var boards []BoardResponse
	if err := json.Unmarshal(body, &boards); err != nil {
		return BoardResponse{}, err
	}
	result = boards[0]
	for _, board := range boards {
		if prefix != nil && !strings.HasPrefix(board.Name, *prefix) {
			continue
		}

		// TODO: I would like a warning for this scenario
		if result.LastActivity == nil || board.LastActivity == nil {
			continue
		}

		if result.LastActivity.Value().Before(board.LastActivity.Value()) {
			result = board
		}
	}

	return result, nil
}

func (c *Client) ListBoardLists(ctx context.Context, boardID string) ([]BoardListResponse, error) {
	var result []BoardListResponse
	url := fmt.Sprintf("/boards/%s/lists", boardID)
	body, err := c.request(ctx, http.MethodGet, url, nil, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) ListBoardLabels(ctx context.Context, boardID string) ([]BoardLabelResponse, error) {
	var result []BoardLabelResponse
	url := fmt.Sprintf("/boards/%s/labels", boardID)
	body, err := c.request(ctx, http.MethodGet, url, nil, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) CreateBoardList(ctx context.Context, boardID string, payload NewBoardListRequest) (BoardListResponse, error) {
	var result BoardListResponse
	url := fmt.Sprintf("/boards/%s/lists", boardID)
	if err := validate(payload); err != nil {
		return BoardListResponse{}, err
	}
	body, err := c.request(ctx, http.MethodPost, url, nil, payload, http.StatusOK)
	if err != nil {
		return BoardListResponse{}, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return BoardListResponse{}, err
	}
	return result, nil
}

func (c *Client) CreateBoardLabel(ctx context.Context, boardID string, payload NewBoardLabelRequest) error {
	url := fmt.Sprintf("/boards/%s/labels", boardID)
	if err := validate(payload); err != nil {
		return err
	}
	_, err := c.request(ctx, http.MethodPost, url, nil, payload, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ListBoardCards(ctx context.Context, boardID string) ([]CardResponse, error) {
	var result []CardResponse
	url := fmt.Sprintf("/boards/%s/cards", boardID)
	body, err := c.request(ctx, http.MethodGet, url, nil, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetChecklists(ctx context.Context, checklistIDs []string) ([]ChecklistResponse, error) {
	var result []ChecklistResponse
	for _, checklistID := range checklistIDs {
		var checklist ChecklistResponse
		url := fmt.Sprintf("/checklists/%s", checklistID)
		body, err := c.request(ctx, http.MethodGet, url, nil, nil, http.StatusOK)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, &checklist); err != nil {
			return nil, err
		}
		result = append(result, checklist)
	}
	return result, nil
}

func (c *Client) GetChecklistItems(ctx context.Context, checklistID string) ([]CheckitemResponse, error) {
	var result []CheckitemResponse
	url := fmt.Sprintf("/checklists/%s/checkItems", checklistID)
	body, err := c.request(ctx, http.MethodGet, url, nil, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) CreateCard(ctx context.Context, listID string, payload NewCardRequest) (CardResponse, error) {
	var result CardResponse
	if err := validate(payload); err != nil {
		return CardResponse{}, err
	}
	query := url.Values{}
	query.Add("idList", listID)
	body, err := c.request(ctx, http.MethodPost, "/cards", query, payload, http.StatusOK)
	if err != nil {
		return CardResponse{}, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return CardResponse{}, err
	}
	return result, nil
}

func (c *Client) UpdateCard(ctx context.Context, cardID string, payload UpdateCardRequest) error {
	url := fmt.Sprintf("/cards/%s", cardID)
	if err := validate(payload); err != nil {
		return err
	}
	_, err := c.request(ctx, http.MethodPut, url, nil, payload, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateChecklist(ctx context.Context, cardID string, payload NewChecklistRequest) (ChecklistResponse, error) {
	var result ChecklistResponse
	url := fmt.Sprintf("/cards/%s/checklists", cardID)
	if err := validate(payload); err != nil {
		return ChecklistResponse{}, err
	}
	body, err := c.request(ctx, http.MethodPost, url, nil, payload, http.StatusOK)
	if err != nil {
		return ChecklistResponse{}, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return ChecklistResponse{}, err
	}
	return result, nil
}

func (c *Client) CreateCheckitem(ctx context.Context, checklistID string, payload NewCheckitemRequest) error {
	url := fmt.Sprintf("/checklists/%s/checkItems", checklistID)
	if err := validate(payload); err != nil {
		return err
	}
	_, err := c.request(ctx, http.MethodPost, url, nil, payload, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateCheckitem(ctx context.Context, cardID string, checkitemID string, payload UpdateCheckitemRequest) error {
	url := fmt.Sprintf("/cards/%s/checkItem/%s", cardID, checkitemID)
	if err := validate(payload); err != nil {
		return err
	}
	_, err := c.request(ctx, http.MethodPut, url, nil, payload, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CopyBoard(_ context.Context, _ string) error {
	// TODO: Implementation
	return nil
}
