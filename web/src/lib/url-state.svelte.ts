/**
 * URL-driven state utilities for filter/sort/page/tab controls.
 *
 * Reading (reactive) — use page.url.searchParams directly in components:
 *
 *   import { page } from '$app/state';
 *   const q = $derived(page.url.searchParams.get('q') ?? '');
 *   const p = $derived(intParam(page.url.searchParams, 'page', 1));
 *
 * Writing — helpers below keep goto(resolve(...)) at the call site as required:
 *
 *   import { setUrlParam, setUrlParams } from '$lib/url-state';
 *   await setUrlParam('q', 'value');
 *   await setUrlParams({ page: '2', sort: 'hr' });
 */

import { goto } from '$app/navigation';
import { resolve } from '$app/paths';
import { page } from '$app/state';
import { SvelteURLSearchParams } from 'svelte/reactivity';

export type ParamValue = string | number | null | undefined;

type ParamOverrides = Record<string, ParamValue>;

const NAVIGATION_OPTS = { replaceState: true, noScroll: true, keepFocus: true } as const;
const currentRouteId = $derived(page.route.id);
const currentSearchParams = $derived(page.url.searchParams);
const currentHash = $derived(page.url.hash);

function withOverrides(base: URLSearchParams, overrides: ParamOverrides): SvelteURLSearchParams {
  const next = new SvelteURLSearchParams(base);
  for (const [key, value] of Object.entries(overrides)) {
    if (value === null || value === undefined || value === '') {
      next.delete(key);
    } else {
      next.set(key, String(value));
    }
  }
  return next;
}

function nextSearch(overrides: ParamOverrides): string {
  return withOverrides(currentSearchParams, overrides).toString();
}

/** Sets a single URL search parameter, replacing the current history entry. */
export async function setUrlParam(key: string, value: ParamValue): Promise<void> {
  const qs = nextSearch({ [key]: value });
  const routeId = currentRouteId;
  if (!routeId) return;
  let href = resolve(routeId);
  if (qs) href += `?${qs}`;
  if (currentHash) href += currentHash;
  await goto(href, NAVIGATION_OPTS);
}

/** Sets multiple URL search parameters atomically. */
export async function setUrlParams(params: ParamOverrides): Promise<void> {
  const qs = nextSearch(params);
  const routeId = currentRouteId;
  if (!routeId) return;
  let href = resolve(routeId);
  if (qs) href += `?${qs}`;
  if (currentHash) href += currentHash;
  await goto(href, NAVIGATION_OPTS);
}

/**
 * Returns a resolved href string merging overrides into the current search params.
 * Use for <a href={urlWith({page: 2})}> to avoid imperative navigation.
 */
export function urlWith(overrides: ParamOverrides): string {
  const qs = nextSearch(overrides);
  const routeId = currentRouteId;
  if (!routeId) return `${page.url.pathname}${page.url.search}${page.url.hash}`;
  let href = resolve(routeId);
  if (qs) href += `?${qs}`;
  if (currentHash) href += currentHash;
  return href;
}

/** Parses an integer from URLSearchParams, returning defaultValue when absent or invalid. */
export function intParam(params: URLSearchParams, key: string, defaultValue: number): number {
  const raw = params.get(key);
  if (raw === null) return defaultValue;
  const n = parseInt(raw, 10);
  return isNaN(n) ? defaultValue : n;
}
