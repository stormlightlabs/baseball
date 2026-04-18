<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { AsyncListResource } from '$lib/players/resources.svelte';
  import type { TableRow } from '$lib/players/types';
  import { rowColumns } from '$lib/players/types';
  import { onMount } from 'svelte';

  let teamId = $derived(page.params.id ?? '');
  let year = $derived(page.url.searchParams.get('year') ?? '');

  const rosterResource = new AsyncListResource<TableRow>();
  let lastKey = '';

  let columns = $derived(rowColumns(rosterResource.items));

  async function refresh(force = false): Promise<void> {
    const id = teamId;
    const y = year;
    if (!y) {
      rosterResource.clear();
      return;
    }

    const key = `${id}|${y}`;
    if (!force && key === lastKey) return;
    lastKey = key;

    await rosterResource.load(async () => {
      const payload = await apiFetch<Record<string, unknown>[]>(EP.seasonTeamRoster(y, id));
      return Array.isArray(payload) ? payload : [];
    });
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });
</script>

{#if !year}
  <p class="mt-4 font-mono text-xs text-muted">Enter a season year in the sidebar to view the roster.</p>
{:else if rosterResource.loading}
  <p class="mt-4 font-mono text-xs text-muted">Loading…</p>
{:else if rosterResource.error}
  <p class="mt-4 font-mono text-xs text-warning">{rosterResource.error}</p>
{:else}
  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">{year} Roster</div>
    {#if rosterResource.items.length === 0}
      <p class="font-mono text-xs text-muted">No roster data found for {year}.</p>
    {:else}
      <SortableTable {columns} rows={rosterResource.items} />
    {/if}
  </div>
{/if}
