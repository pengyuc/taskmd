package parser

import (
	"testing"
	"time"

	"github.com/driangle/taskmd/sdk/go/model"
)

func TestParseTaskContent_ValidTask(t *testing.T) {
	content := []byte(`---
id: "001"
title: "Test Task"
status: pending
priority: high
effort: medium
dependencies: ["002", "003"]
tags:
  - test
  - cli
group: backend
created: 2026-02-08
---

# Test Task

This is the task body.

- Item 1
- Item 2
`)

	task, err := ParseTaskContent("test.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "001" {
		t.Errorf("expected ID '001', got '%s'", task.ID)
	}

	if task.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", task.Title)
	}

	if task.Status != model.StatusPending {
		t.Errorf("expected status 'pending', got '%s'", task.Status)
	}

	if task.Priority != model.PriorityHigh {
		t.Errorf("expected priority 'high', got '%s'", task.Priority)
	}

	if task.Effort != model.EffortMedium {
		t.Errorf("expected effort 'medium', got '%s'", task.Effort)
	}

	if len(task.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(task.Dependencies))
	}

	if len(task.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(task.Tags))
	}

	if task.Group != "backend" {
		t.Errorf("expected group 'backend', got '%s'", task.Group)
	}

	expectedDate := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	if !task.Created.Equal(expectedDate) {
		t.Errorf("expected created date %v, got %v", expectedDate, task.Created)
	}

	if task.Body == "" {
		t.Error("expected body to be parsed")
	}

	if task.FilePath != "test.md" {
		t.Errorf("expected FilePath 'test.md', got '%s'", task.FilePath)
	}
}

func TestParseTaskContent_Phase(t *testing.T) {
	content := []byte(`---
id: "050"
title: "Phase Task"
status: pending
phase: "v0.2"
---

# Phase Task

A task with a phase.
`)

	task, err := ParseTaskContent("test.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.Phase != "v0.2" {
		t.Errorf("expected phase 'v0.2', got '%s'", task.Phase)
	}
}

func TestParseTaskContent_NoPhase(t *testing.T) {
	content := []byte(`---
id: "051"
title: "No Phase"
status: pending
---

Body.
`)

	task, err := ParseTaskContent("test.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.Phase != "" {
		t.Errorf("expected empty phase, got '%s'", task.Phase)
	}
}

func TestParseTaskContent_QuotedCreatedDate(t *testing.T) {
	content := []byte(`---
id: "010"
title: "Quoted Date Task"
status: pending
created: "2026-02-23"
---

# Quoted Date Task
`)

	task, err := ParseTaskContent("test.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedDate := time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC)
	if !task.Created.Equal(expectedDate) {
		t.Errorf("expected created date %v, got %v", expectedDate, task.Created)
	}
}

func TestParseTaskContent_MinimalTask(t *testing.T) {
	content := []byte(`---
id: "002"
title: "Minimal Task"
---

Body content here.
`)

	task, err := ParseTaskContent("minimal.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "002" {
		t.Errorf("expected ID '002', got '%s'", task.ID)
	}

	if task.Title != "Minimal Task" {
		t.Errorf("expected title 'Minimal Task', got '%s'", task.Title)
	}

	if task.Body != "Body content here." {
		t.Errorf("expected body 'Body content here.', got '%s'", task.Body)
	}
}

func TestParseTaskContent_EmptyFile(t *testing.T) {
	content := []byte("")

	_, err := ParseTaskContent("empty.md", content)
	if err == nil {
		t.Error("expected error for empty file")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %T", err)
	} else if parseErr.FilePath != "empty.md" {
		t.Errorf("expected FilePath 'empty.md', got '%s'", parseErr.FilePath)
	}
}

func TestParseTaskContent_MissingFrontmatter(t *testing.T) {
	content := []byte(`# No Frontmatter

This file has no frontmatter.
`)

	_, err := ParseTaskContent("no-frontmatter.md", content)
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}
}

