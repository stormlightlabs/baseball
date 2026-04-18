import { apiFetch } from '$lib/api';
import type { ApiListPayload } from '$lib/players/types';
import { normalizeApiList } from '$lib/players/types';

export async function fetchList<T>(endpoint: string): Promise<T[]> {
  const payload = await apiFetch<ApiListPayload<T>>(endpoint);
  return normalizeApiList(payload);
}
