package trello

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

func (c *Client) CopyBoard(ctx context.Context, boardID string) error {
	// TODO: Implementation
	return nil
}
