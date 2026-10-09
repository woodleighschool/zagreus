package main

import (
	"encoding/json"
	"io"

	"github.com/woodleighschool/zagreus/internal/app"
)

func writeSyncResults(writer io.Writer, results []app.Result) error {
	return writeJSON(writer, results)
}

func writePlans(writer io.Writer, plans []app.Intent, output string) error {
	if output == "json" {
		return writeJSON(writer, plans)
	}
	// TODO: Human output
	return nil
}

func writeJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}
