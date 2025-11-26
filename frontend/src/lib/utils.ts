import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleDateString('ko-KR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

export function formatDateTime(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleString('ko-KR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function formatRelativeTime(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return '방금 전';
  if (diffMins < 60) return `${diffMins}분 전`;
  if (diffHours < 24) return `${diffHours}시간 전`;
  if (diffDays < 7) return `${diffDays}일 전`;

  return formatDate(d);
}

export function getPriorityColor(priority: string): string {
  const colors = {
    low: 'text-gray-600 bg-gray-100',
    medium: 'text-blue-600 bg-blue-100',
    high: 'text-orange-600 bg-orange-100',
    urgent: 'text-red-600 bg-red-100',
  };
  return colors[priority as keyof typeof colors] || colors.medium;
}

export function getStatusColor(status: string): string {
  const colors = {
    open: 'text-green-600 bg-green-100',
    in_progress: 'text-yellow-600 bg-yellow-100',
    closed: 'text-gray-600 bg-gray-100',
  };
  return colors[status as keyof typeof colors] || colors.open;
}

export function getStatusText(status: string): string {
  const texts = {
    open: '열림',
    in_progress: '진행 중',
    closed: '닫힘',
  };
  return texts[status as keyof typeof texts] || '알 수 없음';
}

export function generateIssueKey(projectKey: string, issueNumber: number): string {
  return `${projectKey}-${issueNumber}`;
}

// Issue type utilities
export function getIssueTypeColor(issueType: string): string {
  const colors = {
    bug: 'text-red-600 bg-red-100',
    improvement: 'text-blue-600 bg-blue-100',
    epic: 'text-purple-600 bg-purple-100',
    feature: 'text-green-600 bg-green-100',
    task: 'text-gray-600 bg-gray-100',
    subtask: 'text-cyan-600 bg-cyan-100',
  };
  return colors[issueType as keyof typeof colors] || colors.task;
}

export function getIssueTypeLabel(issueType: string): string {
  const labels = {
    bug: '결함',
    improvement: '개선',
    epic: '에픽',
    feature: '신규 기능',
    task: '작업',
    subtask: '하위 작업',
  };
  return labels[issueType as keyof typeof labels] || '작업';
}

export function getIssueTypeIcon(issueType: string): string {
  const icons = {
    bug: '🐛',
    improvement: '⚡',
    epic: '🎯',
    feature: '✨',
    task: '📋',
    subtask: '📌',
  };
  return icons[issueType as keyof typeof icons] || '📋';
}
