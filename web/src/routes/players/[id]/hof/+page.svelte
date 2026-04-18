<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import Chart from '$lib/components/Chart.svelte';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import type { HofEntry } from '$lib/players/types';
  import { createHallOfFameChartConfig } from '$lib/players/charts';
  import { normalizeHallOfFamePayload } from '$lib/players/normalizers';
  import { AsyncListResource } from '$lib/players/resources.svelte';
  import { onMount } from 'svelte';

  let playerId = $derived(page.params.id ?? '');

  const hofResource = new AsyncListResource<HofEntry>();
  let lastKey = '';

  let hofChartConfig = $derived(createHallOfFameChartConfig(hofResource.items));

  async function refresh(force = false): Promise<void> {
    const id = playerId;
    if (!force && id === lastKey) return;
    lastKey = id;

    await hofResource.load(async () => {
      const payload = await apiFetch<{ records?: HofEntry[] | null; data?: HofEntry[] }>(EP.playerHallOfFame(id));
      return normalizeHallOfFamePayload(payload);
    });
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-3">Hall of Fame voting history</div>
  {#if hofResource.loading}
    <p class="font-mono text-xs text-muted">Loading…</p>
  {:else if hofResource.items.length === 0}
    <p class="font-mono text-xs text-muted">No Hall of Fame data on record.</p>
  {:else}
    {#if hofResource.items.some((entry) => entry.pct != null)}
      <div class="mb-4">
        <Chart config={hofChartConfig} height={100} />
      </div>
    {/if}
    <SortableTable
      columns={[
        { key: 'year_inducted', label: 'Year', sortable: true },
        { key: 'inducted', label: 'Inducted', format: (v) => (v ? 'Yes' : 'No') },
        { key: 'votes', label: 'Votes', sortable: true },
        { key: 'ballots', label: 'Ballots', sortable: true },
        { key: 'pct', label: 'Vote %', sortable: true, format: (v) => (v != null ? Number(v).toFixed(1) + '%' : '—') },
        { key: 'category', label: 'Category' }
      ]}
      rows={hofResource.items} />
  {/if}
</div>