func TestParseTaskContent_MalformedYAML(t *testing.T) {
	content := []byte(`---
id: "003"
title: "Bad YAML"
invalid: [unclosed
---

Body
`)

	_, err := ParseTaskContent("bad-yaml.md", content)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %T", err)
	} else if !contains(parseErr.Message, "YAML") {
		t.Errorf("expected error message to mention YAML, got: %s", parseErr.Message)
	}
}

func TestParseTaskContent_UnclosedFrontmatter(t *testing.T) {
	content := []byte(`---
id: "004"
title: "Unclosed"

This has no closing delimiter
`)

	_, err := ParseTaskContent("unclosed.md", content)
	if err == nil {
		t.Error("expected error for unclosed frontmatter")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %T", err)
	} else if !contains(parseErr.Message, "frontmatter") {
		t.Errorf("expected error message to mention frontmatter, got: %s", parseErr.Message)
	}
}

func TestParseTaskContent_MissingID(t *testing.T) {
	content := []byte(`---
title: "No ID"
---

Body
`)

	_, err := ParseTaskContent("no-id.md", content)
	if err == nil {
		t.Error("expected error for missing ID")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %T", err)
	} else if !contains(parseErr.Message, "required fields") {
		t.Errorf("expected error message to mention required fields, got: %s", parseErr.Message)
	}
}

func TestParseTaskContent_MissingTitle(t *testing.T) {
	content := []byte(`---
id: "005"
---

Body
`)

	_, err := ParseTaskContent("no-title.md", content)
	if err == nil {
		t.Error("expected error for missing title")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %T", err)
	} else if !contains(parseErr.Message, "required fields") {
		t.Errorf("expected error message to mention required fields, got: %s", parseErr.Message)
	}
}

func TestParseTaskContent_EmptyBody(t *testing.T) {
	content := []byte(`---
id: "006"
title: "Empty Body"
---
`)

	task, err := ParseTaskContent("empty-body.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.Body != "" {
		t.Errorf("expected empty body, got '%s'", task.Body)
	}
}

