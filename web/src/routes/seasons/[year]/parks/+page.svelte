<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { normalizeParkFactors } from '$lib/seasons/normalizers';
  import type { SeasonParkFactor } from '$lib/seasons/types';
  import { onMount } from 'svelte';

  let routeYear = $derived(Number(page.params.year ?? ''));
  let selectedYear = $derived(Number.isFinite(routeYear) && routeYear > 0 ? routeYear : new Date().getFullYear());

  let parkFactors = $state<SeasonParkFactor[]>([]);
  let parkFactorsLoading = $state(false);
  let parkFactorsError = $state<string | null>(null);

  let parkFactorsRequestVersion = 0;
  let lastParkFactorsKey = '';

  let parkFactorRows = $derived.by(() => {
    return parkFactors
      .map((factor) => ({
        park_id: factor.park_id ?? '—',
        runs_factor: factor.runs_factor,
        hr_factor: factor.hr_factor,
        games_sampled: factor.games_sampled,
        provider: factor.provider ?? '—'
      }))
      .toSorted((a, b) => {
        const hrA = a.hr_factor ?? -1;
        const hrB = b.hr_factor ?? -1;
        if (hrA !== hrB) return hrB - hrA;
        return a.park_id.localeCompare(b.park_id);
      });
  });

  const parkFactorColumns = [
    { key: 'park_id', label: 'Park', sortable: true },
    {
      key: 'runs_factor',
      label: 'Runs',
      sortable: true,
      format: (value: unknown) => fmtFloat(value as number | undefined, 1)
    },
    {
      key: 'hr_factor',
      label: 'HR',
      sortable: true,
      format: (value: unknown) => fmtFloat(value as number | undefined, 1)
    },
    {
      key: 'games_sampled',
      label: 'Games',
      sortable: true,
      format: (value: unknown) => fmtInt(value as number | undefined)
    },
    { key: 'provider', label: 'Provider', sortable: true }
  ];

  onMount(() => {
    void refreshParkFactors(true);
  });

  afterNavigate(() => {
    void refreshParkFactors();
  });

  async function refreshParkFactors(force = false): Promise<void> {
    const key = String(selectedYear);
    if (!force && key === lastParkFactorsKey) return;
    lastParkFactorsKey = key;

    const requestVersion = ++parkFactorsRequestVersion;
    parkFactorsLoading = true;
    parkFactorsError = null;

    try {
      const payload = await apiFetch<unknown>(EP.seasonParkFactors(selectedYear));
      if (requestVersion !== parkFactorsRequestVersion) return;
      parkFactors = normalizeParkFactors(payload);
    } catch (error) {
      if (requestVersion !== parkFactorsRequestVersion) return;
      parkFactorsError = toErrorMessage(error, 'Failed to load park factors.');
      parkFactors = [];
    } finally {
      if (requestVersion === parkFactorsRequestVersion) parkFactorsLoading = false;
    }
  }

  function toErrorMessage(error: unknown, fallback: string): string {
    if (error instanceof Error && error.message.trim().length > 0) return error.message;
    return fallback;
  }

  function fmtFloat(value: number | undefined, digits = 2): string {
    if (value == null) return '—';
    return value.toFixed(digits);
  }

  function fmtInt(value: number | undefined): string {
    if (value == null) return '—';
    return Math.round(value).toLocaleString();
  }
</script>

<section class="rounded-lg border border-outline bg-crust p-4">
  <div class="mb-2 flex items-center justify-between gap-2">
    <div class="panel-label mb-0 border-0 p-0">Park factors snapshot</div>
    <a
      class="font-mono text-[0.68rem] text-primary underline-offset-2 hover:underline"
      href={resolve('/docs/api-computed')}>
      API docs
    </a>
  </div>

  {#if parkFactorsError}
    <p class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
      {parkFactorsError}
    </p>
  {/if}

  {#if parkFactorsLoading}
    <p class="font-mono text-[0.75rem] text-muted">Loading park factors...</p>
  {:else if parkFactorRows.length === 0}
    <p class="font-mono text-[0.75rem] text-muted">No park factor rows were returned for this season.</p>
  {:else}
    <SortableTable columns={parkFactorColumns} rows={parkFactorRows} />
  {/if}
</section>
