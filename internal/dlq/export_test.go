package dlq

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func newExportStore(t *testing.T) Store {
	t.Helper()
	s := NewInMemoryStore()
	jobs := []*Job{
		{ID: "j1", Queue: "email", Status: StatusFailed, Retries: 1, MaxRetries: 3, CreatedAt: time.Now(), LastError: errors.New("smtp timeout")},
		{ID: "j2", Queue: "sms", Status: StatusTerminal, Retries: 3, MaxRetries: 3, CreatedAt: time.Now()},
		{ID: "j3", Queue: "email", Status: StatusFailed, Retries: 0, MaxRetries: 5, CreatedAt: time.Now()},
	}
	for _, j := range jobs {
		if err := s.Save(j); err != nil {
			t.Fatalf("setup save: %v", err)
		}
	}
	return s
}

func TestExporter_JSON(t *testing.T) {
	e := NewExporter(newExportStore(t))
	var buf bytes.Buffer
	err := e.Export(&buf, ExportOptions{Format: FormatJSON})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var jobs []*Job
	if err := json.Unmarshal(buf.Bytes(), &jobs); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if len(jobs) != 3 {
		t.Errorf("expected 3 jobs, got %d", len(jobs))
	}
}

func TestExporter_JSON_WithFilter(t *testing.T) {
	e := NewExporter(newExportStore(t))
	var buf bytes.Buffer
	queue := "email"
	err := e.Export(&buf, ExportOptions{
		Format: FormatJSON,
		Filter: Filter{Queue: &queue},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var jobs []*Job
	if err := json.Unmarshal(buf.Bytes(), &jobs); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 email jobs, got %d", len(jobs))
	}
}

func TestExporter_CSV(t *testing.T) {
	e := NewExporter(newExportStore(t))
	var buf bytes.Buffer
	if err := e.Export(&buf, ExportOptions{Format: FormatCSV}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	// 1 header + 3 data rows
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "id,queue,status") {
		t.Errorf("unexpected header: %s", lines[0])
	}
}

func TestExporter_CSV_LastErrorIncluded(t *testing.T) {
	e := NewExporter(newExportStore(t))
	var buf bytes.Buffer
	if err := e.Export(&buf, ExportOptions{Format: FormatCSV}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "smtp timeout") {
		t.Error("expected last error in csv output")
	}
}

func TestExporter_UnsupportedFormat(t *testing.T) {
	e := NewExporter(newExportStore(t))
	var buf bytes.Buffer
	err := e.Export(&buf, ExportOptions{Format: "xml"})
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("unexpected error message: %v", err)
	}
}
