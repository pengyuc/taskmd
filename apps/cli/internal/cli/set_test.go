package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/driangle/taskmd/sdk/go/taskfile"
)

func createSetTestFiles(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	tasks := map[string]string{
		"001-setup.md": `---
id: "001"
title: "Setup project"
status: pending
priority: high
effort: small
dependencies: []
tags: ["infra"]
created: 2026-02-08
---

# Setup project

Initial project setup with build tooling.
`,
		"002-auth.md": `---
id: "002"
title: "Implement authentication"
status: in-progress
priority: critical
effort: large
dependencies: ["001"]
tags: ["backend", "security"]
created: 2026-02-08
---

# Implement authentication

Add JWT-based auth with refresh tokens.
`,
		"003-ui.md": `---
id: "003"
title: "Build UI components"
status: blocked
priority: medium
effort: medium
dependencies: ["002"]
tags: ["frontend"]
created: 2026-02-08
---

# Build UI components

Create reusable component library.
`,
	}

	for filename, content := range tasks {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	return tmpDir
}

func createMultilineTagTestFile(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	content := `---
id: "010"
title: "Multiline tags task"
status: pending
priority: high
effort: small
dependencies: []
tags:
  - backend
  - api
created: 2026-02-08
---

# Multiline tags task

Task with multiline YAML tags.
`
	path := filepath.Join(tmpDir, "010-multiline.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return tmpDir
}

func resetSetFlags() {
	setTaskID = ""
	setStatus = ""
	setPriority = ""
	setEffort = ""
	setType = ""
	setOwner = ""
	setParent = ""
	setDependsOn = ""
	setPhase = ""
	setDone = false
	setDryRun = false
	setVerify = false
	setAddTags = nil
	setRemoveTags = nil
	setAddPRs = nil
	setRemovePRs = nil
	setAddTouches = nil
	setRemoveTouches = nil
	taskDir = "."

	// Reset cobra flag Changed state to avoid test interference
	for _, name := range []string{"status", "parent", "depends-on", "phase"} {
		if f := setCmd.Flags().Lookup(name); f != nil {
			f.Changed = false
		}
	}
}

func captureSetOutput(t *testing.T) (string, error) {
	t.Helper()
	return captureSetOutputWithArgs(t, nil)
}

func captureSetOutputWithArgs(t *testing.T, args []string) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runSet(setCmd, args)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

func TestSet_Status(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setStatus = "completed"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Updated task 001") {
		t.Error("Expected confirmation message")
	}
	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("Expected status change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "status: completed") {
		t.Error("Expected file to contain updated status")
	}
}

func TestSet_Priority(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setPriority = "low"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "priority: high -> low") {
		t.Errorf("Expected priority change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "priority: low") {
		t.Error("Expected file to contain updated priority")
	}
}

func TestSet_Effort(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "002"
	setEffort = "small"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "effort: large -> small") {
		t.Errorf("Expected effort change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	if !strings.Contains(string(content), "effort: small") {
		t.Error("Expected file to contain updated effort")
	}
}

func TestSet_Owner(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setOwner = "alice"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "owner: (unset) -> alice") {
		t.Errorf("Expected owner change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "owner: alice") {
		t.Errorf("Expected file to contain owner: alice, got:\n%s", string(content))
	}
}

func TestSet_OwnerUpdateExisting(t *testing.T) {
	tmpDir := t.TempDir()

	content := `---
id: "020"
title: "Task with owner"
status: pending
owner: alice
created: 2026-02-08
---

# Task with owner
`
	path := filepath.Join(tmpDir, "020-owned.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "020"
	setOwner = "bob"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "owner: alice -> bob") {
		t.Errorf("Expected owner change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(path)
	if !strings.Contains(string(updated), "owner: bob") {
		t.Errorf("Expected file to contain owner: bob, got:\n%s", string(updated))
	}
}

func TestSet_DoneFlag(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setDone = true

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("Expected --done to set status to completed, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "status: completed") {
		t.Error("Expected file to contain completed status")
	}
}

func TestSet_MultipleFields(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "003"
	setStatus = "in-progress"
	setPriority = "critical"
	setEffort = "large"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status: blocked -> in-progress") {
		t.Error("Expected status change in output")
	}
	if !strings.Contains(output, "priority: medium -> critical") {
		t.Error("Expected priority change in output")
	}
	if !strings.Contains(output, "effort: medium -> large") {
		t.Error("Expected effort change in output")
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "003-ui.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "status: in-progress") {
		t.Error("Expected file to contain updated status")
	}
	if !strings.Contains(fileStr, "priority: critical") {
		t.Error("Expected file to contain updated priority")
	}
	if !strings.Contains(fileStr, "effort: large") {
		t.Error("Expected file to contain updated effort")
	}
}

func TestSet_AllValidStatuses(t *testing.T) {
	statuses := []string{"pending", "in-progress", "completed", "in-review", "blocked", "cancelled"}
	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			tmpDir := createSetTestFiles(t)
			resetSetFlags()
			taskDir = tmpDir
			setTaskID = "001"
			setStatus = status

			_, err := captureSetOutput(t)
			if err != nil {
				t.Fatalf("unexpected error for status %q: %v", status, err)
			}

			content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
			if !strings.Contains(string(content), "status: "+status) {
				t.Errorf("Expected file to contain status: %s", status)
			}
		})
	}
}

