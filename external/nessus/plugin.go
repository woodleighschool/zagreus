package nessus

import (
	"encoding/json"
	"fmt"
)

type PluginDetails struct {
	Name              string
	CVSS3Vector       *string
	CVSS3BaseScore    *string
	ExternalResources []string
	VulnAge           *string
}

type pluginResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	FamilyName string `json:"family_name"`
	Attributes []struct {
		Name  string `json:"attribute_name"`
		Value string `json:"attribute_value"`
	} `json:"attributes"`
}

func (p *PluginDetails) UnmarshalJSON(b []byte) error {
	var r pluginResponse

	if err := json.Unmarshal(b, &r); err != nil {
		return fmt.Errorf("failed to unmarshal plugin to intermediary struct")
	}
	p.Name = r.Name
	for _, attribute := range r.Attributes {
		switch attribute.Name {
		case "cvss3_vector":
			p.CVSS3Vector = &attribute.Value
		case "cvss3_base_score":
			p.CVSS3BaseScore = &attribute.Value
		case "see_also":
			p.ExternalResources = append(p.ExternalResources, attribute.Value)
		case "age_of_vuln":
			p.VulnAge = &attribute.Value
		default:
			continue
		}
	}
	return nil
}
