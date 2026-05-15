package dlq

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ExportFormat represents the output format for job exports.
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
)

// ExportOptions configures the export behavior.
type ExportOptions struct {
	Format  ExportFormat
	Filter  Filter
	Pretty  bool
}

// Exporter writes DLQ jobs to an io.Writer in the requested format.
type Exporter struct {
	store Store
}

// NewExporter creates an Exporter backed by the given Store.
func NewExporter(store Store) *Exporter {
	return &Exporter{store: store}
}

// Export writes all jobs matching opts.Filter to w in the specified format.
func (e *Exporter) Export(w io.Writer, opts ExportOptions) error {
	jobs, err := e.store.List(opts.Filter)
	if err != nil {
		return fmt.Errorf("export: list jobs: %w", err)
	}

	switch opts.Format {
	case FormatJSON:
		return e.writeJSON(w, jobs, opts.Pretty)
	case FormatCSV:
		return e.writeCSV(w, jobs)
	default:
		return fmt.Errorf("export: unsupported format %q", opts.Format)
	}
}

func (e *Exporter) writeJSON(w io.Writer, jobs []*Job, pretty bool) error {
	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(jobs); err != nil {
		return fmt.Errorf("export: encode json: %w", err)
	}
	return nil
}

func (e *Exporter) writeCSV(w io.Writer, jobs []*Job) error {
	_, err := fmt.Fprintln(w, "id,queue,status,retries,max_retries,created_at,last_error")
	if err != nil {
		return err
	}
	for _, j := range jobs {
		lastErr := ""
		if j.LastError != nil {
			lastErr = j.LastError.Error()
		}
		_, err = fmt.Fprintf(w, "%s,%s,%s,%d,%d,%s,%s\n",
			j.ID,
			j.Queue,
			j.Status,
			j.Retries,
			j.MaxRetries,
			j.CreatedAt.Format(time.RFC3339),
			lastErr,
		)
		if err != nil {
			return fmt.Errorf("export: write csv row: %w", err)
		}
	}
	return nil
}
