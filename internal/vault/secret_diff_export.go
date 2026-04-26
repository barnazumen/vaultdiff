package vault

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/your-org/vaultdiff/internal/diff"
)

// DiffExportFormat defines the supported export formats for diff results.
type DiffExportFormat string

const (
	DiffExportJSON DiffExportFormat = "json"
	DiffExportCSV  DiffExportFormat = "csv"
)

// DiffExportRecord represents a single flattened diff entry for export.
type DiffExportRecord struct {
	Path      string    `json:"path"`
	Key       string    `json:"key"`
	ChangeType string   `json:"change_type"`
	OldValue  string    `json:"old_value,omitempty"`
	NewValue  string    `json:"new_value,omitempty"`
	ExportedAt time.Time `json:"exported_at"`
}

// DiffExportResult holds metadata and records for an exported diff.
type DiffExportResult struct {
	Path       string             `json:"path"`
	FromVersion int               `json:"from_version"`
	ToVersion   int               `json:"to_version"`
	ExportedAt  time.Time         `json:"exported_at"`
	Records     []DiffExportRecord `json:"records"`
}

// ExportDiffToFile writes the diff changes for a secret path to a file in the
// specified format (json or csv). It returns the number of records written.
func ExportDiffToFile(
	path string,
	fromVersion, toVersion int,
	changes []diff.Change,
	outputPath string,
	format DiffExportFormat,
) (int, error) {
	now := time.Now().UTC()

	records := make([]DiffExportRecord, 0, len(changes))
	for _, c := range changes {
		records = append(records, DiffExportRecord{
			Path:       path,
			Key:        c.Key,
			ChangeType: string(c.Type),
			OldValue:   c.OldValue,
			NewValue:   c.NewValue,
			ExportedAt: now,
		})
	}

	// Sort records by key for deterministic output.
	sort.Slice(records, func(i, j int) bool {
		return records[i].Key < records[j].Key
	})

	switch format {
	case DiffExportJSON:
		return len(records), exportDiffJSON(path, fromVersion, toVersion, records, outputPath, now)
	case DiffExportCSV:
		return len(records), exportDiffCSV(records, outputPath)
	default:
		return 0, fmt.Errorf("unsupported export format: %q", format)
	}
}

// exportDiffJSON writes the diff result as a structured JSON file.
func exportDiffJSON(
	path string,
	fromVersion, toVersion int,
	records []DiffExportRecord,
	outputPath string,
	now time.Time,
) error {
	result := DiffExportResult{
		Path:        path,
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		ExportedAt:  now,
		Records:     records,
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("encode diff JSON: %w", err)
	}
	return nil
}

// exportDiffCSV writes the diff records as a CSV file with a header row.
func exportDiffCSV(records []DiffExportRecord, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Write header.
	if err := w.Write([]string{"path", "key", "change_type", "old_value", "new_value", "exported_at"}); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	for _, r := range records {
		row := []string{
			r.Path,
			r.Key,
			r.ChangeType,
			r.OldValue,
			r.NewValue,
			r.ExportedAt.Format(time.RFC3339),
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}
	return nil
}
