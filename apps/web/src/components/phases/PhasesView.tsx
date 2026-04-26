import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { PhaseInfo } from "../../hooks/use-config.ts";
import { Task } from "../../api/types.ts";
import { PhaseCard } from "./PhaseCard.tsx";

interface PhasesViewProps {
  phases: PhaseInfo[];
  tasks: Task[];
}

function computeStats(tasks: Task[]) {
  return {
    total: tasks.length,
    completed: tasks.filter((t) => t.status === "completed").length,
    inProgress: tasks.filter((t) => t.status === "in-progress").length,
    inReview: tasks.filter((t) => t.status === "in-review").length,
    pending: tasks.filter((t) => t.status === "pending" || t.status === "open" || t.status === "").length,
    blocked: tasks.filter((t) => t.status === "blocked").length,
  };
}

export function PhasesView({ phases, tasks }: PhasesViewProps) {
  const navigate = useNavigate();

  const { phaseStats, unphasedStats } = useMemo(() => {
    const statsMap = new Map<string, Task[]>();
    const unphased: Task[] = [];

    for (const task of tasks) {
      if (task.phase) {
        const list = statsMap.get(task.phase) ?? [];
        list.push(task);
        statsMap.set(task.phase, list);
      } else {
        unphased.push(task);
      }
    }

    return {
      phaseStats: phases.map((p) => ({
        phase: p,
        stats: computeStats(statsMap.get(p.id) ?? []),
      })),
      unphasedStats: computeStats(unphased),
    };
  }, [phases, tasks]);

  if (phases.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-sm text-gray-500 dark:text-gray-400">
          No phases configured. Add phases to your{" "}
          <code className="text-xs bg-gray-100 dark:bg-gray-700 px-1.5 py-0.5 rounded">.taskmd.yaml</code>{" "}
          to use this view.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {phaseStats.map(({ phase, stats }) => (
          <PhaseCard key={phase.id} phase={phase} stats={stats} />
        ))}
      </div>

      {unphasedStats.total > 0 && (
        <div className="bg-white rounded-lg border border-gray-200 p-4 dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200">Unphased Tasks</h3>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                {unphasedStats.total} task{unphasedStats.total !== 1 ? "s" : ""} not assigned to any phase
              </p>
            </div>
            <button
              type="button"
              onClick={() => navigate("/tasks")}
              className="text-xs text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
            >
              View all →
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
