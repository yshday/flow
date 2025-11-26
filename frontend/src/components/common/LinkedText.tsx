import React, { useState, useRef } from 'react';
import { Link } from 'react-router-dom';
import { parseIssueLinks, type ParsedIssueLink } from '../../lib/issueLinkParser';
import { useIssueByNumber } from '../../hooks/useIssues';
import { getStatusColor, getPriorityColor } from '../../lib/utils';

interface LinkedTextProps {
  text: string;
  projectId: number;
  projectKey: string;
  className?: string;
}

interface IssueLinkProps {
  link: ParsedIssueLink;
  projectId: number;
  projectKey: string;
}

const IssueLink: React.FC<IssueLinkProps> = ({ link, projectId, projectKey }) => {
  const [showTooltip, setShowTooltip] = useState(false);
  const linkRef = useRef<HTMLAnchorElement>(null);

  // Determine if it's a reference to the current project
  const isCurrentProject = !link.projectKey || link.projectKey === projectKey;

  // Only fetch if it's the current project (we don't have project key to ID mapping yet)
  const { data: issue, isLoading } = useIssueByNumber(
    isCurrentProject ? projectId : 0,
    link.issueNumber
  );

  const handleMouseEnter = () => setShowTooltip(true);
  const handleMouseLeave = () => setShowTooltip(false);

  // If we have an issue, link to it; otherwise show the text without link
  if (!isCurrentProject) {
    // For cross-project links, just show styled text (future: resolve project key)
    return (
      <span className="text-blue-600 cursor-not-allowed opacity-70" title="Cross-project links not yet supported">
        {link.text}
      </span>
    );
  }

  if (!issue && !isLoading) {
    // Issue not found
    return (
      <span className="text-gray-500 line-through" title="Issue not found">
        {link.text}
      </span>
    );
  }

  return (
    <span className="relative inline-block">
      <Link
        ref={linkRef}
        to={issue ? `/projects/${projectId}/issues/${issue.id}` : '#'}
        className={`text-blue-600 hover:text-blue-800 hover:underline font-medium ${
          isLoading ? 'opacity-50' : ''
        }`}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
      >
        {link.text}
      </Link>

      {/* Tooltip */}
      {showTooltip && issue && (
        <div className="absolute z-50 left-0 top-full mt-1 w-72 p-3 bg-white rounded-lg shadow-lg border border-gray-200">
          <div className="space-y-2">
            {/* Issue key and status */}
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-blue-600">
                {projectKey}-{issue.issue_number}
              </span>
              <span className={`px-1.5 py-0.5 text-xs rounded ${getStatusColor(issue.status)}`}>
                {issue.status === 'open' ? '열림' : issue.status === 'in_progress' ? '진행 중' : '닫힘'}
              </span>
              <span className={`px-1.5 py-0.5 text-xs rounded ${getPriorityColor(issue.priority)}`}>
                {issue.priority}
              </span>
            </div>

            {/* Title */}
            <p className="text-sm font-medium text-gray-900 line-clamp-2">
              {issue.title}
            </p>

            {/* Description preview */}
            {issue.description && (
              <p className="text-xs text-gray-500 line-clamp-2">
                {issue.description}
              </p>
            )}
          </div>
        </div>
      )}
    </span>
  );
};

/**
 * LinkedText Component
 *
 * Renders text with automatic issue link detection and conversion.
 * Issue references like #123 or TPP-123 become clickable links with tooltips.
 */
export const LinkedText: React.FC<LinkedTextProps> = ({
  text,
  projectId,
  projectKey,
  className = '',
}) => {
  const segments = parseIssueLinks(text);

  return (
    <span className={className}>
      {segments.map((segment, index) => {
        if (segment.type === 'text') {
          return <span key={index}>{segment.text}</span>;
        }

        return (
          <IssueLink
            key={index}
            link={segment}
            projectId={projectId}
            projectKey={projectKey}
          />
        );
      })}
    </span>
  );
};

export default LinkedText;
