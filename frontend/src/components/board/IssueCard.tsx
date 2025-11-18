import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import type { Issue } from '../../types';
import { generateIssueKey, getPriorityColor } from '../../lib/utils';

interface IssueCardProps {
  issue: Issue;
  projectKey: string;
  onClick?: () => void;
}

export default function IssueCard({ issue, projectKey, onClick }: IssueCardProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: issue.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      onClick={onClick}
      className="bg-white p-3 rounded-lg shadow-sm border border-gray-200 hover:shadow-md transition-shadow cursor-pointer"
    >
      {/* Issue Key */}
      <div className="text-xs font-medium text-blue-600 mb-1">
        {generateIssueKey(projectKey, issue.issue_number)}
      </div>

      {/* Title */}
      <h4 className="text-sm font-medium text-gray-900 mb-2">{issue.title}</h4>

      {/* Footer */}
      <div className="flex items-center justify-between">
        {/* Priority Badge */}
        <span
          className={`px-2 py-0.5 text-xs font-medium rounded ${getPriorityColor(
            issue.priority
          )}`}
        >
          {issue.priority}
        </span>

        {/* Labels */}
        {issue.labels && issue.labels.length > 0 && (
          <div className="flex gap-1">
            {issue.labels.slice(0, 2).map((label) => (
              <span
                key={label.id}
                className="px-2 py-0.5 text-xs rounded"
                style={{
                  backgroundColor: `${label.color}20`,
                  color: label.color,
                }}
              >
                {label.name}
              </span>
            ))}
            {issue.labels.length > 2 && (
              <span className="px-2 py-0.5 text-xs text-gray-500">
                +{issue.labels.length - 2}
              </span>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
