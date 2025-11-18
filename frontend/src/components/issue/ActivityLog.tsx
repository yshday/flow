import { useActivities } from '../../hooks/useIssues';
import type { Activity } from '../../types';

interface ActivityLogProps {
  issueId: number;
}

export default function ActivityLog({ issueId }: ActivityLogProps) {
  const { data: activities, isLoading } = useActivities(issueId, { limit: 50 });

  const formatActivityMessage = (activity: Activity): string => {
    const { action, entity_type, field_name, old_value, new_value } = activity;

    switch (action) {
      case 'created':
        return `이슈를 생성했습니다`;
      case 'updated':
        if (field_name) {
          const fieldNames: Record<string, string> = {
            title: '제목',
            description: '설명',
            status: '상태',
            priority: '우선순위',
            assignee_id: '담당자',
            milestone_id: '마일스톤',
            column_id: '컬럼',
          };
          const displayName = fieldNames[field_name] || field_name;

          if (old_value && new_value) {
            return `${displayName}을(를) "${old_value}"에서 "${new_value}"(으)로 변경했습니다`;
          }
          return `${displayName}을(를) 업데이트했습니다`;
        }
        return '이슈를 업데이트했습니다';
      case 'moved':
        return `이슈를 이동했습니다`;
      case 'added':
        if (entity_type === 'label') {
          return `라벨을 추가했습니다`;
        }
        return `${entity_type}을(를) 추가했습니다`;
      case 'removed':
        if (entity_type === 'label') {
          return `라벨을 제거했습니다`;
        }
        return `${entity_type}을(를) 제거했습니다`;
      case 'deleted':
        return `${entity_type}을(를) 삭제했습니다`;
      default:
        return `${action} ${entity_type}`;
    }
  };

  const formatTime = (dateString: string): string => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return '방금 전';
    if (diffMins < 60) return `${diffMins}분 전`;
    if (diffHours < 24) return `${diffHours}시간 전`;
    if (diffDays < 7) return `${diffDays}일 전`;

    return new Date(dateString).toLocaleDateString('ko-KR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (isLoading) {
    return (
      <div className="py-4 text-center text-gray-500">
        활동 로그를 불러오는 중...
      </div>
    );
  }

  if (!activities || activities.length === 0) {
    return (
      <div className="py-4 text-center text-gray-500">
        활동 기록이 없습니다.
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-gray-900">활동</h3>
      <div className="space-y-3">
        {activities.map((activity) => (
          <div key={activity.id} className="flex gap-3 text-sm">
            {/* User Avatar */}
            <div className="flex-shrink-0">
              <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white font-medium">
                {activity.user?.username?.[0]?.toUpperCase() ||
                 activity.user?.email?.[0]?.toUpperCase() ||
                 '?'}
              </div>
            </div>

            {/* Activity Content */}
            <div className="flex-1 min-w-0">
              <div className="flex items-baseline gap-2">
                <span className="font-medium text-gray-900">
                  {activity.user?.username || activity.user?.email || '알 수 없음'}
                </span>
                <span className="text-gray-600">
                  {formatActivityMessage(activity)}
                </span>
              </div>
              <div className="text-gray-500 text-xs mt-1">
                {formatTime(activity.created_at)}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
