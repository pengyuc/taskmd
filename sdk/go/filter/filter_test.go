package filter

import (
	"testing"

	"github.com/driangle/taskmd/sdk/go/model"
)

func TestApply_OwnerFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Owner: "alice"},
		{ID: "002", Title: "Task B", Owner: "bob"},
		{ID: "003", Title: "Task C", Owner: ""},
	}

	filtered, err := Apply(tasks, []string{"owner=alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 task, got %d", len(filtered))
	}
	if filtered[0].ID != "001" {
		t.Errorf("expected task 001, got %s", filtered[0].ID)
	}
}

func TestApply_MultipleFilters(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Status: model.StatusPending, Owner: "alice"},
		{ID: "002", Title: "Task B", Status: model.StatusPending, Owner: "bob"},
		{ID: "003", Title: "Task C", Status: model.StatusCompleted, Owner: "alice"},
	}

	filtered, err := Apply(tasks, []string{"status=pending", "owner=alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 task, got %d", len(filtered))
	}
	if filtered[0].ID != "001" {
		t.Errorf("expected task 001, got %s", filtered[0].ID)
	}
}

func TestApply_InvalidFilterFormat(t *testing.T) {
	tasks := []*model.Task{{ID: "001"}}

	_, err := Apply(tasks, []string{"badfilter"})
	if err == nil {
		t.Fatal("expected error for invalid filter format")
	}
}

func TestApply_TypeFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Feature A", Type: model.TypeFeature},
		{ID: "002", Title: "Bug fix B", Type: model.TypeBug},
		{ID: "003", Title: "Chore C", Type: model.TypeChore},
		{ID: "004", Title: "No type"},
	}

	filtered, err := Apply(tasks, []string{"type=bug"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 task, got %d", len(filtered))
	}
	if filtered[0].ID != "002" {
		t.Errorf("expected task 002, got %s", filtered[0].ID)
	}
}

func TestApply_GroupWildcardFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Group: "cli/graph"},
		{ID: "002", Title: "Task B", Group: "cli/next"},
		{ID: "003", Title: "Task C", Group: "web/board"},
	}

	filtered, err := Apply(tasks, []string{"group=cli/*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(filtered))
	}
	if filtered[0].ID != "001" || filtered[1].ID != "002" {
		t.Errorf("expected tasks 001 and 002, got %s and %s", filtered[0].ID, filtered[1].ID)
	}
}

func TestApply_TouchesWildcardFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Touches: []string{"cli/graph", "web/board"}},
		{ID: "002", Title: "Task B", Touches: []string{"web/api"}},
		{ID: "003", Title: "Task C", Touches: []string{"docs"}},
	}

	filtered, err := Apply(tasks, []string{"touches=web/*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(filtered))
	}
	if filtered[0].ID != "001" || filtered[1].ID != "002" {
		t.Errorf("expected tasks 001 and 002, got %s and %s", filtered[0].ID, filtered[1].ID)
	}
}

func TestApply_GroupExactStillWorks(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Group: "cli"},
		{ID: "002", Title: "Task B", Group: "web"},
	}

	filtered, err := Apply(tasks, []string{"group=cli"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 1 || filtered[0].ID != "001" {
		t.Fatalf("expected task 001, got %v", filtered)
	}
}

func TestApply_ParentFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Parent"},
		{ID: "002", Title: "Child A", Parent: "001"},
		{ID: "003", Title: "Child B", Parent: "001"},
		{ID: "004", Title: "Orphan"},
	}

	t.Run("filter by parent ID", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"parent=001"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(filtered))
		}
	})

	t.Run("filter parent=true", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"parent=true"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 tasks with parent, got %d", len(filtered))
		}
	})

	t.Run("filter parent=false", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"parent=false"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 tasks without parent, got %d", len(filtered))
		}
	})
}

