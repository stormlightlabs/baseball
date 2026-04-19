/** Base URL: set VITE_API_BASE_URL in .env.production; defaults to /v1 for local proxy. */
const BASE_RAW: string = (import.meta.env.VITE_API_BASE_URL as string) || '/v1';
const BASE = BASE_RAW.endsWith('/') ? BASE_RAW.slice(0, -1) : BASE_RAW;

export type Params = Record<string, string | number | boolean | null | undefined>;

type PaginatedParams = Params & { page?: number; per_page?: number };

export type PaginatedResponse<T> = { data: T[]; page: number; per_page: number; total: number };

export type DatasetCoverage = { from?: number; to?: number };

export type DatasetStatus = {
  id: string;
  name: string;
  source: string;
  required?: boolean;
  healthy?: boolean;
  coverage_from?: number;
  coverage_to?: number;
  last_loaded_at?: string;
  row_count: number;
  tables?: Record<string, number>;
};

export type MetaResponse = {
  version: string;
  generated_at: string;
  coverage: Record<string, DatasetCoverage>;
  schema_hashes: Record<string, string>;
  datasets: DatasetStatus[];
};

function buildUrl(path: string, params?: Params): string {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  const origin = globalThis.window !== undefined ? globalThis.location.origin : 'http://localhost:8080';
  const base = BASE.startsWith('http') ? BASE : origin + BASE;
  const url = new URL(base + normalizedPath);
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      if (v !== undefined && v !== null && v !== '') {
        url.searchParams.set(k, String(v));
      }
    }
  }
  return url.toString();
}

export async function apiFetch<T>(path: string, params?: Params): Promise<T> {
  const url = buildUrl(path, params);
  const res = await fetch(url, { credentials: 'include' });
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error((body as { error: string }).error ?? `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export function fetchPaginated<T>(path: string, params?: PaginatedParams): Promise<PaginatedResponse<T>> {
  return apiFetch<PaginatedResponse<T>>(path, params);
}

/** Returns the resolved URL for a given API path + params (useful for ApiPanel/ApiMirrorStrip). */
export function apiUrl(path: string, params?: Params): string {
  return buildUrl(path, params);
}

/** Returns a URL/href for non-fetch links (OAuth/docs) without forcing an origin. */
export function apiHref(path: string): string {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return `${BASE}${normalizedPath}`;
}
