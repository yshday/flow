import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Modal from '../common/Modal';
import ColorPicker from '../common/ColorPicker';
import { useCreateLabel, useUpdateLabel } from '../../hooks/useIssues';
import type { Label } from '../../types';
import { toast } from '../../stores/toastStore';

const labelSchema = z.object({
  name: z.string().min(1, '라벨 이름을 입력해주세요').max(50, '라벨 이름은 50자 이하여야 합니다'),
  color: z.string().regex(/^#[0-9a-fA-F]{6}$/, '올바른 색상 코드를 선택해주세요'),
  description: z.string().max(200, '설명은 200자 이하여야 합니다').optional(),
});

type LabelFormData = z.infer<typeof labelSchema>;

interface LabelModalProps {
  isOpen: boolean;
  onClose: () => void;
  projectId: number;
  label?: Label | null;
}

export default function LabelModal({ isOpen, onClose, projectId, label }: LabelModalProps) {
  const isEditing = !!label;
  const { mutateAsync: createLabel, isPending: isCreating } = useCreateLabel(projectId);
  const { mutateAsync: updateLabel, isPending: isUpdating } = useUpdateLabel(
    label?.id || 0,
    projectId
  );

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
    watch,
  } = useForm<LabelFormData>({
    resolver: zodResolver(labelSchema),
    defaultValues: {
      name: '',
      color: '#3b82f6', // blue-500
      description: '',
    },
  });

  const selectedColor = watch('color');

  useEffect(() => {
    if (label) {
      reset({
        name: label.name,
        color: label.color,
        description: label.description || '',
      });
    } else {
      reset({
        name: '',
        color: '#3b82f6',
        description: '',
      });
    }
  }, [label, reset]);

  const onSubmit = async (data: LabelFormData) => {
    try {
      if (isEditing) {
        await updateLabel(data);
        toast.success('라벨이 성공적으로 수정되었습니다.');
      } else {
        await createLabel(data);
        toast.success('라벨이 성공적으로 생성되었습니다.');
      }
      onClose();
      reset();
    } catch (error) {
      console.error('Failed to save label:', error);
      toast.error(isEditing ? '라벨 수정에 실패했습니다.' : '라벨 생성에 실패했습니다.');
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={isEditing ? '라벨 수정' : '새 라벨'}>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {/* Name */}
        <div>
          <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
            이름 <span className="text-red-500">*</span>
          </label>
          <input
            id="name"
            type="text"
            {...register('name')}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="예: 버그, 기능 요청, 문서"
          />
          {errors.name && <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>}
        </div>

        {/* Color */}
        <div>
          <ColorPicker
            value={selectedColor}
            onChange={(color) => setValue('color', color)}
            label="색상"
          />
          {errors.color && <p className="mt-1 text-sm text-red-600">{errors.color.message}</p>}
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
            placeholder="라벨에 대한 설명을 입력하세요..."
          />
          {errors.description && (
            <p className="mt-1 text-sm text-red-600">{errors.description.message}</p>
          )}
        </div>

        {/* Preview */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">미리보기</label>
          <div className="flex items-center gap-2">
            <span
              className="px-3 py-1 text-sm font-medium rounded"
              style={{
                backgroundColor: selectedColor + '20',
                color: selectedColor,
                border: `1px solid ${selectedColor}`,
              }}
            >
              {watch('name') || '라벨 이름'}
            </span>
          </div>
        </div>

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
