<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { AsyncValueResource } from '$lib/players/resources.svelte';
  import type { TableRow } from '$lib/players/types';
  import { rowColumns } from '$lib/players/types';
  import { normalizeTeamBattingStats } from '$lib/teams/normalizers';
  import type { TeamBattingStats } from '$lib/teams/types';
  import { onMount } from 'svelte';

  let teamId = $derived(page.params.id ?? '');
  let year = $derived(page.url.searchParams.get('year') ?? '');

  const battingResource = new AsyncValueResource<TeamBattingStats>();
  let lastKey = '';

  let playerRows = $derived((battingResource.value?.players ?? []) as TableRow[]);
  let columns = $derived(rowColumns(playerRows));

  async function refresh(force = false): Promise<void> {
    const id = teamId;
    const y = year;
    if (!y) {
      battingResource.clear();
      return;
    }

    const key = `${id}|${y}`;
    if (!force && key === lastKey) return;
    lastKey = key;

    await battingResource.load(async () => {
      const payload = await apiFetch<Record<string, unknown>>(EP.seasonTeamBatting(y, id), { players: true });
      return normalizeTeamBattingStats(payload);
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
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Enter a season year in the sidebar to view batting stats.</p>
{:else if battingResource.loading}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Loading…</p>
{:else if battingResource.error}
  <p class="mt-4 font-mono text-[0.78rem] text-warning">{battingResource.error}</p>
{:else if battingResource.value}
  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">{year} Team Batting</div>

    <div class="mb-4 grid grid-cols-4 gap-2 text-center">
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{battingResource.value.g ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">G</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{battingResource.value.r ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">R</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{battingResource.value.hr ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">HR</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{fmtRate(battingResource.value.ops)}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">OPS</div>
      </div>
    </div>

    {#if playerRows.length === 0}
      <p class="font-mono text-[0.78rem] text-muted">No player batting rows found for {year}.</p>
    {:else}
      <SortableTable {columns} rows={playerRows} />
    {/if}
  </div>
{:else}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">No batting data found for {year}.</p>
{/if}