func TestExtractFrontmatter_NoFrontmatter(t *testing.T) {
	content := []byte("Just plain content")

	frontmatter, body, err := extractFrontmatter(content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(frontmatter) != 0 {
		t.Error("expected empty frontmatter")
	}

	if body != "Just plain content" {
		t.Errorf("expected body to be full content, got '%s'", body)
	}
}

func TestExtractFrontmatter_WithWhitespace(t *testing.T) {
	content := []byte(`---
id: "007"
title: "Whitespace Test"
---

Body with leading/trailing whitespace

`)

	frontmatter, body, err := extractFrontmatter(content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(frontmatter) == 0 {
		t.Error("expected frontmatter")
	}

	if body != "Body with leading/trailing whitespace" {
		t.Errorf("expected trimmed body, got '%s'", body)
	}
}

func TestParseTaskContent_DeriveIDFromFilename(t *testing.T) {
	content := []byte(`---
title: "My Task"
status: pending
---

Body
`)

	task, err := ParseTaskContent("009-add-feature.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "009" {
		t.Errorf("expected ID '009' derived from filename, got '%s'", task.ID)
	}

	if task.Title != "My Task" {
		t.Errorf("expected title from frontmatter 'My Task', got '%s'", task.Title)
	}
}

func TestParseTaskContent_DeriveTitleFromFilename(t *testing.T) {
	content := []byte(`---
status: pending
---

Body
`)

	task, err := ParseTaskContent("012-setup-database.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "012" {
		t.Errorf("expected ID '012', got '%s'", task.ID)
	}

	if task.Title != "setup database" {
		t.Errorf("expected title 'setup database' derived from filename, got '%s'", task.Title)
	}
}

func TestParseTaskContent_NonNumericFilenameNoDerivation(t *testing.T) {
	content := []byte(`---
title: "Some Task"
---

Body
`)

	_, err := ParseTaskContent("readme.md", content)
	if err == nil {
		t.Error("expected error for non-numeric filename with missing ID")
	}
}

func TestParseTaskContent_ExternalID(t *testing.T) {
	content := []byte(`---
id: "042"
title: "Synced from Jira"
status: pending
external_id: "PROJ-123"
---

# Synced from Jira

This task was synced from an external system.
`)

	task, err := ParseTaskContent("042-synced-from-jira.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ExternalID != "PROJ-123" {
		t.Errorf("expected ExternalID 'PROJ-123', got '%s'", task.ExternalID)
	}
}

func TestParseTaskContent_ExternalIDEmpty(t *testing.T) {
	content := []byte(`---
id: "043"
title: "Regular task"
status: pending
---

# Regular task
`)

	task, err := ParseTaskContent("043-regular-task.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ExternalID != "" {
		t.Errorf("expected empty ExternalID, got '%s'", task.ExternalID)
	}
}

func TestDeriveFieldsFromFilename_PrefixedID(t *testing.T) {
	content := []byte(`---
title: "Fix Login"
status: pending
---

Body
`)

	task, err := ParseTaskContent("dr-001-fix-login.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "dr-001" {
		t.Errorf("expected ID 'dr-001' derived from prefixed filename, got '%s'", task.ID)
	}

	if task.Title != "Fix Login" {
		t.Errorf("expected title from frontmatter, got '%s'", task.Title)
	}
}

func TestDeriveFieldsFromFilename_PrefixedTitleDerivation(t *testing.T) {
	content := []byte(`---
status: pending
---

Body
`)

	task, err := ParseTaskContent("dr-001-fix-login.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "dr-001" {
		t.Errorf("expected ID 'dr-001', got '%s'", task.ID)
	}

	if task.Title != "fix login" {
		t.Errorf("expected title 'fix login' derived from slug, got '%s'", task.Title)
	}
}

func TestDeriveFieldsFromFilename_RandomID(t *testing.T) {
	content := []byte(`---
status: pending
---

Body
`)

	task, err := ParseTaskContent("a3f9x2-slug-title.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "a3f9x2" {
		t.Errorf("expected ID 'a3f9x2' from random filename, got '%s'", task.ID)
	}

	if task.Title != "slug title" {
		t.Errorf("expected title 'slug title', got '%s'", task.Title)
	}
}

func TestDeriveFieldsFromFilename_SequentialRegression(t *testing.T) {
	content := []byte(`---
status: pending
---

Body
`)

	task, err := ParseTaskContent("009-add-feature.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "009" {
		t.Errorf("expected ID '009', got '%s'", task.ID)
	}

	if task.Title != "add feature" {
		t.Errorf("expected title 'add feature', got '%s'", task.Title)
	}
}

func TestDeriveFieldsFromFilename_NonIDRejection(t *testing.T) {
	content := []byte(`---
title: "Some Task"
---

Body
`)

	_, err := ParseTaskContent("readme.md", content)
	if err == nil {
		t.Error("expected error for non-ID filename with missing ID")
	}
}

func TestDeriveFieldsFromFilename_TruncatedUUID(t *testing.T) {
	content := []byte(`---
status: pending
---

Body
`)

	task, err := ParseTaskContent("f47ac10b-fix-bug.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "f47ac10b" {
		t.Errorf("expected ID 'f47ac10b' from truncated UUID filename, got '%s'", task.ID)
	}

	if task.Title != "fix bug" {
		t.Errorf("expected title 'fix bug', got '%s'", task.Title)
	}
}

func TestDeriveFieldsFromFilename_LongHexUUID(t *testing.T) {
	content := []byte(`---
status: pending
---

Body
`)

	task, err := ParseTaskContent("f47ac10b58cc-fix-bug.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "f47ac10b58cc" {
		t.Errorf("expected ID 'f47ac10b58cc' from long hex UUID filename, got '%s'", task.ID)
	}

	if task.Title != "fix bug" {
		t.Errorf("expected title 'fix bug', got '%s'", task.Title)
	}
}

func TestDeriveFieldsFromFilename_FullUUID(t *testing.T) {
	content := []byte(`---
status: pending
---

Body
`)

	task, err := ParseTaskContent("f47ac10b-58cc-4372-a567-0e02b2c3d479-my-task.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "f47ac10b-58cc-4372-a567-0e02b2c3d479" {
		t.Errorf("expected full UUID ID, got '%s'", task.ID)
	}

	if task.Title != "my task" {
		t.Errorf("expected title 'my task', got '%s'", task.Title)
	}
}

func TestDeriveFieldsFromFilename_FullUUIDNoSlug(t *testing.T) {
	content := []byte(`---
title: "UUID Task"
status: pending
---

Body
`)

	task, err := ParseTaskContent("f47ac10b-58cc-4372-a567-0e02b2c3d479.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "f47ac10b-58cc-4372-a567-0e02b2c3d479" {
		t.Errorf("expected full UUID ID, got '%s'", task.ID)
	}

	if task.Title != "UUID Task" {
		t.Errorf("expected title from frontmatter, got '%s'", task.Title)
	}
}

func TestSplitFilenameID_UUIDPatterns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantID   string
		wantSlug string
	}{
		{"truncated 8-char hex", "f47ac10b-fix-bug", "f47ac10b", "fix-bug"},
		{"long hex 12 chars", "f47ac10b58cc-slug", "f47ac10b58cc", "slug"},
		{"full UUID with slug", "f47ac10b-58cc-4372-a567-0e02b2c3d479-slug", "f47ac10b-58cc-4372-a567-0e02b2c3d479", "slug"},
		{"full UUID no slug", "f47ac10b-58cc-4372-a567-0e02b2c3d479", "f47ac10b-58cc-4372-a567-0e02b2c3d479", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, slug := splitFilenameID(tt.input)
			if id != tt.wantID {
				t.Errorf("splitFilenameID(%q) id = %q, want %q", tt.input, id, tt.wantID)
			}
			if slug != tt.wantSlug {
				t.Errorf("splitFilenameID(%q) slug = %q, want %q", tt.input, slug, tt.wantSlug)
			}
		})
	}
}

