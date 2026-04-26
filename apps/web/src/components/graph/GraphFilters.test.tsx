import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { GraphFilters } from "./GraphFilters.tsx";

describe("GraphFilters", () => {
  it("renders Status label and all button", () => {
    render(
      <GraphFilters
        selectedStatuses={new Set()}
        onToggleStatus={() => {}}
        onClearFilters={() => {}}
      />,
    );

    expect(screen.getByText("Status:")).toBeInTheDocument();
    expect(screen.getByText("all")).toBeInTheDocument();
  });

  it("renders all status buttons", () => {
    render(
      <GraphFilters
        selectedStatuses={new Set()}
        onToggleStatus={() => {}}
        onClearFilters={() => {}}
      />,
    );

    expect(screen.getByText("pending")).toBeInTheDocument();
    expect(screen.getByText("in-progress")).toBeInTheDocument();
    expect(screen.getByText("in-review")).toBeInTheDocument();
    expect(screen.getByText("completed")).toBeInTheDocument();
    expect(screen.getByText("blocked")).toBeInTheDocument();
    expect(screen.getByText("cancelled")).toBeInTheDocument();
  });

  it("calls onToggleStatus when a status button is clicked", () => {
    const onToggleStatus = vi.fn();

    render(
      <GraphFilters
        selectedStatuses={new Set()}
        onToggleStatus={onToggleStatus}
        onClearFilters={() => {}}
      />,
    );

    fireEvent.click(screen.getByText("blocked"));

    expect(onToggleStatus).toHaveBeenCalledWith("blocked");
  });

  it("calls onClearFilters when all button is clicked", () => {
    const onClearFilters = vi.fn();

    render(
      <GraphFilters
        selectedStatuses={new Set(["pending"])}
        onToggleStatus={() => {}}
        onClearFilters={onClearFilters}
      />,
    );

    fireEvent.click(screen.getByText("all"));

    expect(onClearFilters).toHaveBeenCalledOnce();
  });
});
