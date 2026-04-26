import { describe, it, expect } from "vitest";
import { findMatchedNodeIds, filterGraphByStatus } from "./graph-utils.ts";
import type { GraphData } from "../../api/types.ts";

const sampleData: GraphData = {
  nodes: [
    { id: "001", title: "Setup project", status: "completed" },
    { id: "002", title: "Add auth", status: "pending" },
    { id: "003", title: "Write tests", status: "in-progress" },
  ],
  edges: [
    { from: "001", to: "002" },
    { from: "002", to: "003" },
  ],
  cycles: [],
};

describe("findMatchedNodeIds", () => {
  it("returns empty set for empty query", () => {
    const result = findMatchedNodeIds(sampleData, "");
    expect(result.size).toBe(0);
  });

  it("matches by title (case insensitive)", () => {
    const result = findMatchedNodeIds(sampleData, "auth");
    expect(result).toEqual(new Set(["002"]));
  });

  it("matches by ID", () => {
    const result = findMatchedNodeIds(sampleData, "003");
    expect(result).toEqual(new Set(["003"]));
  });

  it("matches multiple nodes", () => {
    // "t" matches "Setup project", "Add auth" (no), "Write tests"
    const result = findMatchedNodeIds(sampleData, "test");
    expect(result).toEqual(new Set(["003"]));
  });

  it("returns empty set when nothing matches", () => {
    const result = findMatchedNodeIds(sampleData, "zzz");
    expect(result.size).toBe(0);
  });
});

describe("filterGraphByStatus", () => {
  it("returns original data when no statuses are selected", () => {
    const result = filterGraphByStatus(sampleData, new Set());
    expect(result).toBe(sampleData);
  });

  it("filters nodes by status", () => {
    const result = filterGraphByStatus(sampleData, new Set(["pending"]));
    expect(result.nodes).toHaveLength(1);
    expect(result.nodes[0].id).toBe("002");
  });

  it("keeps only edges where both endpoints are visible", () => {
    const result = filterGraphByStatus(sampleData, new Set(["completed", "pending"]));
    expect(result.nodes).toHaveLength(2);
    // Edge 001→002 should be kept, edge 002→003 should be dropped
    expect(result.edges).toHaveLength(1);
    expect(result.edges[0]).toEqual({ from: "001", to: "002" });
  });

  it("preserves cycles from original data", () => {
    const dataWithCycles: GraphData = { ...sampleData, cycles: [["001", "002"]] };
    const result = filterGraphByStatus(dataWithCycles, new Set(["pending"]));
    expect(result.cycles).toEqual([["001", "002"]]);
  });

  it("filters by in-review status", () => {
    const dataWithReview: GraphData = {
      nodes: [
        ...sampleData.nodes,
        { id: "004", title: "Review PR", status: "in-review" },
      ],
      edges: [
        ...sampleData.edges,
        { from: "003", to: "004" },
      ],
      cycles: [],
    };
    const result = filterGraphByStatus(dataWithReview, new Set(["in-review"]));
    expect(result.nodes).toHaveLength(1);
    expect(result.nodes[0].id).toBe("004");
    expect(result.edges).toHaveLength(0);
  });
});
