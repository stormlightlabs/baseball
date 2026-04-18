import { apiFetch } from './api';
import type { MetaResponse } from './api';

// FIXME: getters should be $derived state to keep them reactive
class MetaStore {
  data = $state<MetaResponse | null>(null);
  loading = $state(false);
  error = $state<string | null>(null);

  async init() {
    if (this.data || this.loading) return;
    this.loading = true;
    this.error = null;
    try {
      this.data = await apiFetch<MetaResponse>('/meta');
    } catch (e) {
      this.error = e instanceof Error ? e.message : 'Failed to load API metadata';
    } finally {
      this.loading = false;
    }
  }

  get version(): string {
    return this.data?.version ?? '—';
  }

  get status(): 'online' | 'loading' | 'offline' {
    if (this.data) return 'online';
    if (this.loading) return 'loading';
    return 'offline';
  }

  get coverage(): MetaResponse['coverage'] {
    return this.data?.coverage ?? {};
  }

  get datasets(): MetaResponse['datasets'] {
    return this.data?.datasets ?? [];
  }

  get dataFrom(): string {
    const years = Object.values(this.data?.coverage ?? {})
      .map((c) => c.from)
      .filter((y): y is number => y !== undefined);
    return years.length ? String(Math.min(...years)) : '—';
  }

  get dataTo(): string {
    const years = Object.values(this.data?.coverage ?? {})
      .map((c) => c.to)
      .filter((y): y is number => y !== undefined);
    return years.length ? String(Math.max(...years)) : '—';
  }

  get generatedAt(): string {
    if (!this.data?.generated_at) return '—';
    return new Date(this.data.generated_at).toLocaleDateString();
  }
}

export const meta = new MetaStore();
