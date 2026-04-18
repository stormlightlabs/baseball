<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { TEAMS_COLUMNS } from '$lib/common/constants';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { fetchList } from '$lib/players/fetchers';
  import { AsyncListResource } from '$lib/players/resources.svelte';
  import type { PlayerTeam } from '$lib/players/types';
  import { onMount } from 'svelte';

  let playerId = $derived(page.params.id ?? '');

  const teamsResource = new AsyncListResource<PlayerTeam>();
  let lastKey = '';

  async function refresh(force = false): Promise<void> {
    const id = playerId;
    if (!force && id === lastKey) return;
    lastKey = id;

    await teamsResource.load(() => fetchList<PlayerTeam>(EP.playerTeams(id)));
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-3">Teams</div>
  {#if teamsResource.loading}
    <p class="font-mono text-xs text-muted">Loading…</p>
  {:else if teamsResource.items.length === 0}
    <p class="font-mono text-[0.78rem] text-muted">No team data on record.</p>
  {:else}
    <SortableTable
      columns={TEAMS_COLUMNS}
      rows={teamsResource.items.map((team) => ({ ...team, team: team.team ?? team.team_id ?? '?' }))} />
  {/if}
</div>
