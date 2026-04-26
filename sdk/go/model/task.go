package model

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// FlexibleTime wraps time.Time with a custom YAML unmarshaler that handles
// both quoted strings ("2025-01-15") and native YAML dates (2025-01-15).
type FlexibleTime struct {
	time.Time
}

// dateFormats lists formats to try when the YAML value is a quoted string.
var dateFormats = []string{
	"2006-01-02",
	time.RFC3339,
	"2006-01-02T15:04:05",
	time.DateTime,
	time.DateOnly,
}

func (ft *FlexibleTime) UnmarshalYAML(value *yaml.Node) error {
	// Native YAML timestamp tag — yaml.v3 already parsed it.
	if value.Tag == "!!timestamp" {
		var t time.Time
		if err := value.Decode(&t); err != nil {
			return err
		}
		ft.Time = t
		return nil
	}

	// Quoted string — try multiple date formats.
	s := value.Value
	for _, layout := range dateFormats {
		if t, err := time.Parse(layout, s); err == nil {
			ft.Time = t
			return nil
		}
	}

	return fmt.Errorf("cannot parse %q as a date", s)
}

// NewFlexibleTime wraps a time.Time into a FlexibleTime.
func NewFlexibleTime(t time.Time) FlexibleTime {
	return FlexibleTime{Time: t}
}

func (ft FlexibleTime) MarshalYAML() (any, error) {
	if ft.Time.IsZero() {
		return nil, nil
	}
	return ft.Time.Format("2006-01-02"), nil
}

// Status represents the current state of a task
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in-progress"
	StatusCompleted  Status = "completed"
	StatusInReview   Status = "in-review"
	StatusBlocked    Status = "blocked"
	StatusCancelled  Status = "cancelled"
)

// IsResolved returns true if the status represents a terminal state
// where the task is no longer active (completed or cancelled).
func (s Status) IsResolved() bool {
	return s == StatusCompleted || s == StatusCancelled
}

// Priority represents the importance level of a task
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// Effort represents the estimated effort required
type Effort string

const (
	EffortSmall  Effort = "small"
	EffortMedium Effort = "medium"
	EffortLarge  Effort = "large"
)

// TaskType represents the classification of a work item
type TaskType string

const (
	TypeFeature     TaskType = "feature"
	TypeBug         TaskType = "bug"
	TypeImprovement TaskType = "improvement"
	TypeChore       TaskType = "chore"
	TypeDocs        TaskType = "docs"
)

// VerifyStep represents a single verification check defined in task frontmatter.
type VerifyStep struct {
	Type  string `yaml:"type" json:"type"`
	Run   string `yaml:"run,omitempty" json:"run,omitempty"`
	Dir   string `yaml:"dir,omitempty" json:"dir,omitempty"`
	Check string `yaml:"check,omitempty" json:"check,omitempty"`
}

// ValidateVerifySteps checks that each step has a type and the required fields for that type.
// Returns a list of human-readable validation errors.
func ValidateVerifySteps(steps []VerifyStep) []string {
	var errs []string
	for i, s := range steps {
		if s.Type == "" {
			errs = append(errs, fmt.Sprintf("verify[%d]: missing required field 'type'", i))
			continue
		}
		switch s.Type {
		case "bash":
			if s.Run == "" {
				errs = append(errs, fmt.Sprintf("verify[%d]: bash step missing required field 'run'", i))
			}
		case "assert":
			if s.Check == "" {
				errs = append(errs, fmt.Sprintf("verify[%d]: assert step missing required field 'check'", i))
			}
		}
	}
	return errs
}

// Task represents a parsed task from a markdown file
type Task struct {
	// Frontmatter fields
	ID           string       `yaml:"id" json:"id"`
	Title        string       `yaml:"title" json:"title"`
	Status       Status       `yaml:"status" json:"status"`
	Priority     Priority     `yaml:"priority" json:"priority,omitempty"`
	Effort       Effort       `yaml:"effort" json:"effort,omitempty"`
	Type         TaskType     `yaml:"type" json:"type,omitempty"`
	Dependencies []string     `yaml:"dependencies" json:"dependencies"`
	Tags         []string     `yaml:"tags" json:"tags"`
	Touches      []string     `yaml:"touches" json:"touches,omitempty"`
	Context      []string     `yaml:"context" json:"context,omitempty"`
	Group        string       `yaml:"group" json:"group,omitempty"`
	Owner        string       `yaml:"owner" json:"owner,omitempty"`
	Phase        string       `yaml:"phase,omitempty" json:"phase,omitempty"`
	Parent       string       `yaml:"parent,omitempty" json:"parent,omitempty"`
	Created           FlexibleTime `yaml:"created_at" json:"created_at"`
	CreatedDeprecated FlexibleTime `yaml:"created" json:"-"`
	Completed    FlexibleTime `yaml:"completed_at,omitempty" json:"completed_at,omitempty"`
	CancelledAt  FlexibleTime `yaml:"cancelled_at,omitempty" json:"cancelled_at,omitempty"`
	Verify       []VerifyStep `yaml:"verify,omitempty" json:"verify,omitempty"`
	ExternalID   string       `yaml:"external_id,omitempty" json:"external_id,omitempty"`
	PRs          []string     `yaml:"pr,omitempty" json:"pr,omitempty"`

	// Content fields
	Body     string `json:"-"`
	FilePath string `json:"file_path"`

	// Worklog metadata (populated on demand, not from frontmatter)
	WorklogEntries int        `json:"worklog_entries,omitempty" yaml:"-"`
	WorklogUpdated *time.Time `json:"worklog_updated,omitempty" yaml:"-"`
}

// IsValid checks if the task has required fields
func (t *Task) IsValid() bool {
	return t.ID != "" && t.Title != ""
}

// GetGroup returns the group, prioritizing frontmatter over derived value
func (t *Task) GetGroup() string {
	return t.Group
}
