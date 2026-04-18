<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { AsyncValueResource } from '$lib/players/resources.svelte';
  import type { TableRow } from '$lib/players/types';
  import { rowColumns } from '$lib/players/types';
  import { normalizeTeamFieldingStats } from '$lib/teams/normalizers';
  import type { TeamFieldingStats } from '$lib/teams/types';
  import { onMount } from 'svelte';

  let teamId = $derived(page.params.id ?? '');
  let year = $derived(page.url.searchParams.get('year') ?? '');

  const fieldingResource = new AsyncValueResource<TeamFieldingStats>();
  let lastKey = '';

  let playerRows = $derived((fieldingResource.value?.players ?? []) as TableRow[]);
  let columns = $derived(rowColumns(playerRows));

  async function refresh(force = false): Promise<void> {
    const id = teamId;
    const y = year;
    if (!y) {
      fieldingResource.clear();
      return;
    }

    const key = `${id}|${y}`;
    if (!force && key === lastKey) return;
    lastKey = key;

    await fieldingResource.load(async () => {
      const payload = await apiFetch<Record<string, unknown>>(EP.seasonTeamFielding(y, id), { players: true });
      return normalizeTeamFieldingStats(payload);
    });
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });

  const fmtRate = (value: number | undefined) => (value != null ? Number(value).toFixed(3) : '—');
</script>

{#if !year}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Enter a season year in the sidebar to view fielding stats.</p>
{:else if fieldingResource.loading}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Loading…</p>
{:else if fieldingResource.error}
  <p class="mt-4 font-mono text-[0.78rem] text-warning">{fieldingResource.error}</p>
{:else if fieldingResource.value}
  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">{year} Team Fielding</div>

    <div class="mb-4 grid grid-cols-4 gap-2 text-center">
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{fieldingResource.value.e ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">E</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{fieldingResource.value.dp ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">DP</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{fieldingResource.value.a ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">A</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{fmtRate(fieldingResource.value.fpct)}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">FPCT</div>
      </div>
    </div>

    {#if playerRows.length === 0}
      <p class="font-mono text-[0.78rem] text-muted">No player fielding rows found for {year}.</p>
    {:else}
      <SortableTable {columns} rows={playerRows} />
    {/if}
  </div>
{:else}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">No fielding data found for {year}.</p>
{/if}
