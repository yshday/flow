import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Modal from '../common/Modal';
import { useCreateMilestone, useUpdateMilestone } from '../../hooks/useMilestones';
import type { Milestone } from '../../types';
import { toast } from '../../stores/toastStore';

const milestoneSchema = z.object({
  title: z.string().min(1, '마일스톤 제목을 입력해주세요').max(100, '제목은 100자 이하여야 합니다'),
  description: z.string().max(500, '설명은 500자 이하여야 합니다').optional(),
  due_date: z.string().optional(),
  status: z.enum(['open', 'closed']).default('open'),
});

type MilestoneFormData = z.infer<typeof milestoneSchema>;

interface MilestoneModalProps {
  isOpen: boolean;
  onClose: () => void;
  projectId: number;
  milestone?: Milestone | null;
}

export default function MilestoneModal({
  isOpen,
  onClose,
  projectId,
  milestone,
}: MilestoneModalProps) {
  const isEditing = !!milestone;
  const { mutateAsync: createMilestone, isPending: isCreating } = useCreateMilestone(projectId);
  const { mutateAsync: updateMilestone, isPending: isUpdating } = useUpdateMilestone(
    milestone?.id || 0,
    projectId
  );

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<MilestoneFormData>({
    resolver: zodResolver(milestoneSchema),
    defaultValues: {
      title: '',
      description: '',
      due_date: '',
      status: 'open',
    },
  });

  useEffect(() => {
    if (milestone) {
      reset({
        title: milestone.title,
        description: milestone.description || '',
        due_date: milestone.due_date ? milestone.due_date.split('T')[0] : '',
        status: milestone.status,
      });
    } else {
      reset({
        title: '',
        description: '',
        due_date: '',
        status: 'open',
      });
    }
  }, [milestone, reset]);

  const onSubmit = async (data: MilestoneFormData) => {
    try {
      const payload = {
        title: data.title,
        description: data.description,
        due_date: data.due_date ? new Date(data.due_date).toISOString() : undefined,
        status: data.status,
      };

      if (isEditing) {
        await updateMilestone(payload);
        toast.success('마일스톤이 성공적으로 수정되었습니다.');
      } else {
        await createMilestone(payload);
        toast.success('마일스톤이 성공적으로 생성되었습니다.');
      }
      onClose();
      reset();
    } catch (error) {
      console.error('Failed to save milestone:', error);
      toast.error(
        isEditing ? '마일스톤 수정에 실패했습니다.' : '마일스톤 생성에 실패했습니다.'
      );
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditing ? '마일스톤 수정' : '새 마일스톤'}
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {/* Title */}
        <div>
          <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-1">
            제목 <span className="text-red-500">*</span>
          </label>
          <input
            id="title"
            type="text"
            {...register('title')}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="예: v1.0 릴리즈, 베타 버전"
          />
          {errors.title && <p className="mt-1 text-sm text-red-600">{errors.title.message}</p>}
        </div>

        {/* Description */}
        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
            설명 (선택)
          </label>
          <textarea
            id="description"
            {...register('description')}
            rows={3}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="마일스톤에 대한 설명을 입력하세요..."
          />
          {errors.description && (
            <p className="mt-1 text-sm text-red-600">{errors.description.message}</p>
          )}
        </div>

        {/* Due Date */}
        <div>
          <label htmlFor="due_date" className="block text-sm font-medium text-gray-700 mb-1">
            마감일 (선택)
          </label>
          <input
            id="due_date"
            type="date"
            {...register('due_date')}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          {errors.due_date && (
            <p className="mt-1 text-sm text-red-600">{errors.due_date.message}</p>
          )}
        </div>

        {/* Status */}
        {isEditing && (
          <div>
            <label htmlFor="status" className="block text-sm font-medium text-gray-700 mb-1">
              상태
            </label>
            <select
              id="status"
              {...register('status')}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="open">진행 중</option>
              <option value="closed">완료</option>
            </select>
          </div>
        )}

        {/* Actions */}
        <div className="flex justify-end space-x-3 pt-4 border-t">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
            disabled={isCreating || isUpdating}
          >
            취소
          </button>
          <button
            type="submit"
            disabled={isCreating || isUpdating}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isCreating || isUpdating
              ? isEditing
                ? '수정 중...'
                : '생성 중...'
              : isEditing
              ? '수정'
              : '생성'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
