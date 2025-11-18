import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Modal from '../common/Modal';
import { useCreateColumn, useUpdateColumn } from '../../hooks/useProjects';
import type { BoardColumn } from '../../types';

const columnSchema = z.object({
  name: z.string().min(1, '컬럼 이름을 입력해주세요').max(50, '컬럼 이름은 50자 이하여야 합니다'),
});

type ColumnFormData = z.infer<typeof columnSchema>;

interface ColumnModalProps {
  isOpen: boolean;
  onClose: () => void;
  projectId: number;
  column?: BoardColumn; // If provided, edit mode
  nextPosition: number; // For new columns
}

export default function ColumnModal({
  isOpen,
  onClose,
  projectId,
  column,
  nextPosition,
}: ColumnModalProps) {
  const isEditMode = !!column;
  const { mutateAsync: createColumn, isPending: isCreating, error: createError } = useCreateColumn(projectId);
  const { mutateAsync: updateColumn, isPending: isUpdating, error: updateError } = useUpdateColumn(
    column?.id || 0,
    projectId
  );

  const isPending = isCreating || isUpdating;
  const error = createError || updateError;

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<ColumnFormData>({
    resolver: zodResolver(columnSchema),
    defaultValues: {
      name: column?.name || '',
    },
  });

  const onSubmit = async (data: ColumnFormData) => {
    try {
      if (isEditMode) {
        await updateColumn({
          name: data.name,
          position: column.position,
        });
      } else {
        await createColumn({
          name: data.name,
          position: nextPosition,
        });
      }
      reset();
      onClose();
    } catch (error) {
      console.error('Failed to save column:', error);
    }
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title={isEditMode ? '컬럼 수정' : '새 컬럼 만들기'}>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded">
            컬럼 {isEditMode ? '수정' : '생성'}에 실패했습니다. 다시 시도해주세요.
          </div>
        )}

        <div>
          <label htmlFor="name" className="block text-sm font-medium text-gray-700">
            컬럼 이름 *
          </label>
          <input
            {...register('name')}
            type="text"
            id="name"
            className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            placeholder="예: To Do, In Review, QA"
          />
          {errors.name && (
            <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
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
            {isPending ? (isEditMode ? '수정 중...' : '생성 중...') : (isEditMode ? '컬럼 수정' : '컬럼 생성')}
          </button>
        </div>
      </form>
    </Modal>
  );
}
