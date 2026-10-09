package config

import (
	"fmt"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

type Duration struct {
	time.Duration

	set bool
}

func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("duration must be a string")
	}
	value := strings.TrimSpace(node.Value)
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value, err)
	}
	d.Duration = parsed
	d.set = true
	return nil
}

type DateRange struct {
	Start time.Time
	End   time.Time
}

func (d *DateRange) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("date range must be a string")
	}
	dates := strings.Split(node.Value, "-")
	if len(dates) != 2 {
		return fmt.Errorf("date range must be in format {date}-{date}")
	}
	sd, err := time.Parse("2/1/06", dates[0])
	if err != nil {
		return fmt.Errorf("start date %s: %w", dates[0], err)
	}
	ed, err := time.Parse("2/1/06", dates[1])
	if err != nil {
		return fmt.Errorf("end date %s: %w", dates[1], err)
	}

	if sd.After(ed) || sd.Equal(ed) {
		return fmt.Errorf("start date must be before end date")
	}
	d.Start = sd
	d.End = ed
	return nil
}

func (d *DateRange) In(time time.Time) bool {
	return time.After(d.Start) && time.Before(d.End)
}
