package nessus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetScanDetails(ctx context.Context, scanID int) ([]HostDetails, error) {
	var result []HostDetails
	scan, err := c.getScan(ctx, scanID)
	if err != nil {
		return nil, err
	}
	for _, host := range scan.Hosts {
		details, err := c.getScanHostDetails(ctx, scanID, host.ID)
		if err != nil {
			return nil, fmt.Errorf("unable to get scan details for host %d: %w", host.ID, err)
		}
		result = append(result, details)
	}
	return result, nil
}

func (c *Client) GetPluginDetails(ctx context.Context, pluginID int) (PluginDetails, error) {
	var result PluginDetails
	url := fmt.Sprintf("/plugins/plugin/%d", pluginID)
	body, err := c.request(ctx, http.MethodGet, url, nil, nil, http.StatusOK)
	if err != nil {
		return PluginDetails{}, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return PluginDetails{}, err
	}
	return result, nil
}

func (c *Client) getScan(ctx context.Context, scanID int) (Scan, error) {
	var result Scan
	url := fmt.Sprintf("/scans/%d", scanID)
	body, err := c.request(ctx, http.MethodGet, url, nil, nil, http.StatusOK)
	if err != nil {
		return Scan{}, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return Scan{}, err
	}
	return result, nil
}

func (c *Client) getScanHostDetails(ctx context.Context, scanID int, hostID int) (HostDetails, error) {
	var result HostDetails
	url := fmt.Sprintf("/scans/%d/hosts/%d", scanID, hostID)
	body, err := c.request(ctx, http.MethodGet, url, nil, nil, http.StatusOK)
	if err != nil {
		return HostDetails{}, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return HostDetails{}, err
	}
	return result, nil
}
