<script lang="ts">
  import { apiFetch, type DatasetStatus } from '$lib/api';
  import ApiMirrorStrip from '$lib/components/ApiMirrorStrip.svelte';
  import EraDisclaimer from '$lib/components/EraDisclaimer.svelte';
  import EraLegend from '$lib/components/EraLegend.svelte';
  import { EP } from '$lib/endpoints';
  import { STATIC_ERAS } from '$lib/eras';
  import SingleColLayout from '$lib/layouts/SingleColLayout.svelte';
  import { meta } from '$lib/meta.svelte.js';
  import { onMount } from 'svelte';

  type WinExpEraApi = {
    start_year: number;
    end_year: number;
    label: string;
    state_count: number;
    total_sample: number;
  };

  let weEras = $state<WinExpEraApi[]>([]);
  let weLoading = $state(true);
  let weError = $state<string | null>(null);

  onMount(async () => {
    meta.init();
    try {
      const result = await apiFetch<WinExpEraApi[]>(EP.winExpectancyEras);
      weEras = Array.isArray(result) ? result : [];
    } catch (e) {
      weError = e instanceof Error ? e.message : 'Failed to load';
    } finally {
      weLoading = false;
    }
  });

  function dsById(id: string): DatasetStatus | undefined {
    return meta.datasets.find((d) => d.id === id);
  }

  function healthDotCls(ds: DatasetStatus | undefined): string {
    if (!ds) return 'bg-outline';
    if (ds.healthy === false) return 'bg-warning';
    return 'bg-secondary';
  }

  function coverageLabel(ds: DatasetStatus | undefined): string {
    if (!ds) return '-';
    if (ds.coverage_from != null && ds.coverage_to != null) {
      return `${ds.coverage_from} – ${ds.coverage_to}`;
    }
    return `${ds.row_count.toLocaleString()} rows`;
  }

  function lastLoadedLabel(ds: DatasetStatus | undefined): string {
    if (!ds?.last_loaded_at) return '-';
    return new Date(ds.last_loaded_at).toLocaleDateString();
  }

  const lahmanDs = $derived(dsById('lahman'));
  const retroDs = $derived(dsById('retrosheet'));

  const CROSSWALK_ENDPOINTS = [
    { ep: EP.metaCrosswalkPlayers, by: 'player_id · retro_id · mlbam_id · bbref_id' },
    { ep: EP.metaCrosswalkTeams, by: 'team_id · franchise_id · mlbam_team_id' }
  ] as const;

  const fedEra = STATIC_ERAS.find((e) => e.code === 'fed');
  const nlgEra = STATIC_ERAS.find((e) => e.code === 'nlg');
  const historicalEras = [fedEra, nlgEra].filter((e): e is NonNullable<typeof e> => e != null);
</script>

