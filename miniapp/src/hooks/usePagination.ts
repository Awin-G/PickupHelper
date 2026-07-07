import { useState, useCallback } from 'react';
import type { PaginatedList } from '@/api/types';

interface UsePaginationOptions<T> {
  fetchFn: (params: { page: number; page_size: number }) => Promise<PaginatedList<T>>;
  pageSize?: number;
}

export function usePagination<T>({ fetchFn, pageSize = 20 }: UsePaginationOptions<T>) {
  const [list, setList] = useState<T[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetchFn({ page: 1, page_size: pageSize });
      setList(res.list);
      setPage(1);
      setHasMore(res.list.length >= pageSize);
    } finally {
      setLoading(false);
    }
  }, [fetchFn, pageSize]);

  const loadMore = useCallback(async () => {
    if (!hasMore || loading) return;
    setLoading(true);
    try {
      const nextPage = page + 1;
      const res = await fetchFn({ page: nextPage, page_size: pageSize });
      setList((prev) => [...prev, ...res.list]);
      setPage(nextPage);
      setHasMore(res.list.length >= pageSize);
    } finally {
      setLoading(false);
    }
  }, [fetchFn, hasMore, loading, page, pageSize]);

  return { list, loading, hasMore, refresh, loadMore };
}
