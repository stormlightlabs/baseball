<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { apiFetch } from '$lib/api';
  import { AsyncListResource } from '$lib/players/resources.svelte';
  import { normalizeApiList, rowColumns, type ApiListPayload, type TableRow } from '$lib/players/types';
  import { onMount } from 'svelte';
  import SortableTable from './SortableTable.svelte';

  let { endpoint, label }: { endpoint: string; label: string } = $props();

  const rows = new AsyncListResource<TableRow>();
  let lastEndpoint = '';

  async function loadRows(ep: string): Promise<void> {
    await rows.load(
      async () => {
        const payload = await apiFetch<ApiListPayload<TableRow>>(ep);
        return normalizeApiList(payload);
      },
      { resetOnLoad: true }
    );
  }

  async function refresh(force = false): Promise<void> {
    const ep = endpoint;
    if (!ep) {
      rows.clear();
      lastEndpoint = '';
      return;
    }
    if (!force && ep === lastEndpoint) return;
    lastEndpoint = ep;
    await loadRows(ep);
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });

  let columns = $derived(rowColumns(rows.items));
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-3">{label}</div>
  {#if rows.loading}
    <p class="font-mono text-xs text-muted">Loading…</p>
  {:else if rows.error}
    <p class="font-mono text-xs text-warning">{rows.error}</p>
  {:else if rows.items.length === 0}
    <p class="font-mono text-xs text-muted">No data on record.</p>
  {:else}
    <SortableTable {columns} rows={rows.items} />
  {/if}
</div>
