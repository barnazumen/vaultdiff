package diff

import (
	"bytes"
	"strings"
	"testing"
)

func TestRender_NoColor(t *testing.T) {
	result := &Result{
		Changes: []Change{
			{Key: "token", Type: Added, NewValue: "abc"},
			{Key: "old_key", Type: Removed, OldValue: "xyz"},
			{Key: "pass", Type: Modified, OldValue: "old", NewValue: "new"},
		},
	}

	var buf bytes.Buffer
	Render(&buf, result, FormatOptions{Color: false})
	out := buf.String()

	if !strings.Contains(out, "+ token = abc") {
		t.Errorf("missing added line, got:\n%s", out)
	}
	if !strings.Contains(out, "- old_key = xyz") {
		t.Errorf("missing removed line, got:\n%s", out)
	}
	if !strings.Contains(out, "~ pass") {
		t.Errorf("missing modified header, got:\n%s", out)
	}
}

func TestRender_ShowUnchanged(t *testing.T) {
	result := &Result{
		Changes: []Change{
			{Key: "stable", Type: Unchanged, OldValue: "v", NewValue: "v"},
		},
	}

	var buf bytes.Buffer
	Render(&buf, result, FormatOptions{ShowUnchanged: true})
	if !strings.Contains(buf.String(), "stable") {
		t.Error("expected unchanged key in output")
	}
}

func TestRender_HideUnchanged(t *testing.T) {
	result := &Result{
		Changes: []Change{
			{Key: "stable", Type: Unchanged, OldValue: "v", NewValue: "v"},
		},
	}

	var buf bytes.Buffer
	Render(&buf, result, FormatOptions{ShowUnchanged: false})
	if strings.Contains(buf.String(), "stable") {
		t.Error("expected unchanged key to be hidden")
	}
}