func TestSet_CancelledStatus(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "002"
	setStatus = "cancelled"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error when setting status to cancelled: %v", err)
	}

	if !strings.Contains(output, "status: in-progress -> cancelled") {
		t.Errorf("Expected status change to cancelled in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	if !strings.Contains(string(content), "status: cancelled") {
		t.Error("Expected file to contain status: cancelled")
	}
}

func TestSet_AllValidPriorities(t *testing.T) {
	priorities := []string{"low", "medium", "high", "critical"}
	for _, priority := range priorities {
		t.Run(priority, func(t *testing.T) {
			tmpDir := createSetTestFiles(t)
			resetSetFlags()
			taskDir = tmpDir
			setTaskID = "001"
			setPriority = priority

			_, err := captureSetOutput(t)
			if err != nil {
				t.Fatalf("unexpected error for priority %q: %v", priority, err)
			}

			content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
			if !strings.Contains(string(content), "priority: "+priority) {
				t.Errorf("Expected file to contain priority: %s", priority)
			}
		})
	}
}

func TestSet_AllValidEfforts(t *testing.T) {
	efforts := []string{"small", "medium", "large"}
	for _, effort := range efforts {
		t.Run(effort, func(t *testing.T) {
			tmpDir := createSetTestFiles(t)
			resetSetFlags()
			taskDir = tmpDir
			setTaskID = "002"
			setEffort = effort

			_, err := captureSetOutput(t)
			if err != nil {
				t.Fatalf("unexpected error for effort %q: %v", effort, err)
			}

			content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
			if !strings.Contains(string(content), "effort: "+effort) {
				t.Errorf("Expected file to contain effort: %s", effort)
			}
		})
	}
}

func TestSet_InvalidStatus(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setStatus = "invalid"

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("Expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("Expected 'invalid status' error, got: %v", err)
	}
}

func TestSet_InvalidPriority(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setPriority = "urgent"

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("Expected error for invalid priority")
	}
	if !strings.Contains(err.Error(), "invalid priority") {
		t.Errorf("Expected 'invalid priority' error, got: %v", err)
	}
}

func TestSet_InvalidEffort(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setEffort = "huge"

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("Expected error for invalid effort")
	}
	if !strings.Contains(err.Error(), "invalid effort") {
		t.Errorf("Expected 'invalid effort' error, got: %v", err)
	}
}

func TestSet_Type(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setType = "bug"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "type: (unset) -> bug") {
		t.Errorf("Expected type change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "type: bug") {
		t.Errorf("Expected file to contain type: bug, got:\n%s", string(content))
	}
}

func TestSet_InvalidType(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setType = "task"

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("Expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "invalid type") {
		t.Errorf("Expected 'invalid type' error, got: %v", err)
	}
}

func TestSet_TaskNotFound(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "nonexistent"
	setStatus = "completed"

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("Expected error for non-existent task")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestSet_NoFlagsProvided(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("Expected error when no update flags provided")
	}
	if !strings.Contains(err.Error(), "nothing to update") {
		t.Errorf("Expected 'nothing to update' error, got: %v", err)
	}
}

func TestSet_DoneWithStatusMutuallyExclusive(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setDone = true
	setStatus = "blocked"

	// Mark the --status flag as changed to simulate CLI usage
	setCmd.Flags().Set("status", "blocked")
	defer func() {
		// Reset the changed state by creating a fresh flag set lookup
		setCmd.Flags().Set("status", "")
	}()

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("Expected error when --done and --status are both set")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("Expected 'mutually exclusive' error, got: %v", err)
	}
}

func TestSet_BodyPreserved(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "002"
	setStatus = "completed"

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	fileStr := string(content)

	if !strings.Contains(fileStr, "# Implement authentication") {
		t.Error("Expected body heading to be preserved")
	}
	if !strings.Contains(fileStr, "Add JWT-based auth with refresh tokens.") {
		t.Error("Expected body content to be preserved")
	}
}

