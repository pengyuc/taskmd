import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { FieldGrid, MetadataFields } from "./TaskEditFormFields.tsx";
import { STATUSES, PRIORITIES, EFFORTS, TYPES } from "./TaskTable/constants.ts";

function getSelectByLabel(container: HTMLElement, label: string): HTMLSelectElement {
  const labelEl = Array.from(container.querySelectorAll("label")).find(
    (l) => l.textContent?.trim() === label,
  )!;
  return labelEl.parentElement!.querySelector("select")!;
}

function getInputByLabel(container: HTMLElement, label: string): HTMLInputElement {
  const labelEl = Array.from(container.querySelectorAll("label")).find(
    (l) => l.textContent?.trim() === label,
  )!;
  return labelEl.parentElement!.querySelector("input")!;
}

describe("FieldGrid", () => {
  function renderFieldGrid(overrides = {}) {
    const defaults = {
      status: "pending",
      onStatusChange: vi.fn(),
      priority: "medium",
      onPriorityChange: vi.fn(),
      effort: "small",
      onEffortChange: vi.fn(),
      taskType: "feature",
      onTaskTypeChange: vi.fn(),
      inputClasses: "test-input",
    };
    const props = { ...defaults, ...overrides };
    return { ...render(<FieldGrid {...props} />), props };
  }

  it("renders all four select fields with labels", () => {
    renderFieldGrid();
    expect(screen.getByText("Status")).toBeInTheDocument();
    expect(screen.getByText("Priority")).toBeInTheDocument();
    expect(screen.getByText("Effort")).toBeInTheDocument();
    expect(screen.getByText("Type")).toBeInTheDocument();
    expect(document.querySelectorAll("select")).toHaveLength(4);
  });

  it("renders all status options", () => {
    const { container } = renderFieldGrid();
    const select = getSelectByLabel(container, "Status");
    const options = Array.from(select.options).map((o) => o.value);
    expect(options).toEqual(STATUSES);
  });

  it("exposes all 6 spec statuses including in-review", () => {
    const { container } = renderFieldGrid();
    const select = getSelectByLabel(container, "Status");
    const options = Array.from(select.options).map((o) => o.value);
    expect(options).toEqual(
      expect.arrayContaining([
        "pending",
        "in-progress",
        "in-review",
        "completed",
        "blocked",
        "cancelled",
      ]),
    );
    expect(options).toHaveLength(6);
  });

  it("renders priority options with empty default", () => {
    const { container } = renderFieldGrid();
    const select = getSelectByLabel(container, "Priority");
    const options = Array.from(select.options).map((o) => o.value);
    expect(options).toEqual(["", ...PRIORITIES]);
    expect(select.options[0].textContent).toBe("-");
  });

  it("renders effort options with empty default", () => {
    const { container } = renderFieldGrid();
    const select = getSelectByLabel(container, "Effort");
    const options = Array.from(select.options).map((o) => o.value);
    expect(options).toEqual(["", ...EFFORTS]);
    expect(select.options[0].textContent).toBe("-");
  });

  it("renders type options with empty default", () => {
    const { container } = renderFieldGrid();
    const select = getSelectByLabel(container, "Type");
    const options = Array.from(select.options).map((o) => o.value);
    expect(options).toEqual(["", ...TYPES]);
    expect(select.options[0].textContent).toBe("-");
  });

  it("calls onStatusChange when status is changed", async () => {
    const { container, props } = renderFieldGrid();
    const select = getSelectByLabel(container, "Status");
    await userEvent.selectOptions(select, "completed");
    expect(props.onStatusChange).toHaveBeenCalledWith("completed");
  });

  it("calls onPriorityChange when priority is changed", async () => {
    const { container, props } = renderFieldGrid();
    const select = getSelectByLabel(container, "Priority");
    await userEvent.selectOptions(select, "high");
    expect(props.onPriorityChange).toHaveBeenCalledWith("high");
  });

  it("calls onEffortChange when effort is changed", async () => {
    const { container, props } = renderFieldGrid();
    const select = getSelectByLabel(container, "Effort");
    await userEvent.selectOptions(select, "large");
    expect(props.onEffortChange).toHaveBeenCalledWith("large");
  });

  it("calls onTaskTypeChange when type is changed", async () => {
    const { container, props } = renderFieldGrid();
    const select = getSelectByLabel(container, "Type");
    await userEvent.selectOptions(select, "bug");
    expect(props.onTaskTypeChange).toHaveBeenCalledWith("bug");
  });
});

describe("MetadataFields", () => {
  function renderMetadataFields(overrides = {}) {
    const defaults = {
      phase: "",
      onPhaseChange: vi.fn(),
      owner: "",
      onOwnerChange: vi.fn(),
      parent: "",
      onParentChange: vi.fn(),
      tags: "",
      onTagsChange: vi.fn(),
      inputClasses: "test-input",
    };
    const props = { ...defaults, ...overrides };
    return { ...render(<MetadataFields {...props} />), props };
  }

  it("renders all four input fields with labels", () => {
    renderMetadataFields();
    expect(screen.getByText("Phase")).toBeInTheDocument();
    expect(screen.getByText("Owner")).toBeInTheDocument();
    expect(screen.getByText("Parent")).toBeInTheDocument();
    expect(screen.getByText("Tags (comma-separated)")).toBeInTheDocument();
    expect(document.querySelectorAll("input")).toHaveLength(4);
  });

  it("renders correct placeholders", () => {
    renderMetadataFields();
    expect(screen.getByPlaceholderText("e.g. v1.0")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("e.g. alice")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("e.g. 045")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("e.g. backend, api, feature")).toBeInTheDocument();
  });

  it("calls onOwnerChange when owner input changes", async () => {
    const { container, props } = renderMetadataFields();
    const input = getInputByLabel(container, "Owner");
    await userEvent.type(input, "bob");
    expect(props.onOwnerChange).toHaveBeenCalled();
  });

  it("calls onParentChange when parent input changes", async () => {
    const { container, props } = renderMetadataFields();
    const input = getInputByLabel(container, "Parent");
    await userEvent.type(input, "042");
    expect(props.onParentChange).toHaveBeenCalled();
  });

  it("calls onTagsChange when tags input changes", async () => {
    const { props } = renderMetadataFields();
    await userEvent.type(screen.getByPlaceholderText("e.g. backend, api, feature"), "test");
    expect(props.onTagsChange).toHaveBeenCalled();
  });
});
