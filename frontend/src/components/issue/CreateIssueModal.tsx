import { memo } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Modal from '../common/Modal';
import { useCreateIssue } from '../../hooks/useIssues';
import { useBoardColumns } from '../../hooks/useProjects';

const issueSchema = z.object({
  title: z.string().min(1, '이슈 제목을 입력해주세요'),
  description: z.string().optional(),
  priority: z.enum(['low', 'medium', 'high', 'urgent']).default('medium'),
});

type IssueFormData = z.infer<typeof issueSchema>;

interface CreateIssueModalProps {
  isOpen: boolean;
  onClose: () => void;
  projectId: number;
  onSuccess?: () => void;
}

function CreateIssueModal({
  isOpen,
  onClose,
  projectId,
  onSuccess,
}: CreateIssueModalProps) {
  const { mutateAsync: createIssue, isPending, error } = useCreateIssue(projectId);
  const { data: columns } = useBoardColumns(projectId);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<IssueFormData>({
    resolver: zodResolver(issueSchema),
    defaultValues: {
      priority: 'medium',
    },
  });

  const onSubmit = async (data: IssueFormData) => {
    try {
      // Assign to first column (Backlog) by default
      const defaultColumnId = columns?.[0]?.id;

      await createIssue({
        title: data.title,
        description: data.description || '',
        priority: data.priority,
        column_id: defaultColumnId,
      });
      reset();
      onClose();
      onSuccess?.();
    } catch (error) {
      console.error('Failed to create issue:', error);
    }
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="새 이슈 만들기">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded">
            이슈 생성에 실패했습니다. 다시 시도해주세요.
          </div>
        )}

        <div>
          <label htmlFor="title" className="block text-sm font-medium text-gray-700">
            제목 *
          </label>
          <input
            {...register('title')}
            type="text"
            id="title"
            className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            placeholder="버그 수정, 기능 추가 등"
          />
          {errors.title && (
            <p className="mt-1 text-sm text-red-600">{errors.title.message}</p>
          )}
        </div>

        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700">
            설명
          </label>
          <textarea
            {...register('description')}
            id="description"
            rows={4}
            className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            placeholder="이슈에 대한 자세한 설명을 입력해주세요"
          />
          {errors.description && (
            <p className="mt-1 text-sm text-red-600">{errors.description.message}</p>
          )}
        </div>

        <div>
          <label htmlFor="priority" className="block text-sm font-medium text-gray-700">
            우선순위 *
          </label>
          <select
            {...register('priority')}
            id="priority"
            className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="low">낮음 (Low)</option>
            <option value="medium">보통 (Medium)</option>
            <option value="high">높음 (High)</option>
            <option value="urgent">긴급 (Urgent)</option>
          </select>
          {errors.priority && (
            <p className="mt-1 text-sm text-red-600">{errors.priority.message}</p>
          )}
        </div>

        <div className="flex justify-end space-x-3 pt-4">
          <button
            type="button"
            onClick={handleClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            취소
          </button>
          <button
            type="submit"
            disabled={isPending}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isPending ? '생성 중...' : '이슈 생성'}
          </button>
        </div>
      </form>
    </Modal>
  );
}

// Memoize to prevent unnecessary re-renders
export default memo(CreateIssueModal);
