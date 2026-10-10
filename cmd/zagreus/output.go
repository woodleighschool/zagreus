package main

import (
	"encoding/json"
	"io"

	"github.com/woodleighschool/zagreus/internal/app"
)

func writeSyncResults(writer io.Writer, results []app.Result) error {
	return writeJSON(writer, results)
}

func writePlans(writer io.Writer, plans []app.Intent) error {
	return writeJSON(writer, plans)
}

func writeJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}
