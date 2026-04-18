<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { SALARIES_COLUMNS } from '$lib/common/constants';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { fetchList } from '$lib/players/fetchers';
  import { AsyncListResource } from '$lib/players/resources.svelte';
  import type { Salary } from '$lib/players/types';
  import { onMount } from 'svelte';

  let playerId = $derived(page.params.id ?? '');

  const salariesResource = new AsyncListResource<Salary>();
  let lastKey = '';

  async function refresh(force = false): Promise<void> {
    const id = playerId;
    if (!force && id === lastKey) return;
    lastKey = id;

    await salariesResource.load(() => fetchList<Salary>(EP.playerSalaries(id)));
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-3">Salaries</div>
  {#if salariesResource.loading}
    <p class="font-mono text-xs text-muted">Loading…</p>
  {:else if salariesResource.items.length === 0}
    <p class="font-mono text-[0.78rem] text-muted">No salary data on record.</p>
  {:else}
    <SortableTable
      columns={SALARIES_COLUMNS}
      rows={salariesResource.items.map((salary) => ({ ...salary, team: salary.team ?? salary.team_id ?? '?' }))} />
  {/if}
</div>
