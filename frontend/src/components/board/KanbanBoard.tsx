import { useCallback, useMemo, useState } from 'react';
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  useDroppable,
} from '@dnd-kit/core';
import type {
  DragEndEvent,
  DragOverEvent,
  DragStartEvent,
} from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import type { BoardColumn, Issue } from '../../types';
import IssueCard from './IssueCard';
import ColumnModal from './ColumnModal';
import { useDeleteColumn } from '../../hooks/useProjects';

interface KanbanBoardProps {
  columns: BoardColumn[];
  issues: Issue[];
  projectKey: string;
  projectId: number;
  onIssueMove: (issueId: number, columnId: number, version: number) => void;
  onIssueClick?: (issueId: number) => void;
}

// Droppable column wrapper
function DroppableColumn({ column, children }: { column: BoardColumn; children: React.ReactNode }) {
  const { setNodeRef } = useDroppable({
    id: column.id,
  });

  return (
    <div ref={setNodeRef} className="flex-shrink-0 w-80 bg-gray-50 rounded-lg p-4">
      {children}
    </div>
  );
}

export default function KanbanBoard({
  columns,
  issues,
  projectKey,
  projectId,
  onIssueMove,
  onIssueClick,
}: KanbanBoardProps) {
  const [activeIssue, setActiveIssue] = useState<Issue | null>(null);
  const [isColumnModalOpen, setIsColumnModalOpen] = useState(false);
  const [editingColumn, setEditingColumn] = useState<BoardColumn | undefined>(undefined);

  const { mutateAsync: deleteColumn } = useDeleteColumn(projectId);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  );

  // Group issues by column
  const issuesByColumn = useMemo(() => {
    const grouped: Record<number, Issue[]> = {};
    columns.forEach((column) => {
      grouped[column.id] = issues.filter(
        (issue) => issue.column_id === column.id
      );
    });
    return grouped;
  }, [columns, issues]);

  const handleDragStart = useCallback((event: DragStartEvent) => {
    const { active } = event;
    const issue = issues.find((i) => i.id === active.id);
    if (issue) {
      setActiveIssue(issue);
    }
  }, [issues]);

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;
      setActiveIssue(null);

      if (!over) return;

      const activeIssue = issues.find((i) => i.id === active.id);
      if (!activeIssue) return;

      // Find the target column
      let targetColumnId: number | null = null;

      // Check if dropped directly on a column
      const overColumn = columns.find((c) => c.id === over.id);
      if (overColumn) {
        targetColumnId = overColumn.id;
      } else {
        // Check if dropped on another issue
        const overIssue = issues.find((i) => i.id === over.id);
        if (overIssue) {
          targetColumnId = overIssue.column_id;
        }
      }

      if (targetColumnId && targetColumnId !== activeIssue.column_id) {
        onIssueMove(activeIssue.id, targetColumnId, activeIssue.version);
      }
    },
    [issues, columns, onIssueMove]
  );

  const handleAddColumn = useCallback(() => {
    setEditingColumn(undefined);
    setIsColumnModalOpen(true);
  }, []);

  const handleEditColumn = useCallback((column: BoardColumn) => {
    setEditingColumn(column);
    setIsColumnModalOpen(true);
  }, []);

  const handleDeleteColumn = useCallback(async (columnId: number) => {
    if (window.confirm('이 컬럼을 삭제하시겠습니까? 컬럼 내의 모든 이슈는 유지됩니다.')) {
      try {
        await deleteColumn(columnId);
      } catch (error) {
        console.error('Failed to delete column:', error);
        alert('컬럼 삭제에 실패했습니다.');
      }
    }
  }, [deleteColumn]);

  const handleCloseModal = useCallback(() => {
    setIsColumnModalOpen(false);
    setEditingColumn(undefined);
  }, []);

  return (
    <DndContext
      sensors={sensors}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="flex gap-4 overflow-x-auto pb-4">
        {columns.map((column) => (
          <DroppableColumn key={column.id} column={column}>
            {/* Column Header */}
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <h3 className="font-semibold text-gray-900">{column.name}</h3>
                <span className="text-sm text-gray-500">
                  {issuesByColumn[column.id]?.length || 0}
                </span>
              </div>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => handleEditColumn(column)}
                  className="p-1 text-gray-500 hover:text-gray-700 rounded hover:bg-gray-200"
                  title="컬럼 수정"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </button>
                <button
                  onClick={() => handleDeleteColumn(column.id)}
                  className="p-1 text-gray-500 hover:text-red-600 rounded hover:bg-gray-200"
                  title="컬럼 삭제"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>

            {/* Issue List */}
            <SortableContext
              items={issuesByColumn[column.id]?.map((i) => i.id) || []}
              strategy={verticalListSortingStrategy}
              id={column.id.toString()}
            >
              <div className="space-y-2 min-h-[200px]">
                {issuesByColumn[column.id]?.map((issue) => (
                  <IssueCard
                    key={issue.id}
                    issue={issue}
                    projectKey={projectKey}
                    onClick={() => onIssueClick?.(issue.id)}
                  />
                ))}
              </div>
            </SortableContext>
          </DroppableColumn>
        ))}

        {/* Add Column Button */}
        <div className="flex-shrink-0 w-80">
          <button
            onClick={handleAddColumn}
            className="w-full h-full min-h-[200px] bg-gray-100 border-2 border-dashed border-gray-300 rounded-lg p-4 text-gray-500 hover:text-gray-700 hover:border-gray-400 hover:bg-gray-50 transition-colors flex flex-col items-center justify-center gap-2"
          >
            <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            <span className="text-sm font-medium">새 컬럼 추가</span>
          </button>
        </div>
      </div>

      {/* Drag Overlay */}
      <DragOverlay>
        {activeIssue ? (
          <IssueCard issue={activeIssue} projectKey={projectKey} />
        ) : null}
      </DragOverlay>

      {/* Column Management Modal */}
      <ColumnModal
        isOpen={isColumnModalOpen}
        onClose={handleCloseModal}
        projectId={projectId}
        column={editingColumn}
        nextPosition={columns.length}
      />
    </DndContext>
  );
}