func TestSet_OtherFieldsPreserved(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "002"
	setStatus = "completed"

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	fileStr := string(content)

	// Verify non-updated fields are preserved
	if !strings.Contains(fileStr, "priority: critical") {
		t.Error("Expected priority to be preserved")
	}
	if !strings.Contains(fileStr, "effort: large") {
		t.Error("Expected effort to be preserved")
	}
	if !strings.Contains(fileStr, `dependencies: ["001"]`) {
		t.Error("Expected dependencies to be preserved")
	}
	if !strings.Contains(fileStr, `tags: ["backend", "security"]`) {
		t.Error("Expected tags to be preserved")
	}
	if !strings.Contains(fileStr, "created: 2026-02-08") {
		t.Error("Expected created date to be preserved")
	}
}

func TestSet_FrontmatterBounds(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string
		wantOpen  int
		wantClose int
	}{
		{
			name:      "standard frontmatter",
			lines:     []string{"---", "id: foo", "---", "body"},
			wantOpen:  0,
			wantClose: 2,
		},
		{
			name:      "no frontmatter",
			lines:     []string{"# Just a heading", "body"},
			wantOpen:  -1,
			wantClose: -1,
		},
		{
			name:      "unclosed frontmatter",
			lines:     []string{"---", "id: foo"},
			wantOpen:  -1,
			wantClose: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			open, closeIdx := taskfile.FindFrontmatterBounds(tt.lines)
			if open != tt.wantOpen || closeIdx != tt.wantClose {
				t.Errorf("findFrontmatterBounds() = (%d, %d), want (%d, %d)",
					open, closeIdx, tt.wantOpen, tt.wantClose)
			}
		})
	}
}

func TestSet_MatchByTitle(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "Setup project"
	setStatus = "completed"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Updated task 001") {
		t.Error("Expected confirmation for task found by title match")
	}
}

