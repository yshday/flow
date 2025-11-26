import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Modal from './Modal';
import { useCreateProject } from '../../hooks/useProjects';
import { apiClient } from '../../api/client';
import type { ProjectTemplate } from '../../types';

const projectSchema = z.object({
  name: z.string().min(1, '프로젝트 이름을 입력해주세요'),
  key: z
    .string()
    .min(2, '프로젝트 키는 최소 2자 이상이어야 합니다')
    .max(10, '프로젝트 키는 최대 10자까지 가능합니다')
    .regex(/^[A-Z]+$/, '프로젝트 키는 대문자만 사용 가능합니다'),
  description: z.string().optional(),
  template_id: z.number().optional(),
});

type ProjectFormData = z.infer<typeof projectSchema>;

interface CreateProjectModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: () => void;
}

// 템플릿 아이콘 매핑
const templateIcons: Record<string, string> = {
  '칸반 기본': '📋',
  '스크럼 스프린트': '🏃',
  '버그 트래킹': '🐛',
  '기능 개발': '🚀',
};

export default function CreateProjectModal({
  isOpen,
  onClose,
  onSuccess,
}: CreateProjectModalProps) {
  const { mutateAsync: createProject, isPending, error } = useCreateProject();
  const [templates, setTemplates] = useState<ProjectTemplate[]>([]);
  const [loadingTemplates, setLoadingTemplates] = useState(false);
  const [selectedTemplateId, setSelectedTemplateId] = useState<number | undefined>(undefined);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
  } = useForm<ProjectFormData>({
    resolver: zodResolver(projectSchema),
  });

  // 템플릿 로드
  useEffect(() => {
    if (isOpen) {
      setLoadingTemplates(true);
      apiClient
        .get<ProjectTemplate[]>('/templates/projects')
        .then((response) => {
          const data = response.data;
          setTemplates(data);
          // 첫 번째 템플릿을 기본으로 선택
          if (data.length > 0) {
            setSelectedTemplateId(data[0].id);
            setValue('template_id', data[0].id);
          }
        })
        .catch((err) => {
          console.error('Failed to load templates:', err);
        })
        .finally(() => {
          setLoadingTemplates(false);
        });
    }
  }, [isOpen, setValue]);

  const handleTemplateSelect = (templateId: number) => {
    setSelectedTemplateId(templateId);
    setValue('template_id', templateId);
  };

  const onSubmit = async (data: ProjectFormData) => {
    try {
      await createProject({
        name: data.name,
        key: data.key,
        description: data.description,
        template_id: data.template_id,
      });
      reset();
      setSelectedTemplateId(undefined);
      onClose();
      onSuccess?.();
    } catch (error) {
      console.error('Failed to create project:', error);
    }
  };

  const handleClose = () => {
    reset();
    setSelectedTemplateId(undefined);
    onClose();
  };

  const selectedTemplate = templates.find((t) => t.id === selectedTemplateId);

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="새 프로젝트 만들기">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 px-4 py-3 rounded">
            프로젝트 생성에 실패했습니다. 프로젝트 키가 이미 사용 중일 수 있습니다.
          </div>
        )}

        {/* 템플릿 선택 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            템플릿 선택
          </label>
          {loadingTemplates ? (
            <div className="text-center py-4 text-gray-500 dark:text-gray-400">
              템플릿 로딩 중...
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-3">
              {templates.map((template) => (
                <button
                  key={template.id}
                  type="button"
                  onClick={() => handleTemplateSelect(template.id)}
                  className={`p-3 border rounded-lg text-left transition-all ${
                    selectedTemplateId === template.id
                      ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20 ring-2 ring-blue-500'
                      : 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500 hover:bg-gray-50 dark:hover:bg-gray-700'
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <span className="text-xl">
                      {templateIcons[template.name] || '📁'}
                    </span>
                    <div>
                      <div className="font-medium text-gray-900 dark:text-gray-100">
                        {template.name}
                      </div>
                      {template.description && (
                        <div className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                          {template.description}
                        </div>
                      )}
                    </div>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* 선택된 템플릿 상세 정보 */}
        {selectedTemplate && (
          <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-3 text-sm">
            <div className="flex flex-wrap gap-2">
              <span className="text-gray-500 dark:text-gray-400">컬럼:</span>
              {selectedTemplate.config.columns.map((col, idx) => (
                <span
                  key={idx}
                  className="px-2 py-0.5 bg-white dark:bg-gray-700 rounded text-xs border border-gray-200 dark:border-gray-600"
                >
                  {col.name}
                </span>
              ))}
            </div>
            {selectedTemplate.config.labels.length > 0 && (
              <div className="flex flex-wrap gap-2 mt-2">
                <span className="text-gray-500 dark:text-gray-400">라벨:</span>
                {selectedTemplate.config.labels.slice(0, 5).map((label, idx) => (
                  <span
                    key={idx}
                    className="px-2 py-0.5 rounded text-xs text-white"
                    style={{ backgroundColor: label.color }}
                  >
                    {label.name}
                  </span>
                ))}
                {selectedTemplate.config.labels.length > 5 && (
                  <span className="text-xs text-gray-500 dark:text-gray-400">
                    +{selectedTemplate.config.labels.length - 5}
                  </span>
                )}
              </div>
            )}
          </div>
        )}

        <div>
          <label htmlFor="name" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
            프로젝트 이름 *
          </label>
          <input
            {...register('name')}
            type="text"
            id="name"
            className="mt-1 block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            placeholder="My Awesome Project"
          />
          {errors.name && (
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.name.message}</p>
          )}
        </div>

        <div>
          <label htmlFor="key" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
            프로젝트 키 *
          </label>
          <input
            {...register('key')}
            type="text"
            id="key"
            className="mt-1 block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 uppercase"
            placeholder="MAP"
            maxLength={10}
          />
          <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
            대문자만 사용 가능 (예: PROJ, ITP)
          </p>
          {errors.key && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.key.message}</p>}
        </div>

        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
            설명
          </label>
          <textarea
            {...register('description')}
            id="description"
            rows={3}
            className="mt-1 block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            placeholder="프로젝트에 대한 간단한 설명을 입력해주세요"
          />
          {errors.description && (
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.description.message}</p>
          )}
        </div>

        <div className="flex justify-end space-x-3 pt-4">
          <button
            type="button"
            onClick={handleClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            취소
          </button>
          <button
            type="submit"
            disabled={isPending}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isPending ? '생성 중...' : '프로젝트 생성'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
