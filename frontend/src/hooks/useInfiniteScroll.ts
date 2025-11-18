import { useEffect, useRef } from 'react';

interface UseInfiniteScrollOptions {
  /** Callback to load more data */
  onLoadMore: () => void;
  /** Whether there is more data to load */
  hasNextPage: boolean;
  /** Whether data is currently being loaded */
  isLoading: boolean;
  /** Root margin for intersection observer (default: '100px') */
  rootMargin?: string;
  /** Threshold for intersection observer (default: 0.1) */
  threshold?: number;
}

/**
 * Custom hook for infinite scrolling using IntersectionObserver
 * @param options Configuration options
 * @returns Ref to attach to the sentinel element
 */
export function useInfiniteScroll({
  onLoadMore,
  hasNextPage,
  isLoading,
  rootMargin = '100px',
  threshold = 0.1,
}: UseInfiniteScrollOptions) {
  const sentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const [entry] = entries;

        // Load more when sentinel is visible and we have more pages and not currently loading
        if (entry.isIntersecting && hasNextPage && !isLoading) {
          onLoadMore();
        }
      },
      {
        rootMargin,
        threshold,
      }
    );

    observer.observe(sentinel);

    // Cleanup
    return () => {
      observer.disconnect();
    };
  }, [onLoadMore, hasNextPage, isLoading, rootMargin, threshold]);

  return sentinelRef;
}
