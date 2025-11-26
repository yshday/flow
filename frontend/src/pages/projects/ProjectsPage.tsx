import { useParams } from 'react-router-dom';
import AppLayout from '../../components/layout/AppLayout';
import IssueTreeView from '../../components/issues/IssueTreeView';

export default function ProjectsPage() {
  const { id } = useParams<{ id: string }>();

  return (
    <AppLayout>
      {id ? (
        <IssueTreeView projectId={parseInt(id)} />
      ) : (
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <p className="text-gray-500 dark:text-gray-400 text-lg mb-2">프로젝트를 선택하세요</p>
            <p className="text-gray-400 dark:text-gray-500 text-sm">
              왼쪽 사이드바에서 프로젝트를 선택하거나 새 프로젝트를 만들어보세요
            </p>
          </div>
        </div>
      )}
    </AppLayout>
  );
}