func TestApply_PriorityComparison(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Low", Priority: model.PriorityLow},
		{ID: "002", Title: "Medium", Priority: model.PriorityMedium},
		{ID: "003", Title: "High", Priority: model.PriorityHigh},
		{ID: "004", Title: "Critical", Priority: model.PriorityCritical},
		{ID: "005", Title: "Unset"},
	}

	tests := []struct {
		name        string
		expr        string
		expectedIDs []string
	}{
		{">=medium", "priority>=medium", []string{"002", "003", "004"}},
		{">medium", "priority>medium", []string{"003", "004"}},
		{"<=medium", "priority<=medium", []string{"001", "002"}},
		{"<medium", "priority<medium", []string{"001"}},
		{">=critical", "priority>=critical", []string{"004"}},
		{">critical", "priority>critical", nil},
		{"<=low", "priority<=low", []string{"001"}},
		{"<low", "priority<low", nil},
		{">=low (all set)", "priority>=low", []string{"001", "002", "003", "004"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered, err := Apply(tasks, []string{tt.expr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(filtered) != len(tt.expectedIDs) {
				t.Fatalf("expected %d tasks, got %d", len(tt.expectedIDs), len(filtered))
			}
			for i, id := range tt.expectedIDs {
				if filtered[i].ID != id {
					t.Errorf("result[%d]: expected %s, got %s", i, id, filtered[i].ID)
				}
			}
		})
	}
}

func TestApply_EffortComparison(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Small", Effort: model.EffortSmall},
		{ID: "002", Title: "Medium", Effort: model.EffortMedium},
		{ID: "003", Title: "Large", Effort: model.EffortLarge},
		{ID: "004", Title: "Unset"},
	}

	tests := []struct {
		name        string
		expr        string
		expectedIDs []string
	}{
		{">=medium", "effort>=medium", []string{"002", "003"}},
		{">small", "effort>small", []string{"002", "003"}},
		{"<=medium", "effort<=medium", []string{"001", "002"}},
		{"<large", "effort<large", []string{"001", "002"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered, err := Apply(tasks, []string{tt.expr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(filtered) != len(tt.expectedIDs) {
				t.Fatalf("expected %d tasks, got %d", len(tt.expectedIDs), len(filtered))
			}
			for i, id := range tt.expectedIDs {
				if filtered[i].ID != id {
					t.Errorf("result[%d]: expected %s, got %s", i, id, filtered[i].ID)
				}
			}
		})
	}
}

func TestApply_ComparisonErrors(t *testing.T) {
	tasks := []*model.Task{{ID: "001"}}

	tests := []struct {
		name string
		expr string
	}{
		{"unsupported field", "status>=pending"},
		{"invalid priority value", "priority>=unknown"},
		{"invalid effort value", "effort>huge"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Apply(tasks, []string{tt.expr})
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestApply_ComparisonCombinedWithEquality(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Status: model.StatusPending, Priority: model.PriorityLow},
		{ID: "002", Status: model.StatusPending, Priority: model.PriorityHigh},
		{ID: "003", Status: model.StatusCompleted, Priority: model.PriorityHigh},
	}

	filtered, err := Apply(tasks, []string{"status=pending", "priority>=high"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != "002" {
		t.Fatalf("expected task 002, got %v", filtered)
	}
}

func TestParseExpr(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		field   string
		op      string
		value   string
		wantErr bool
	}{
		{"equality", "status=pending", "status", "=", "pending", false},
		{"gte", "priority>=high", "priority", ">=", "high", false},
		{"gt", "priority>low", "priority", ">", "low", false},
		{"lte", "effort<=medium", "effort", "<=", "medium", false},
		{"lt", "effort<large", "effort", "<", "large", false},
		{"equality with spaces", " status = pending ", "status", "=", "pending", false},
		{"missing value", "badfilter", "", "", "", true},
		{"unsupported field for op", "status>=pending", "", "", "", true},
		{"invalid value for op", "priority>=nope", "", "", "", true},
		{"missing value after op", "priority>=", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := parseExpr(tt.expr)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.Field != tt.field || c.Op != tt.op || c.Value != tt.value {
				t.Errorf("got {%s %s %s}, want {%s %s %s}", c.Field, c.Op, c.Value, tt.field, tt.op, tt.value)
			}
		})
	}
}

func TestApply_PhaseFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Phase: "v0.2"},
		{ID: "002", Title: "Task B", Phase: "v0.3"},
		{ID: "003", Title: "Task C", Phase: "v0.2"},
		{ID: "004", Title: "Task D"},
	}

	t.Run("exact match", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"phase=v0.2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(filtered))
		}
		if filtered[0].ID != "001" || filtered[1].ID != "003" {
			t.Errorf("expected tasks 001 and 003, got %s and %s", filtered[0].ID, filtered[1].ID)
		}
	})

	t.Run("no match", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"phase=v1.0"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 0 {
			t.Fatalf("expected 0 tasks, got %d", len(filtered))
		}
	})
}
