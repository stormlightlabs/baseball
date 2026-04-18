import { apiFetch } from './api';
import type { MetaResponse } from './api';

class MetaStore {
  data = $state<MetaResponse | null>(null);
  loading = $state(false);
  error = $state<string | null>(null);

  version = $derived(this.data?.version ?? '—');
  coverage = $derived(this.data?.coverage ?? {});
  datasets = $derived(this.data?.datasets ?? []);
  requiredDatasets = $derived(this.datasets.filter((dataset) => dataset.required !== false));
  healthyDatasetCount = $derived(this.datasets.filter((dataset) => dataset.healthy !== false).length);
  requiredHealthyCount = $derived(this.requiredDatasets.filter((dataset) => dataset.healthy !== false).length);
  requiredUnhealthyCount = $derived(this.requiredDatasets.filter((dataset) => dataset.healthy === false).length);

  status = $derived.by((): 'online' | 'loading' | 'degraded' | 'offline' => {
    if (this.loading) return 'loading';
    if (!this.data) return 'offline';
    return this.requiredUnhealthyCount > 0 ? 'degraded' : 'online';
  });

  dataFromYear = $derived.by(() => {
    const years = Object.values(this.coverage)
      .map((c) => c.from)
      .filter((y): y is number => y !== undefined);
    return years.length > 0 ? Math.min(...years) : null;
  });

  dataToYear = $derived.by(() => {
    const years = Object.values(this.coverage)
      .map((c) => c.to)
      .filter((y): y is number => y !== undefined);
    return years.length > 0 ? Math.max(...years) : null;
  });

  dataFrom = $derived(this.dataFromYear != null ? String(this.dataFromYear) : '—');
  dataTo = $derived(this.dataToYear != null ? String(this.dataToYear) : '—');
  generatedAt = $derived.by(() => {
    if (!this.data?.generated_at) return '—';
    return new Date(this.data.generated_at).toLocaleString();
  });

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
}

export const meta = new MetaStore();
