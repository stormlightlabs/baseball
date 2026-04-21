import { afterEach, describe, expect, it, vi } from 'vitest';

import { apiFetch, fetchPaginated } from './api';

describe('api helpers', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('builds URL query params and sends first-party web headers', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue({ ok: true, json: async () => ({ data: [], page: 2, per_page: 20, total: 0 }) });
    vi.stubGlobal('fetch', fetchMock);

    await fetchPaginated('/players', { q: 'judge', page: 2, per_page: 20 });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, options] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toContain('/v1/players');
    expect(url).toContain('q=judge');
    expect(url).toContain('page=2');
    expect(url).toContain('per_page=20');
    expect(options).toMatchObject({ credentials: 'include', headers: { 'X-BigFly-Client': 'web' } });
  });

  it('throws API-provided JSON error messages', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue({
          ok: false,
          status: 429,
          statusText: 'Too Many Requests',
          json: async () => ({ error: 'Rate limit exceeded' })
        })
    );

    await expect(apiFetch('/meta')).rejects.toThrow('Rate limit exceeded');
  });

  it('falls back to HTTP status text when error payload is not JSON', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 503,
        statusText: 'Service Unavailable',
        json: async () => {
          throw new Error('not json');
        }
      })
    );

    await expect(apiFetch('/meta')).rejects.toThrow('Service Unavailable');
  });
});
