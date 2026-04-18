<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { AsyncValueResource } from '$lib/players/resources.svelte';
  import type { TableRow } from '$lib/players/types';
  import { rowColumns } from '$lib/players/types';
  import { normalizeTeamPitchingStats } from '$lib/teams/normalizers';
  import type { TeamPitchingStats } from '$lib/teams/types';
  import { onMount } from 'svelte';

  let teamId = $derived(page.params.id ?? '');
  let year = $derived(page.url.searchParams.get('year') ?? '');

  const pitchingResource = new AsyncValueResource<TeamPitchingStats>();
  let lastKey = '';

  let playerRows = $derived((pitchingResource.value?.players ?? []) as TableRow[]);
  let columns = $derived(rowColumns(playerRows));

  async function refresh(force = false): Promise<void> {
    const id = teamId;
    const y = year;
    if (!y) {
      pitchingResource.clear();
      return;
    }

    const key = `${id}|${y}`;
    if (!force && key === lastKey) return;
    lastKey = key;

    await pitchingResource.load(async () => {
      const payload = await apiFetch<Record<string, unknown>>(EP.seasonTeamPitching(y, id), { players: true });
      return normalizeTeamPitchingStats(payload);
    });
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });

  const fmtRate = (value: number | undefined) => (value != null ? Number(value).toFixed(2) : '—');
</script>

{#if !year}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Enter a season year in the sidebar to view pitching stats.</p>
{:else if pitchingResource.loading}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Loading…</p>
{:else if pitchingResource.error}
  <p class="mt-4 font-mono text-[0.78rem] text-warning">{pitchingResource.error}</p>
{:else if pitchingResource.value}
  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">{year} Team Pitching</div>

    <div class="mb-4 grid grid-cols-4 gap-2 text-center">
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{pitchingResource.value.w ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">W</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{pitchingResource.value.l ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">L</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{fmtRate(pitchingResource.value.era)}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">ERA</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{fmtRate(pitchingResource.value.whip)}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">WHIP</div>
      </div>
    </div>

    {#if playerRows.length === 0}
      <p class="font-mono text-[0.78rem] text-muted">No player pitching rows found for {year}.</p>
    {:else}
      <SortableTable {columns} rows={playerRows} />
    {/if}
  </div>
{:else}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">No pitching data found for {year}.</p>
{/if}
