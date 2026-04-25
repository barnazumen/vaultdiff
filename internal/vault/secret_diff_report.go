package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/your/vaultdiff/internal/diff"
)

// DiffReport holds a structured summary of a diff between two secret versions.
type DiffReport struct {
	Path      string           `json:"path"`
	FromVersion int            `json:"from_version"`
	ToVersion   int            `json:"to_version"`
	Timestamp   time.Time      `json:"timestamp"`
	Changes     []diff.Change  `json:"changes"`
	Summary     ReportSummary  `json:"summary"`
}

// ReportSummary holds counts of change types.
type ReportSummary struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
	Unchanged int `json:"unchanged"`
}

// BuildDiffReport creates a DiffReport from a list of changes.
func BuildDiffReport(path string, fromVersion, toVersion int, changes []diff.Change) DiffReport {
	summary := ReportSummary{}
	for _, c := range changes {
		switch c.Type {
		case diff.Added:
			summary.Added++
		case diff.Removed:
			summary.Removed++
		case diff.Modified:
			summary.Modified++
		case diff.Unchanged:
			summary.Unchanged++
		}
	}
	return DiffReport{
		Path:        path,
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Timestamp:   time.Now().UTC(),
		Changes:     changes,
		Summary:     summary,
	}
}

// WriteReportJSON writes a DiffReport as JSON to the given file path.
func WriteReportJSON(report DiffReport, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report: %w", err)
	}
	return nil
}
