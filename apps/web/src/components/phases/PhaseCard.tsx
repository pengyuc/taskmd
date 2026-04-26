import { useNavigate } from "react-router-dom";
import { PhaseInfo } from "../../hooks/use-config.ts";
import { PhaseProgressBar, PhaseProgress } from "../stats/PhaseProgressBar.tsx";

interface PhaseStats {
  total: number;
  completed: number;
  inProgress: number;
  inReview: number;
  pending: number;
  blocked: number;
}

interface PhaseCardProps {
  phase: PhaseInfo;
  stats: PhaseStats;
}

function statusBadge(label: string, count: number, color: string) {
  if (count === 0) return null;
  return (
    <span className={`inline-flex items-center px-2 py-0.5 text-xs font-medium rounded-full ${color}`}>
      {label}: {count}
    </span>
  );
}

export function PhaseCard({ phase, stats }: PhaseCardProps) {
  const navigate = useNavigate();

  const progress: PhaseProgress = {
    id: phase.id,
    name: phase.name,
    total: stats.total,
    completed: stats.completed,
  };

  return (
    <button
      type="button"
      onClick={() => navigate(`/tasks?phase=${phase.id}`)}
      className="w-full text-left cursor-pointer bg-white rounded-lg border border-gray-200 p-4 hover:border-gray-300 hover:shadow-sm transition-all dark:bg-gray-800 dark:border-gray-700 dark:hover:border-gray-600"
    >
      <div className="mb-3">
        <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100">{phase.name}</h3>
        {phase.description && (
          <p className="mt-1 text-xs text-gray-500 dark:text-gray-400 line-clamp-2">{phase.description}</p>
        )}
      </div>

      <PhaseProgressBar phase={progress} />

      <div className="mt-3 flex flex-wrap gap-1.5">
        {statusBadge("Completed", stats.completed, "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400")}
        {statusBadge("In Progress", stats.inProgress, "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400")}
        {statusBadge("In Review", stats.inReview, "bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400")}
        {statusBadge("Pending", stats.pending, "bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400")}
        {statusBadge("Blocked", stats.blocked, "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400")}
      </div>
    </button>
  );
}
