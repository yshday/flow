import { memo } from 'react';
import { getIssueTypeColor, getIssueTypeLabel, getIssueTypeIcon } from '../../lib/utils';
import type { IssueType } from '../../types';

interface IssueTypeBadgeProps {
  type: IssueType;
  showLabel?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

function IssueTypeBadge({ type, showLabel = true, size = 'sm' }: IssueTypeBadgeProps) {
  const colorClass = getIssueTypeColor(type);
  const label = getIssueTypeLabel(type);
  const icon = getIssueTypeIcon(type);

  const sizeClasses = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-2.5 py-1 text-sm',
    lg: 'px-3 py-1.5 text-base',
  };

  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full font-medium ${colorClass} ${sizeClasses[size]}`}
      title={label}
    >
      <span>{icon}</span>
      {showLabel && <span>{label}</span>}
    </span>
  );
}

export default memo(IssueTypeBadge);
