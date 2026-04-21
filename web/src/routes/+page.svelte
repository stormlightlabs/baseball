<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import ApiMirrorStrip from '$lib/components/ApiMirrorStrip.svelte';
  import EraRangeChip from '$lib/components/EraRangeChip.svelte';
  import LeaderCards from '$lib/components/LeaderCards.svelte';
  import Pill from '$lib/components/Pill.svelte';
  import ScoreboardStrip from '$lib/components/ScoreboardStrip.svelte';
  import SearchInput from '$lib/components/SearchInput.svelte';
  import { STATIC_ERAS } from '$lib/eras';
  import {
    ALL_ENDPOINTS,
    DATASET_UI_HINTS,
    ENTITY_TYPES,
    FEATURED_GROUPS,
    FEATURED_QUERIES,
    QUICK_LINKS
  } from '$lib/home/constants';
  import { meta } from '$lib/meta.svelte.js';
  import { onMount } from 'svelte';

  onMount(() => meta.init());

  const API_DOCS_ROUTE = '/explorer' as const;

  let searchQuery = $state('');
  let activeEntity = $state<string | null>(null);

  const activeApiEndpoint = $derived.by(() => {
    const entity = ENTITY_TYPES.find((e) => e.label === activeEntity);
    return entity ? entity.apiEndpoint : '/v1/search/players';
  });

  function handleSearch() {
    const q = searchQuery.trim();
    if (!q) return;
    const entity = ENTITY_TYPES.find((e) => e.label === activeEntity);
    const path = entity?.path ?? '/players';
    goto(resolve(`${path}?q=${encodeURIComponent(q)}`));
  }

  function handlePill(entity: (typeof ENTITY_TYPES)[number]) {
    const q = searchQuery.trim();
    if (q) {
      goto(resolve(`${entity.path}?q=${encodeURIComponent(q)}`));
    } else {
      activeEntity = activeEntity === entity.label ? null : entity.label;
    }
  }

  let featuredGroup = $state<string>('standard');

  const visibleQueries = $derived(FEATURED_QUERIES.filter((q) => q.group === featuredGroup));

  type CoverageVariant = 'primary' | 'secondary' | 'warning';

  type CoverageEntry = {
    id: string;
    label: string;
    range: string;
    percent: number;
    variant: CoverageVariant;
    href?: string;
    tooltip?: string;
  };

  const coverageData = $derived.by((): CoverageEntry[] => {
    const coverageFrom = meta.dataFromYear ?? 1871;
    const coverageTo = meta.dataToYear ?? new Date().getFullYear();
    const coverageSpan = Math.max(1, coverageTo - coverageFrom);

    if (meta.datasets.length > 0) {
      return meta.datasets.map((ds) => {
        const fallbackCov = meta.coverage[ds.id];
        const covFrom = ds.coverage_from ?? fallbackCov?.from ?? null;
        const covTo = ds.coverage_to ?? fallbackCov?.to ?? null;

        let percent = 0;
        if (covFrom != null && covTo != null) {
          percent = Math.round(((covTo - covFrom) / coverageSpan) * 100);
        } else if (ds.row_count > 0) {
          percent = 100;
        }

        let variant: CoverageVariant = 'secondary';
        if (ds.healthy === false) variant = 'warning';
        else if (ds.id === 'lahman') variant = 'primary';

        return {
          id: ds.id,
          label: ds.name,
          range: covFrom != null && covTo != null ? `${covFrom} – ${covTo}` : `${ds.row_count.toLocaleString()} rows`,
          percent,
          variant,
          href: ds.source || undefined,
          tooltip: DATASET_UI_HINTS[ds.id]?.tooltip
        };
      });
    }
    return [
      { id: 'lahman', label: 'Lahman Database', range: '1871 – 2023', percent: 98, variant: 'primary' },
      { id: 'retrosheet', label: 'Retrosheet', range: '1871 – 2023', percent: 95, variant: 'secondary' },
      { id: 'mlb-statsapi', label: 'MLB StatsAPI', range: '2000 – present', percent: 75, variant: 'warning' }
    ];
  });
</script>

