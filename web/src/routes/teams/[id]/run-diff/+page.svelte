<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { AsyncValueResource } from '$lib/players/resources.svelte';
  import type { TableRow } from '$lib/players/types';
  import { rowColumns } from '$lib/players/types';
  import { normalizeRunDifferentialSeries } from '$lib/teams/normalizers';
  import type { RunDifferentialSeries } from '$lib/teams/types';
  import { onMount } from 'svelte';

  let teamId = $derived(page.params.id ?? '');
  let year = $derived(page.url.searchParams.get('year') ?? '');

  const runDiffResource = new AsyncValueResource<RunDifferentialSeries>();
  let lastKey = '';

  let gameRows = $derived((runDiffResource.value?.games ?? []) as TableRow[]);
  let columns = $derived(rowColumns(gameRows));

  async function refresh(force = false): Promise<void> {
    const id = teamId;
    const y = year;
    if (!y) {
      runDiffResource.clear();
      return;
    }

    const key = `${id}|${y}`;
    if (!force && key === lastKey) return;
    lastKey = key;

    await runDiffResource.load(async () => {
      const payload = await apiFetch<Record<string, unknown>>(EP.teamRunDifferential(id), { season: y });
      return normalizeRunDifferentialSeries(payload);
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
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Enter a season year in the sidebar to view run differential.</p>
{:else if runDiffResource.loading}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Loading…</p>
{:else if runDiffResource.error}
  <p class="mt-4 font-mono text-[0.78rem] text-warning">{runDiffResource.error}</p>
{:else if runDiffResource.value}
  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">Run Differential — {year}</div>

    <div class="mb-4 grid grid-cols-4 gap-2 text-center">
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{runDiffResource.value.games_played ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">Games</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{runDiffResource.value.runs_scored ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">RS</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{runDiffResource.value.runs_allowed ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">RA</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{runDiffResource.value.run_differential ?? '—'}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">Diff</div>
      </div>
    </div>

    {#if gameRows.length > 0}
      <SortableTable {columns} rows={gameRows} />
    {:else}
      <p class="font-mono text-[0.78rem] text-muted">No per-game run differential data found.</p>
    {/if}
  </div>
{:else}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">No run differential data found.</p>
{/if}
