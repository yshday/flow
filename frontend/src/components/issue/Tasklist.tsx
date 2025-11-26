import React, { useState, useRef, useEffect } from 'react';
import {
  useTasklistItems,
  useTasklistProgress,
  useCreateTasklistItem,
  useToggleTasklistItem,
  useDeleteTasklistItem,
  useUpdateTasklistItem,
} from '../../hooks/useTasklist';
import type { TasklistItem } from '../../types';

interface TasklistProps {
  issueId: number;
}

export const Tasklist: React.FC<TasklistProps> = ({ issueId }) => {
  const { data: items = [], isLoading } = useTasklistItems(issueId);
  const { data: progress } = useTasklistProgress(issueId);
  const createItem = useCreateTasklistItem(issueId);
  const toggleItem = useToggleTasklistItem(issueId);
  const deleteItem = useDeleteTasklistItem(issueId);
  const updateItem = useUpdateTasklistItem(issueId);

  const [newItemContent, setNewItemContent] = useState('');
  const [isAdding, setIsAdding] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editContent, setEditContent] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);
  const editInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isAdding && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isAdding]);

  useEffect(() => {
    if (editingId && editInputRef.current) {
      editInputRef.current.focus();
    }
  }, [editingId]);

  const handleAddItem = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newItemContent.trim()) return;

    try {
      await createItem.mutateAsync({ content: newItemContent.trim() });
      setNewItemContent('');
      // Keep the input focused for quick consecutive adds
      inputRef.current?.focus();
    } catch (error) {
      console.error('Failed to add item:', error);
    }
  };

  const handleToggle = async (item: TasklistItem) => {
    try {
      await toggleItem.mutateAsync(item.id);
    } catch (error) {
      console.error('Failed to toggle item:', error);
    }
  };

  const handleDelete = async (itemId: number) => {
    try {
      await deleteItem.mutateAsync(itemId);
    } catch (error) {
      console.error('Failed to delete item:', error);
    }
  };

  const handleStartEdit = (item: TasklistItem) => {
    setEditingId(item.id);
    setEditContent(item.content);
  };

  const handleSaveEdit = async () => {
    if (!editingId || !editContent.trim()) {
      setEditingId(null);
      return;
    }

    try {
      await updateItem.mutateAsync({
        itemId: editingId,
        data: { content: editContent.trim() },
      });
      setEditingId(null);
    } catch (error) {
      console.error('Failed to update item:', error);
    }
  };

  const handleCancelEdit = () => {
    setEditingId(null);
    setEditContent('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      handleCancelEdit();
    } else if (e.key === 'Enter') {
      handleSaveEdit();
    }
  };

  if (isLoading) {
    return (
      <div className="animate-pulse space-y-2">
        <div className="h-4 bg-gray-200 rounded w-1/4"></div>
        <div className="h-6 bg-gray-200 rounded w-full"></div>
        <div className="h-6 bg-gray-200 rounded w-full"></div>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {/* Header with progress */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-gray-700 flex items-center gap-2">
          <svg
            className="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
            />
          </svg>
          Tasklist
          {progress && progress.total > 0 && (
            <span className="text-gray-500 text-xs">
              ({progress.completed}/{progress.total})
            </span>
          )}
        </h3>
        {!isAdding && (
          <button
            onClick={() => setIsAdding(true)}
            className="text-sm text-blue-600 hover:text-blue-800 flex items-center gap-1"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Add item
          </button>
        )}
      </div>

      {/* Progress bar */}
      {progress && progress.total > 0 && (
        <div className="w-full bg-gray-200 rounded-full h-1.5">
          <div
            className="bg-green-500 h-1.5 rounded-full transition-all duration-300"
            style={{ width: `${progress.percent}%` }}
          ></div>
        </div>
      )}

      {/* Tasklist items */}
      <div className="space-y-1">
        {items.map((item) => (
          <div
            key={item.id}
            className={`group flex items-center gap-2 p-2 rounded hover:bg-gray-50 ${
              item.is_completed ? 'opacity-70' : ''
            }`}
          >
            {/* Checkbox */}
            <button
              onClick={() => handleToggle(item)}
              className={`flex-shrink-0 w-5 h-5 rounded border-2 flex items-center justify-center transition-colors ${
                item.is_completed
                  ? 'bg-green-500 border-green-500 text-white'
                  : 'border-gray-300 hover:border-gray-400'
              }`}
            >
              {item.is_completed && (
                <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                    clipRule="evenodd"
                  />
                </svg>
              )}
            </button>

            {/* Content */}
            {editingId === item.id ? (
              <input
                ref={editInputRef}
                type="text"
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                onBlur={handleSaveEdit}
                onKeyDown={handleKeyDown}
                className="flex-1 text-sm border border-blue-300 rounded px-2 py-1 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            ) : (
              <span
                className={`flex-1 text-sm cursor-pointer ${
                  item.is_completed ? 'line-through text-gray-500' : 'text-gray-800'
                }`}
                onClick={() => handleStartEdit(item)}
              >
                {item.content}
              </span>
            )}

            {/* Actions */}
            <div className="flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity flex gap-1">
              <button
                onClick={() => handleDelete(item.id)}
                className="p-1 text-gray-400 hover:text-red-500"
                title="Delete"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Add new item form */}
      {isAdding && (
        <form onSubmit={handleAddItem} className="flex items-center gap-2">
          <div className="flex-shrink-0 w-5 h-5 rounded border-2 border-gray-300"></div>
          <input
            ref={inputRef}
            type="text"
            value={newItemContent}
            onChange={(e) => setNewItemContent(e.target.value)}
            onBlur={() => {
              if (!newItemContent.trim()) {
                setIsAdding(false);
              }
            }}
            placeholder="Add a task..."
            className="flex-1 text-sm border border-gray-300 rounded px-2 py-1 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="submit"
            disabled={!newItemContent.trim() || createItem.isPending}
            className="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {createItem.isPending ? 'Adding...' : 'Add'}
          </button>
          <button
            type="button"
            onClick={() => {
              setIsAdding(false);
              setNewItemContent('');
            }}
            className="px-3 py-1 text-sm text-gray-600 hover:text-gray-800"
          >
            Cancel
          </button>
        </form>
      )}

      {/* Empty state */}
      {items.length === 0 && !isAdding && (
        <div className="text-center py-4">
          <p className="text-sm text-gray-500">No tasks yet</p>
          <button
            onClick={() => setIsAdding(true)}
            className="mt-2 text-sm text-blue-600 hover:text-blue-800"
          >
            Add your first task
          </button>
        </div>
      )}
    </div>
  );
};

export default Tasklist;