func TestIsHexID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"f47ac10b5", true},                          // 9 chars, minimum for hex ID
		{"f47ac10b58cc", true},                       // 12 chars
		{"0123456789abcdef0123456789abcdef", true},   // 32 chars, max
		{"f47ac10b", false},                          // 8 chars, too short (handled by alphanumeric)
		{"f47ac10b58cc4372a5670e02b2c3d4791", false}, // 33 chars, too long
		{"ghijklmn0", false},                         // non-hex chars
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isHexID(tt.input); got != tt.want {
				t.Errorf("isHexID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"123", true},
		{"0", true},
		{"", false},
		{"12a", false},
		{"abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isNumeric(tt.input); got != tt.want {
				t.Errorf("isNumeric(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsAlphanumericID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"a3f9x2", true},
		{"ab1", true},
		{"a1b2c3d4", true},
		{"abc", false},
		{"readme", false},
		{"ab", false},
		{"a1b2c3d4e", false},
		{"A3F9X2", false},
		{"a3-f9", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isAlphanumericID(tt.input); got != tt.want {
				t.Errorf("isAlphanumericID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseTaskContent_CreatedAtField(t *testing.T) {
	content := []byte(`---
id: "060"
title: "New style date"
status: pending
created_at: 2026-03-15
---

# New style date
`)

	task, err := ParseTaskContent("test.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	if !task.Created.Equal(expectedDate) {
		t.Errorf("expected created date %v, got %v", expectedDate, task.Created)
	}
}

func TestParseTaskContent_CreatedAtPreferredOverCreated(t *testing.T) {
	content := []byte(`---
id: "061"
title: "Both dates"
status: pending
created_at: 2026-04-01
created: 2026-01-01
---

# Both dates
`)

	task, err := ParseTaskContent("test.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if !task.Created.Equal(expectedDate) {
		t.Errorf("expected created_at to take precedence, got %v", task.Created)
	}
}

func TestParseTaskContent_DeprecatedCreatedStillWorks(t *testing.T) {
	content := []byte(`---
id: "062"
title: "Old style date"
status: pending
created: 2026-02-20
---

# Old style date
`)

	task, err := ParseTaskContent("test.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedDate := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	if !task.Created.Equal(expectedDate) {
		t.Errorf("expected deprecated created field to work, got %v", task.Created)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
