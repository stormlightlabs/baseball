export type QueryValue = string | number | boolean | null | undefined;

export type QueryOverrides = Record<string, QueryValue>;

export const QUERY_NAV_OPTS = { replaceState: true, noScroll: true, keepFocus: true } as const;

function mergeQuery(base: URLSearchParams, overrides: QueryOverrides): URLSearchParams {
  const next = new URLSearchParams(base);
  for (const [key, value] of Object.entries(overrides)) {
    if (value === null || value === undefined || value === '') {
      next.delete(key);
      continue;
    }
    next.set(key, String(value));
  }
  return next;
}

export function intParam(params: URLSearchParams, key: string, defaultValue: number): number {
  const raw = params.get(key);
  if (raw === null) return defaultValue;
  const n = Number.parseInt(raw, 10);
  return Number.isNaN(n) ? defaultValue : n;
}

export function withMergedQuery(
  baseHref: string,
  base: URLSearchParams,
  overrides: QueryOverrides = {},
  hash = ''
): string {
  const qs = mergeQuery(base, overrides).toString();
  let href = baseHref;
  if (qs) href += `?${qs}`;
  if (hash) href += hash;
  return href;
}