<main class="min-h-full bg-mantle pb-0">
  <section class="mx-auto max-w-3xl px-4 pt-10 pb-7 text-center sm:px-6 sm:pt-12 sm:pb-8 lg:px-8 lg:pt-14">
    <h1 class="mb-3 font-display text-3xl font-bold text-foreground">
      Big
      <span class="text-primary">Fly</span>
    </h1>
    <div class="mb-8 flex flex-col gap-1 text-base text-muted">
      <p>Baseball data from 1871 to now.</p>
      <p>Powered by: SABR · Retrosheet · MLB</p>
    </div>
    <SearchInput bind:value={searchQuery} placeholder="Search players, teams, games…" onsubmit={handleSearch} />

    <div class="mt-3 flex flex-wrap justify-center gap-2">
      {#each ENTITY_TYPES as entity (entity.label)}
        <Pill label={entity.label} active={activeEntity === entity.label} onclick={() => handlePill(entity)} />
      {/each}
    </div>

    {#if searchQuery.trim() || activeEntity}
      <div class="mt-2 font-mono text-xxs text-muted">
        → {activeApiEndpoint}?q=…
      </div>
    {/if}
  </section>

  <section class="mx-auto max-w-3xl px-4 pb-8 sm:px-6 lg:px-8">
    <div class="mb-2 text-center font-mono text-xxs tracking-wider text-muted uppercase">Jump to era</div>
    <div class="flex flex-wrap justify-center gap-2">
      {#each STATIC_ERAS as era (era.code)}
        <EraRangeChip {era} year={era.from} />
      {/each}
    </div>
  </section>

  <ScoreboardStrip />
  <LeaderCards />

  <div class="mx-auto max-w-6xl px-4 pb-6 sm:px-6 lg:px-8">
    <div class="grid grid-cols-1 gap-px overflow-hidden rounded-lg bg-outline md:grid-cols-2 xl:grid-cols-3">
      <div class="bg-crust p-5">
        <div class="panel-label">Quick links</div>
        <ul class="space-y-1">
          {#each QUICK_LINKS as link (link.label)}
            <li>
              <a
                href={resolve(link.path)}
                class="group flex items-center gap-3 rounded-md px-2 py-2.5 no-underline transition-colors hover:bg-surface">
                <div class="min-w-0 flex-1">
                  <div class="text-[0.82rem] font-medium text-foreground transition-colors group-hover:text-primary">
                    {link.label}
                  </div>
                  <div class="truncate font-mono text-xxs text-muted">{link.hint}</div>
                </div>
                <span class="mt-0.5 inline-flex items-center text-muted transition-colors group-hover:text-primary">
                  <i class="i-tabler-arrow-right"></i>
                </span>
              </a>
            </li>
          {/each}
        </ul>
      </div>

      <div class="bg-crust p-5">
        <div class="panel-label">Featured queries</div>
        <div class="mb-3 flex gap-1">
          {#each FEATURED_GROUPS as g (g.key)}
            <button
              class="rounded px-2 py-0.5 font-mono text-[0.63rem] transition-colors {featuredGroup === g.key
                ? 'bg-primary/20 text-primary'
                : 'text-muted hover:text-foreground'}"
              onclick={() => (featuredGroup = g.key)}>
              {g.label}
            </button>
          {/each}
        </div>
        <ul class="space-y-1">
          {#each visibleQueries as q (q.endpoint)}
            <li>
              <a
                href={resolve(API_DOCS_ROUTE)}
                target="_blank"
                rel="noreferrer"
                class="group block rounded-md px-2 py-2 no-underline transition-colors hover:bg-surface">
                <div class="mb-0.5 text-[0.82rem] text-foreground transition-colors group-hover:text-primary">
                  {q.title}
                </div>
                <div class="truncate font-mono text-xxs text-muted">{q.endpoint}</div>
              </a>
            </li>
          {/each}
        </ul>
      </div>

      <div class="bg-crust p-5">
        <div class="panel-label">API health</div>
        {#if meta.loading}
          <div class="grid grid-cols-2 gap-x-4 gap-y-4">
            {#each { length: 8 } as _, i (i)}
              <div>
                <div class="mb-1 h-2 w-12 rounded bg-outline"></div>
                <div class="h-4 w-20 rounded bg-surface"></div>
              </div>
            {/each}
          </div>
        {:else if meta.error}
          <div class="rounded-md bg-surface px-3 py-2 font-mono text-[0.72rem] text-warning">
            {meta.error}
          </div>
        {:else}
          <div class="grid grid-cols-2 gap-x-4 gap-y-4">
            <div>
              <div class="mb-0.5 font-mono text-[0.6rem] tracking-wider text-muted uppercase">status</div>
              {#if meta.status === 'online'}
                <div class="font-mono text-sm text-secondary">Online</div>
              {:else if meta.status === 'degraded'}
                <div class="font-mono text-sm text-rose-500">Degraded</div>
              {:else if meta.status === 'loading'}
                <div class="font-mono text-sm text-warning">Loading</div>
              {:else}
                <div class="font-mono text-sm text-primary">{meta.status}</div>
              {/if}
            </div>
            <div>
              <div class="mb-0.5 font-mono text-[0.6rem] tracking-wider text-muted uppercase">version</div>
              <div class="font-mono text-sm text-foreground">{meta.version}</div>
            </div>
            <div>
              <div class="mb-0.5 font-mono text-[0.6rem] tracking-wider text-muted uppercase">data from</div>
              <div class="font-mono text-sm text-foreground">{meta.dataFrom}</div>
            </div>
            <div>
              <div class="mb-0.5 font-mono text-[0.6rem] tracking-wider text-muted uppercase">data to</div>
              <div class="font-mono text-sm text-foreground">{meta.dataTo}</div>
            </div>
            <div>
              <div class="mb-0.5 font-mono text-[0.6rem] tracking-wider text-muted uppercase">datasets</div>
              <div class="font-mono text-sm text-foreground">
                {meta.healthyDatasetCount}/{meta.datasets.length || '-'} healthy
              </div>
            </div>
            <div>
              <div class="mb-0.5 font-mono text-[0.6rem] tracking-wider text-muted uppercase">required</div>
              <div class="font-mono text-sm text-foreground">
                {meta.requiredHealthyCount}/{meta.requiredDatasets.length || '-'} healthy
              </div>
            </div>
            <div>
              <div class="mb-0.5 font-mono text-[0.6rem] tracking-wider text-muted uppercase">generated</div>
              <div class="font-mono text-sm text-muted">{meta.generatedAt}</div>
            </div>
          </div>
        {/if}
      </div>

      <div class="col-span-3 flex gap-0.5">
        <div class="flex flex-1 flex-col bg-crust p-5">
          <div class="panel-label">Dataset coverage</div>
          <div class="grid flex-1 grid-cols-2 gap-1">
            {#each coverageData as item (item.id)}
              <div class="flex flex-1 flex-col items-center gap-3 rounded border border-outline p-2 hover:bg-surface">
                <div class="flex w-full items-center justify-baseline gap-2">
                  {#if item.href}
                    <a
                      href={item.href}
                      target="_blank"
                      rel="external"
                      class="inline-flex min-w-0 flex-1 items-center gap-1 truncate font-sans text-sm no-underline hover:underline"
                      title={item.tooltip ?? item.label}>
                      <span>
                        {item.label}
                      </span>
                      <i class="i-tabler-external-link"></i>
                    </a>
                  {:else}
                    <span
                      class="min-w-0 flex-1 truncate font-sans text-sm text-foreground"
                      title={item.tooltip ?? item.label}>
                      {item.label}
                    </span>
                  {/if}
                </div>
                <span class="w-full font-mono text-sm text-muted">{item.range}</span>
              </div>
            {/each}
          </div>
        </div>

        <div class="flex-1 bg-crust p-5">
          <div class="panel-label">Endpoints</div>
          <ul class="space-y-1.5">
            {#each ALL_ENDPOINTS as ep (ep)}
              <li>
                <a
                  href={resolve(API_DOCS_ROUTE)}
                  target="_blank"
                  rel="noreferrer"
                  class="block font-mono text-[0.72rem] text-primary no-underline opacity-80 transition-opacity hover:opacity-100">
                  {ep}
                </a>
              </li>
            {/each}
          </ul>
        </div>
      </div>
    </div>
  </div>

  <div class="mx-auto max-w-6xl px-4 pb-8 sm:px-6 lg:px-8">
    <ApiMirrorStrip url="/v1/meta" />
  </div>
</main>