func TestSet_AddSingleTag(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setAddTags = []string{"new-tag"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "tags: [infra] -> [infra, new-tag]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra", "new-tag"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestSet_AddMultipleTags(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setAddTags = []string{"tag-a", "tag-b"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "tags: [infra] -> [infra, tag-a, tag-b]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra", "tag-a", "tag-b"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestSet_RemoveTag(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "002"
	setRemoveTags = []string{"security"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "tags: [backend, security] -> [backend]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	if !strings.Contains(string(content), `tags: ["backend"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestSet_AddAndRemoveTag(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "002"
	setAddTags = []string{"new-feature"}
	setRemoveTags = []string{"security"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "tags: [backend, security] -> [backend, new-feature]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	if !strings.Contains(string(content), `tags: ["backend", "new-feature"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestSet_AddDuplicateTag(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setAddTags = []string{"infra"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Tags should remain unchanged since "infra" already exists.
	if !strings.Contains(output, "tags: [infra] -> [infra]") {
		t.Errorf("Expected no-op tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra"]`) {
		t.Errorf("Expected tags to remain unchanged, got:\n%s", string(content))
	}
}

func TestSet_RemoveNonexistentTag(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setRemoveTags = []string{"nonexistent"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Tags should remain unchanged since "nonexistent" isn't present.
	if !strings.Contains(output, "tags: [infra] -> [infra]") {
		t.Errorf("Expected no-op tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra"]`) {
		t.Errorf("Expected tags to remain unchanged, got:\n%s", string(content))
	}
}

func TestSet_TagOnlyUpdate(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setAddTags = []string{"new-tag"}

	// Should NOT produce "nothing to update" error.
	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("tag-only update should succeed, got error: %v", err)
	}
}

func TestSet_TagsWithOtherFlags(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setStatus = "completed"
	setAddTags = []string{"done-tag"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status: pending -> completed") {
		t.Error("Expected status change in output")
	}
	if !strings.Contains(output, "tags: [infra] -> [infra, done-tag]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "status: completed") {
		t.Error("Expected file to contain updated status")
	}
	if !strings.Contains(fileStr, `tags: ["infra", "done-tag"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", fileStr)
	}
}

func TestSet_TagsPreservedFormat(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setAddTags = []string{"extra"}

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	fileStr := string(content)

	// Inline format should stay inline.
	if !strings.Contains(fileStr, `tags: ["infra", "extra"]`) {
		t.Errorf("Expected inline tag format to be preserved, got:\n%s", fileStr)
	}

	// Other fields should be preserved.
	if !strings.Contains(fileStr, "status: pending") {
		t.Error("Expected status to be preserved")
	}
	if !strings.Contains(fileStr, "# Setup project") {
		t.Error("Expected body to be preserved")
	}
}

func TestSet_MultilineTagFormat(t *testing.T) {
	tmpDir := createMultilineTagTestFile(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "010"
	setAddTags = []string{"new-tag"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "tags: [backend, api] -> [backend, api, new-tag]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "010-multiline.md"))
	fileStr := string(content)

	// Multiline format should stay multiline.
	if !strings.Contains(fileStr, "tags:\n  - backend\n  - api\n  - new-tag") {
		t.Errorf("Expected multiline tag format to be preserved, got:\n%s", fileStr)
	}

	// Other fields should be preserved.
	if !strings.Contains(fileStr, "status: pending") {
		t.Error("Expected status to be preserved")
	}
	if !strings.Contains(fileStr, "# Multiline tags task") {
		t.Error("Expected body to be preserved")
	}
}

func TestSet_MultilineTagRemove(t *testing.T) {
	tmpDir := createMultilineTagTestFile(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "010"
	setRemoveTags = []string{"api"}

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "010-multiline.md"))
	fileStr := string(content)

	if !strings.Contains(fileStr, "tags:\n  - backend\ncreated:") {
		t.Errorf("Expected multiline format with 'api' removed, got:\n%s", fileStr)
	}
}

func TestSet_TagConfirmationOutput(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "002"
	setAddTags = []string{"feature"}
	setRemoveTags = []string{"security"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Updated task 002") {
		t.Error("Expected confirmation message with task ID")
	}
	if !strings.Contains(output, "tags: [backend, security] -> [backend, feature]") {
		t.Errorf("Expected formatted tag change, got: %s", output)
	}
}

func TestSet_Parent(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setParent = "002"

	// Mark the --parent flag as changed
	setCmd.Flags().Set("parent", "002")
	defer setCmd.Flags().Set("parent", "")

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "parent: (unset) -> 002") {
		t.Errorf("Expected parent change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "parent: 002") {
		t.Errorf("Expected file to contain parent: 002, got:\n%s", string(content))
	}
}

func TestSet_ParentClear(t *testing.T) {
	tmpDir := t.TempDir()

	content := `---
id: "030"
title: "Task with parent"
status: pending
parent: "001"
created: 2026-02-08
---

# Task with parent
`
	path := filepath.Join(tmpDir, "030-child.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "030"
	setParent = ""

	// Mark the --parent flag as changed (to clear)
	setCmd.Flags().Set("parent", "")
	defer setCmd.Flags().Set("parent", "")

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "parent: 001 ->") {
		t.Errorf("Expected parent change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(path)
	if strings.Contains(string(updated), "parent: 001") {
		t.Error("Expected parent to be cleared, but still found 'parent: 001'")
	}
}

func createVerifySetTestFile(t *testing.T, id, verifyYAML string) string {
	t.Helper()
	tmpDir := t.TempDir()

	content := fmt.Sprintf(`---
id: "%s"
title: "Task with verify"
status: pending
created: 2026-02-14
%s---

# Task with verify
`, id, verifyYAML)

	path := filepath.Join(tmpDir, id+"-verify.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return tmpDir
}

func TestSet_VerifyPassThenComplete(t *testing.T) {
	tmpDir := createVerifySetTestFile(t, "050", `verify:
  - type: bash
    run: "echo pass"
`)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "050"
	setStatus = "completed"
	setVerify = true

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("expected status change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "050-verify.md"))
	if !strings.Contains(string(content), "status: completed") {
		t.Error("Expected file to contain completed status")
	}
}

func TestSet_VerifyFailAborts(t *testing.T) {
	tmpDir := createVerifySetTestFile(t, "051", `verify:
  - type: bash
    run: "exit 1"
`)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "051"
	setStatus = "completed"
	setVerify = true

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("expected error when verify fails")
	}
	if !strings.Contains(err.Error(), "verification failed") {
		t.Errorf("expected 'verification failed' error, got: %v", err)
	}

	// Status should NOT be changed
	content, _ := os.ReadFile(filepath.Join(tmpDir, "051-verify.md"))
	if strings.Contains(string(content), "status: completed") {
		t.Error("Status should not be changed when verification fails")
	}
}

func TestSet_VerifyNoFieldProceeds(t *testing.T) {
	tmpDir := createVerifySetTestFile(t, "052", "")
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "052"
	setStatus = "completed"
	setVerify = true

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("expected status change, got: %s", output)
	}
}

func TestSet_VerifyNonCompletedSkips(t *testing.T) {
	tmpDir := createVerifySetTestFile(t, "053", `verify:
  - type: bash
    run: "exit 1"
`)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "053"
	setStatus = "in-progress"
	setVerify = true

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should succeed because --verify only gates completion
	if !strings.Contains(output, "status: pending -> in-progress") {
		t.Errorf("expected status change, got: %s", output)
	}
}

func TestSet_PositionalArg(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setStatus = "completed"

	output, err := captureSetOutputWithArgs(t, []string{"001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Updated task 001") {
		t.Error("Expected confirmation message")
	}
	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("Expected status change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "status: completed") {
		t.Error("Expected file to contain updated status")
	}
}

func TestSet_PositionalArgAndFlagSameValue(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setStatus = "completed"

	output, err := captureSetOutputWithArgs(t, []string{"001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Updated task 001") {
		t.Error("Expected confirmation message")
	}
}

func TestSet_PositionalArgAndFlagConflict(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "002"
	setStatus = "completed"

	_, err := captureSetOutputWithArgs(t, []string{"001"})
	if err == nil {
		t.Fatal("Expected error when positional arg and --task-id conflict")
	}
	if !strings.Contains(err.Error(), "conflicting task ID") {
		t.Errorf("Expected 'conflicting task ID' error, got: %v", err)
	}
}

func TestSet_NeitherPositionalNorFlag(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setStatus = "completed"

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("Expected error when neither positional arg nor --task-id provided")
	}
	if !strings.Contains(err.Error(), "task ID required") {
		t.Errorf("Expected 'task ID required' error, got: %v", err)
	}
}

func TestSet_InReviewStatus(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setStatus = "in-review"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status: pending -> in-review") {
		t.Errorf("Expected status change to in-review, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "status: in-review") {
		t.Error("Expected file to contain status: in-review")
	}
}

func TestSet_AddPR(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setAddPRs = []string{"https://github.com/example/repo/pull/1"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "pr:") {
		t.Errorf("Expected PR change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "https://github.com/example/repo/pull/1") {
		t.Errorf("Expected file to contain PR URL, got:\n%s", string(content))
	}
}

func TestSet_RemovePR(t *testing.T) {
	tmpDir := t.TempDir()
	content := `---
id: "040"
title: "Task with PR"
status: in-review
pr: ["https://github.com/example/repo/pull/1", "https://github.com/example/repo/pull/2"]
created: 2026-02-08
---

# Task with PR
`
	path := filepath.Join(tmpDir, "040-pr.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "040"
	setRemovePRs = []string{"https://github.com/example/repo/pull/1"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "pr:") {
		t.Errorf("Expected PR change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(path)
	fileStr := string(updated)
	if strings.Contains(fileStr, "pull/1") {
		t.Error("Expected PR 1 to be removed")
	}
	if !strings.Contains(fileStr, "pull/2") {
		t.Error("Expected PR 2 to be preserved")
	}
}

func TestSet_AddAndRemovePR(t *testing.T) {
	tmpDir := t.TempDir()
	content := `---
id: "041"
title: "Task with PR"
status: in-review
pr: ["https://github.com/example/repo/pull/1"]
created: 2026-02-08
---

# Task with PR
`
	path := filepath.Join(tmpDir, "041-pr.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "041"
	setAddPRs = []string{"https://github.com/example/repo/pull/2"}
	setRemovePRs = []string{"https://github.com/example/repo/pull/1"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "pr:") {
		t.Errorf("Expected PR change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(path)
	fileStr := string(updated)
	if strings.Contains(fileStr, "pull/1") {
		t.Error("Expected PR 1 to be removed")
	}
	if !strings.Contains(fileStr, "pull/2") {
		t.Error("Expected PR 2 to be added")
	}
}

func TestSet_DoneFlag_PRReviewWorkflow(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setDone = true

	// Simulate pr-review workflow via viper
	viper.Set("workflow", "pr-review")
	defer viper.Set("workflow", "")

	// Use positional arg to avoid flag state leakage
	output, err := captureSetOutputWithArgs(t, []string{"001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status: pending -> in-review") {
		t.Errorf("Expected --done to set status to in-review in pr-review workflow, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "status: in-review") {
		t.Error("Expected file to contain in-review status")
	}
}

func TestSet_DependsOn(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "003"
	setDependsOn = "001,002"

	setCmd.Flags().Set("depends-on", "001,002")
	defer func() { setCmd.Flags().Lookup("depends-on").Changed = false }()

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "dependencies: [002] -> [001, 002]") {
		t.Errorf("Expected dependencies change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "003-ui.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, `dependencies: ["001", "002"]`) {
		t.Errorf("Expected file to contain updated dependencies, got:\n%s", fileStr)
	}
}

func TestSet_DependsOn_InvalidID(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setDependsOn = "999"

	setCmd.Flags().Set("depends-on", "999")
	defer func() { setCmd.Flags().Lookup("depends-on").Changed = false }()

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("Expected error for non-existent dependency ID")
	}
	if !strings.Contains(err.Error(), `dependency "999" not found`) {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestSet_DependsOn_CircularDep(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	// 001 has no deps, 002 depends on 001, 003 depends on 002.
	// Setting 001 to depend on 003 creates: 001->003->002->001 (cycle).
	setTaskID = "001"
	setDependsOn = "003"

	setCmd.Flags().Set("depends-on", "003")
	defer func() { setCmd.Flags().Lookup("depends-on").Changed = false }()

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("Expected error for circular dependency")
	}
	if !strings.Contains(err.Error(), "circular dependency detected") {
		t.Errorf("Expected 'circular dependency' error, got: %v", err)
	}
}

func TestSet_DependsOn_SelfDep(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setDependsOn = "001"

	setCmd.Flags().Set("depends-on", "001")
	defer func() { setCmd.Flags().Lookup("depends-on").Changed = false }()

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("Expected error for self-dependency")
	}
	if !strings.Contains(err.Error(), "cannot depend on itself") {
		t.Errorf("Expected 'cannot depend on itself' error, got: %v", err)
	}
}

func TestSet_DependsOn_WithOtherFlags(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "003"
	setStatus = "in-progress"
	setDependsOn = "001"

	setCmd.Flags().Set("depends-on", "001")
	defer func() { setCmd.Flags().Lookup("depends-on").Changed = false }()

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status: blocked -> in-progress") {
		t.Error("Expected status change in output")
	}
	if !strings.Contains(output, "dependencies: [002] -> [001]") {
		t.Errorf("Expected dependencies change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "003-ui.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "status: in-progress") {
		t.Error("Expected file to contain updated status")
	}
	if !strings.Contains(fileStr, `dependencies: ["001"]`) {
		t.Errorf("Expected file to contain updated dependencies, got:\n%s", fileStr)
	}
}

func TestSet_DependsOn_Clear(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "002"
	setDependsOn = ""

	setCmd.Flags().Set("depends-on", "")
	defer func() { setCmd.Flags().Lookup("depends-on").Changed = false }()

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "dependencies: [001] -> []") {
		t.Errorf("Expected dependencies cleared in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	fileStr := string(content)
	// When clearing, the dependencies line should be removed
	if strings.Contains(fileStr, "dependencies:") {
		t.Errorf("Expected dependencies line to be removed, got:\n%s", fileStr)
	}
}

func TestComputeNewTags(t *testing.T) {
	tests := []struct {
		name       string
		current    []string
		addTags    []string
		removeTags []string
		want       []string
	}{
		{
			name:    "add to empty",
			current: nil,
			addTags: []string{"a", "b"},
			want:    []string{"a", "b"},
		},
		{
			name:       "remove from list",
			current:    []string{"a", "b", "c"},
			removeTags: []string{"b"},
			want:       []string{"a", "c"},
		},
		{
			name:       "add and remove",
			current:    []string{"a", "b"},
			addTags:    []string{"c"},
			removeTags: []string{"a"},
			want:       []string{"b", "c"},
		},
		{
			name:    "add duplicate is no-op",
			current: []string{"a", "b"},
			addTags: []string{"a"},
			want:    []string{"a", "b"},
		},
		{
			name:       "remove nonexistent is no-op",
			current:    []string{"a"},
			removeTags: []string{"z"},
			want:       []string{"a"},
		},
		{
			name:       "remove all tags",
			current:    []string{"a"},
			removeTags: []string{"a"},
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := taskfile.ComputeNewTags(tt.current, tt.addTags, tt.removeTags)
			if len(got) != len(tt.want) {
				t.Fatalf("computeNewTags() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("computeNewTags()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSet_AddTouches(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setAddTouches = []string{"cli/graph", "cli/output"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "touches:") {
		t.Errorf("Expected touches change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "cli/graph") {
		t.Errorf("Expected file to contain cli/graph, got:\n%s", fileStr)
	}
	if !strings.Contains(fileStr, "cli/output") {
		t.Errorf("Expected file to contain cli/output, got:\n%s", fileStr)
	}
}

func TestSet_RemoveTouches(t *testing.T) {
	tmpDir := t.TempDir()
	content := `---
id: "050"
title: "Task with touches"
status: pending
touches: ["cli/graph", "cli/output", "web/board"]
created: 2026-02-08
---

# Task with touches
`
	path := filepath.Join(tmpDir, "050-touches.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "050"
	setRemoveTouches = []string{"cli/graph"}

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "touches:") {
		t.Errorf("Expected touches change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(path)
	fileStr := string(updated)
	if strings.Contains(fileStr, "cli/graph") {
		t.Error("Expected cli/graph to be removed")
	}
	if !strings.Contains(fileStr, "cli/output") {
		t.Error("Expected cli/output to be preserved")
	}
	if !strings.Contains(fileStr, "web/board") {
		t.Error("Expected web/board to be preserved")
	}
}

func TestSet_AddTouches_Deduplication(t *testing.T) {
	tmpDir := t.TempDir()
	content := `---
id: "051"
title: "Task with existing touches"
status: pending
touches: ["cli/graph"]
created: 2026-02-08
---

# Task with existing touches
`
	path := filepath.Join(tmpDir, "051-touches-dedup.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "051"
	setAddTouches = []string{"cli/graph", "cli/output"}

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := os.ReadFile(path)
	fileStr := string(updated)
	// cli/graph should appear exactly once (not duplicated)
	count := strings.Count(fileStr, "cli/graph")
	if count != 1 {
		t.Errorf("Expected cli/graph to appear once, appeared %d times in:\n%s", count, fileStr)
	}
	if !strings.Contains(fileStr, "cli/output") {
		t.Error("Expected cli/output to be added")
	}
}

func TestSet_RemoveTouches_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	content := `---
id: "052"
title: "Task with touches"
status: pending
touches: ["cli/graph"]
created: 2026-02-08
---

# Task with touches
`
	path := filepath.Join(tmpDir, "052-touches.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "052"
	setRemoveTouches = []string{"nonexistent/scope"}

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error removing non-existent touches value: %v", err)
	}

	updated, _ := os.ReadFile(path)
	fileStr := string(updated)
	if !strings.Contains(fileStr, "cli/graph") {
		t.Error("Expected cli/graph to be preserved")
	}
}

func TestSet_Touches_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	content := `---
id: "053"
title: "Task for dry run"
status: pending
touches: ["cli/graph"]
created: 2026-02-08
---

# Task for dry run
`
	path := filepath.Join(tmpDir, "053-touches-dry.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "053"
	setAddTouches = []string{"cli/output"}
	setDryRun = true

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Dry run") {
		t.Errorf("Expected dry run message, got: %s", output)
	}

	// File should be unchanged
	updated, _ := os.ReadFile(path)
	fileStr := string(updated)
	if strings.Contains(fileStr, "cli/output") {
		t.Error("Expected file to be unchanged in dry run mode")
	}
}

func TestSet_AddAndRemoveTouches(t *testing.T) {
	tmpDir := t.TempDir()
	content := `---
id: "054"
title: "Task with touches"
status: pending
touches: ["cli/graph", "cli/set"]
created: 2026-02-08
---

# Task with touches
`
	path := filepath.Join(tmpDir, "054-touches.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "054"
	setAddTouches = []string{"cli/output"}
	setRemoveTouches = []string{"cli/graph"}

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := os.ReadFile(path)
	fileStr := string(updated)
	if strings.Contains(fileStr, "cli/graph") {
		t.Error("Expected cli/graph to be removed")
	}
	if !strings.Contains(fileStr, "cli/set") {
		t.Error("Expected cli/set to be preserved")
	}
	if !strings.Contains(fileStr, "cli/output") {
		t.Error("Expected cli/output to be added")
	}
}

func TestSet_AddTouches_EmptyArray(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setAddTouches = []string{"cli/graph"}

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "touches:") {
		t.Error("Expected touches field to be added")
	}
	if !strings.Contains(fileStr, "cli/graph") {
		t.Error("Expected cli/graph to be added")
	}
}

func TestSet_Phase_Add(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setPhase = "v0.2"

	setCmd.Flags().Set("phase", "v0.2")
	defer func() { setCmd.Flags().Lookup("phase").Changed = false }()

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "phase:") {
		t.Errorf("Expected phase change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "phase: v0.2") {
		t.Errorf("Expected file to contain phase: v0.2, got:\n%s", fileStr)
	}
}

func TestSet_Phase_Change(t *testing.T) {
	tmpDir := t.TempDir()
	content := `---
id: "060"
title: "Task with phase"
status: pending
phase: v0.1
created: 2026-02-08
---

# Task with phase
`
	path := filepath.Join(tmpDir, "060-phase.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "060"
	setPhase = "v0.2"

	setCmd.Flags().Set("phase", "v0.2")
	defer func() { setCmd.Flags().Lookup("phase").Changed = false }()

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "phase: v0.1 -> v0.2") {
		t.Errorf("Expected phase change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(path)
	fileStr := string(updated)
	if !strings.Contains(fileStr, "phase: v0.2") {
		t.Errorf("Expected phase to be updated, got:\n%s", fileStr)
	}
}

func TestSet_Phase_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	content := `---
id: "061"
title: "Task with phase"
status: pending
phase: v0.1
created: 2026-02-08
---

# Task with phase
`
	path := filepath.Join(tmpDir, "061-phase.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "061"
	setPhase = ""

	setCmd.Flags().Set("phase", "")
	defer func() { setCmd.Flags().Lookup("phase").Changed = false }()

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "phase:") {
		t.Errorf("Expected phase change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(path)
	fileStr := string(updated)
	if !strings.Contains(fileStr, "phase: ") || strings.Contains(fileStr, "phase: v0.1") {
		t.Errorf("Expected phase to be cleared, got:\n%s", fileStr)
	}
}

func TestSet_CompletedDate_AutoSetOnCompleted(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setStatus = "completed"

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	s := string(content)
	if !strings.Contains(s, "status: completed") {
		t.Error("Expected status to be completed")
	}
	if !strings.Contains(s, "completed_at: ") {
		t.Errorf("Expected completed_at date to be auto-set, got:\n%s", s)
	}
}

func TestSet_CancelledDate_AutoSetOnCancelled(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "001"
	setStatus = "cancelled"

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	s := string(content)
	if !strings.Contains(s, "status: cancelled") {
		t.Error("Expected status to be cancelled")
	}
	if !strings.Contains(s, "cancelled_at: ") {
		t.Errorf("Expected cancelled_at date to be auto-set, got:\n%s", s)
	}
	if strings.Contains(s, "completed_at:") {
		t.Error("Expected no completed_at field when cancelling")
	}
}

func TestSet_CompletedDate_ClearedOnReopen(t *testing.T) {
	tmpDir := t.TempDir()
	taskContent := `---
id: "070"
title: "Completed task"
status: completed
priority: medium
completed_at: 2026-03-01
created: 2026-02-08
---

# Completed task
`
	path := filepath.Join(tmpDir, "070-completed.md")
	if err := os.WriteFile(path, []byte(taskContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "070"
	setStatus = "pending"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "completed_at:") {
		t.Errorf("Expected completed_at change in output, got: %s", output)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, "status: pending") {
		t.Error("Expected status to be pending")
	}
	if strings.Contains(s, "completed_at:") {
		t.Errorf("Expected completed_at field to be removed, got:\n%s", s)
	}
}

func TestSet_CompletedDate_DoneFlag(t *testing.T) {
	tmpDir := createSetTestFiles(t)
	resetSetFlags()
	taskDir = tmpDir
	setDone = true

	_, err := captureSetOutputWithArgs(t, []string{"001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	s := string(content)
	if !strings.Contains(s, "status: completed") {
		t.Error("Expected status to be completed via --done")
	}
	if !strings.Contains(s, "completed_at: ") {
		t.Errorf("Expected completed_at date to be auto-set via --done, got:\n%s", s)
	}
}

func TestSet_CompletedDate_NonStatusChangePreserves(t *testing.T) {
	tmpDir := t.TempDir()
	taskContent := `---
id: "071"
title: "Completed task"
status: completed
priority: medium
completed_at: 2026-03-01
created: 2026-02-08
---

# Completed task
`
	path := filepath.Join(tmpDir, "071-completed.md")
	if err := os.WriteFile(path, []byte(taskContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "071"
	setPriority = "high"

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, "priority: high") {
		t.Error("Expected priority to be updated")
	}
	if !strings.Contains(s, "completed_at: 2026-03-01") {
		t.Errorf("Expected completed_at date to be preserved when not changing status, got:\n%s", s)
	}
}

func TestSet_CancelledDate_ClearedOnReopen(t *testing.T) {
	tmpDir := t.TempDir()
	taskContent := `---
id: "072"
title: "Cancelled task"
status: cancelled
priority: medium
cancelled_at: 2026-03-01
created: 2026-02-08
---

# Cancelled task
`
	path := filepath.Join(tmpDir, "072-cancelled.md")
	if err := os.WriteFile(path, []byte(taskContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "072"
	setStatus = "pending"

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "cancelled_at:") {
		t.Errorf("Expected cancelled_at change in output, got: %s", output)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, "status: pending") {
		t.Error("Expected status to be pending")
	}
	if strings.Contains(s, "cancelled_at:") {
		t.Errorf("Expected cancelled_at field to be removed, got:\n%s", s)
	}
}

func TestSet_CompletedDate_ClearsOnCancel(t *testing.T) {
	tmpDir := t.TempDir()
	taskContent := `---
id: "073"
title: "Completed then cancelled"
status: completed
priority: medium
completed_at: 2026-03-01
created: 2026-02-08
---

# Completed then cancelled
`
	path := filepath.Join(tmpDir, "073-task.md")
	if err := os.WriteFile(path, []byte(taskContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resetSetFlags()
	taskDir = tmpDir
	setTaskID = "073"
	setStatus = "cancelled"

	_, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, "cancelled_at: ") {
		t.Errorf("Expected cancelled_at to be set, got:\n%s", s)
	}
	if strings.Contains(s, "completed_at:") {
		t.Errorf("Expected completed_at to be cleared when cancelling, got:\n%s", s)
	}
}