<SingleColLayout>
  <div class="mb-8">
    <h1 class="mb-1 font-display text-2xl font-bold text-foreground">Data Sources</h1>
    <p class="text-sm text-muted">
      Three primary providers cover baseball from 1871 to the present. All data is ingested, normalized, and served
      through a unified API.
    </p>
  </div>

  <section class="mb-10">
    <div class="panel-label">Data Providers</div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
      <div class="rounded-lg bg-crust p-4">
        <div class="mb-2 flex items-start justify-between gap-2">
          <a
            href={lahmanDs?.source ?? 'https://sabr.org/lahman-database/'}
            target="_blank"
            rel="external"
            class="font-medium text-foreground no-underline hover:text-primary">
            Lahman Database
          </a>
          <span class="mt-1 h-2 w-2 shrink-0 rounded-full {healthDotCls(lahmanDs)}"></span>
        </div>
        <p class="mb-3 text-xs text-muted">
          Comprehensive historical statistics: batting, pitching, fielding, salaries, and awards from 1871 to present.
          Free for non-commercial use.
        </p>
        <div class="mb-1 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Coverage</div>
        <div class="mb-3 font-mono text-sm text-foreground">{coverageLabel(lahmanDs)}</div>
        <div class="space-y-px font-mono text-xxs text-muted">
          <div>/v1/players, /v1/stats/*</div>
          <div>/v1/seasons/*/leaders/*</div>
          <div>/v1/franchises, /v1/teams/*</div>
        </div>
      </div>

      <div class="rounded-lg bg-crust p-4">
        <div class="mb-2 flex items-start justify-between gap-2">
          <a
            href={retroDs?.source ?? 'https://www.retrosheet.org/'}
            target="_blank"
            rel="external"
            class="font-medium text-foreground no-underline hover:text-primary">
            Retrosheet
          </a>
          <span class="mt-1 h-2 w-2 shrink-0 rounded-full {healthDotCls(retroDs)}"></span>
        </div>
        <p class="mb-3 text-xs text-muted">
          Play-by-play game logs and event data enabling per-game, per-play, and per-pitch analysis. Cannot be used for
          commercial purposes.
        </p>
        <div class="mb-1 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Coverage</div>
        <div class="mb-3 font-mono text-sm text-foreground">{coverageLabel(retroDs)}</div>
        <div class="space-y-px font-mono text-xxs text-muted">
          <div>/v1/games, /v1/games/*/events</div>
          <div>/v1/games/*/plays, /*/pitches</div>
          <div>/v1/win-expectancy*</div>
        </div>
      </div>

      <div class="rounded-lg bg-crust p-4">
        <div class="mb-2 flex items-start justify-between gap-2">
          <a
            href="https://statsapi.mlb.com"
            target="_blank"
            rel="external"
            class="font-medium text-foreground no-underline hover:text-primary">
            MLB Stats API
          </a>
          <span class="mt-1 h-2 w-2 shrink-0 rounded-full bg-primary"></span>
        </div>
        <p class="mb-3 text-xs text-muted">
          Official MLB API for live scores, current standings, active player stats, and real-time game feeds. Subject to
          MLB terms of service.
        </p>
        <div class="mb-1 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Coverage</div>
        <div class="mb-3 font-mono text-sm text-foreground">Current season</div>
        <div class="space-y-px font-mono text-xxs text-muted">
          <div>/v1/mlb/schedule, /v1/mlb/standings</div>
          <div>/v1/mlb/stats, /v1/mlb/live/*</div>
          <div>/v1/mlb/people/*</div>
        </div>
      </div>
    </div>
  </section>

  <section class="mb-10">
    <div class="panel-label">Dataset Status</div>
    {#if meta.loading}
      <div class="space-y-2">
        {#each { length: 6 } as _, i (i)}
          <div class="h-10 animate-pulse rounded bg-crust"></div>
        {/each}
      </div>
    {:else if meta.error}
      <div class="rounded bg-crust px-4 py-3 font-mono text-xs text-warning">{meta.error}</div>
    {:else if meta.datasets.length > 0}
      <div class="overflow-hidden rounded-lg border border-outline">
        <table class="w-full">
          <thead>
            <tr class="border-b border-outline bg-crust">
              <th class="px-4 py-2.5 text-left font-mono text-[0.62rem] tracking-wider text-muted uppercase">
                Dataset
              </th>
              <th class="px-4 py-2.5 text-left font-mono text-[0.62rem] tracking-wider text-muted uppercase">
                Coverage
              </th>
              <th
                class="hidden px-4 py-2.5 text-right font-mono text-[0.62rem] tracking-wider text-muted uppercase sm:table-cell">
                Rows
              </th>
              <th
                class="hidden px-4 py-2.5 text-right font-mono text-[0.62rem] tracking-wider text-muted uppercase md:table-cell">
                Last loaded
              </th>
              <th class="px-4 py-2.5 text-center font-mono text-[0.62rem] tracking-wider text-muted uppercase">
                Health
              </th>
            </tr>
          </thead>
          <tbody>
            {#each meta.datasets as ds (ds.id)}
              {@const dotCls = healthDotCls(ds)}
              <tr class="border-b border-outline last:border-0 hover:bg-surface">
                <td class="px-4 py-2.5">
                  <a
                    href={ds.source}
                    target="_blank"
                    rel="external"
                    class="text-sm text-foreground no-underline hover:text-primary">
                    {ds.name}
                  </a>
                </td>
                <td class="px-4 py-2.5 font-mono text-[0.72rem] text-muted">{coverageLabel(ds)}</td>
                <td class="hidden px-4 py-2.5 text-right font-mono text-[0.72rem] text-muted sm:table-cell">
                  {ds.row_count.toLocaleString()}
                </td>
                <td class="hidden px-4 py-2.5 text-right font-mono text-[0.72rem] text-muted md:table-cell">
                  {lastLoadedLabel(ds)}
                </td>
                <td class="px-4 py-2.5 text-center">
                  <span class="inline-block h-2 w-2 rounded-full {dotCls}"></span>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <div class="rounded bg-crust px-4 py-3 font-mono text-xs text-muted">No dataset status available.</div>
    {/if}
  </section>

  <section class="mb-10">
    <div class="panel-label">Era Coverage</div>
    <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
      <div class="rounded-lg bg-crust p-4">
        <div class="mb-3 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Seed Eras</div>
        <EraLegend compact />
      </div>

      <div class="rounded-lg bg-crust p-4">
        <div class="mb-1 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Win Expectancy Eras</div>
        <div class="mb-3 font-mono text-xxs text-muted">
          Dynamic eras from <span class="text-primary">{EP.winExpectancyEras}</span>
        </div>
        {#if weLoading}
          <div class="space-y-2">
            {#each { length: 4 } as _, i (i)}
              <div class="h-7 animate-pulse rounded bg-surface"></div>
            {/each}
          </div>
        {:else if weError}
          <div class="rounded bg-surface px-3 py-2 font-mono text-xs text-warning">{weError}</div>
        {:else if weEras.length === 0}
          <div class="font-mono text-xs text-muted">No win expectancy eras loaded.</div>
        {:else}
          <div class="space-y-1">
            {#each weEras as era (era.start_year)}
              <div class="flex items-center gap-3 rounded px-2 py-1.5 hover:bg-surface">
                <div class="min-w-0 flex-1 text-xs text-foreground">
                  {era.label || `${era.start_year}–${era.end_year}`}
                </div>
                <div class="shrink-0 font-mono text-xxs text-muted">{era.start_year}–{era.end_year}</div>
                {#if era.total_sample > 0}
                  <div class="hidden shrink-0 font-mono text-[0.62rem] text-muted sm:block">
                    {era.total_sample.toLocaleString()} samples
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </section>

  <section class="mb-10">
    <div class="panel-label">ID Crosswalk</div>
    <p class="mb-4 text-sm text-muted">
      Players and teams carry different identifiers across Lahman, Retrosheet, and MLB. The crosswalk maps all three
      systems, plus Baseball Reference IDs, so queries can span source boundaries.
    </p>
    <div class="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
      <div class="rounded-lg bg-crust p-3">
        <div class="mb-1 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Lahman ID</div>
        <div class="mb-2 font-mono text-sm text-primary">ruthba01</div>
        <p class="text-xs text-muted">
          String key throughout the Lahman database. Derived from last name, first name, and a sequential suffix.
        </p>
      </div>
      <div class="rounded-lg bg-crust p-3">
        <div class="mb-1 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Retrosheet ID</div>
        <div class="mb-2 font-mono text-sm text-secondary">ruthb101</div>
        <p class="text-xs text-muted">
          String key in Retrosheet event files. Similar format to Lahman but not interchangeable, so we always use the
          crosswalk to translate.
        </p>
      </div>
      <div class="rounded-lg bg-crust p-3">
        <div class="mb-1 font-mono text-[0.62rem] tracking-wider text-muted uppercase">MLBAM ID</div>
        <div class="mb-2 font-mono text-sm text-foreground">121347</div>
        <p class="text-xs text-muted">
          Numeric ID from the MLB Stats API. Required for all live data endpoints. Sourced from the Chadwick Bureau
          register.
        </p>
      </div>
    </div>
    <div class="overflow-hidden rounded-lg border border-outline">
      <table class="w-full">
        <thead>
          <tr class="border-b border-outline bg-crust">
            <th class="px-4 py-2.5 text-left font-mono text-[0.62rem] tracking-wider text-muted uppercase">Endpoint</th>
            <th class="px-4 py-2.5 text-left font-mono text-[0.62rem] tracking-wider text-muted uppercase"
              >Lookup keys</th>
          </tr>
        </thead>
        <tbody>
          {#each CROSSWALK_ENDPOINTS as row (row.ep)}
            <tr class="border-b border-outline last:border-0 hover:bg-surface">
              <td class="px-4 py-2.5 font-mono text-[0.72rem] text-primary">{row.ep}</td>
              <td class="px-4 py-2.5 font-mono text-[0.72rem] text-muted">{row.by}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </section>

  <section class="mb-10">
    <div class="panel-label">Historical Leagues</div>
    {#if historicalEras.length > 0}
      <div class="mb-4">
        <EraDisclaimer eras={historicalEras} />
      </div>
    {/if}
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      {#if fedEra}
        <div class="rounded-lg bg-crust p-4">
          <div class="mb-1 font-sans text-sm font-medium text-foreground">{fedEra.label}</div>
          <div class="mb-2 font-mono text-xxs text-muted">{fedEra.from}–{fedEra.to}</div>
          <p class="mb-3 text-xs text-muted">{fedEra.caveat}</p>
          <div class="space-y-px font-mono text-xxs text-muted">
            <div>{EP.fedGames}</div>
            <div>{EP.fedTeams}</div>
            <div>{EP.fedPlays}</div>
          </div>
        </div>
      {/if}
      {#if nlgEra}
        <div class="rounded-lg bg-crust p-4">
          <div class="mb-1 font-sans text-sm font-medium text-foreground">{nlgEra.label}</div>
          <div class="mb-2 font-mono text-xxs text-muted">{nlgEra.from}–{nlgEra.to}</div>
          <p class="mb-3 text-xs text-muted">{nlgEra.caveat}</p>
          <div class="space-y-px font-mono text-xxs text-muted">
            <div>{EP.nlgGames}</div>
            <div>{EP.nlgTeams}</div>
            <div>{EP.nlgPlays}</div>
          </div>
        </div>
      {/if}
    </div>
  </section>

  <ApiMirrorStrip url="/v1/meta/datasets" />
</SingleColLayout>
