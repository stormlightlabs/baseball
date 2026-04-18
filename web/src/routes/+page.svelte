<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { ApiMirrorStrip, Chart, CoverageBar, Pill, SearchInput } from '$lib';
  import { meta } from '$lib/meta.svelte.js';
  import type { ChartConfiguration } from 'chart.js';
  import { onMount } from 'svelte';

  onMount(() => meta.init());

  let searchQuery = $state('');
  let activeEntity = $state<string | null>(null);

  const ENTITY_TYPES: { label: string; path: '/players' | '/teams' | '/games' | '/seasons' }[] = [
    { label: 'Players', path: '/players' },
    { label: 'Teams', path: '/teams' },
    { label: 'Franchises', path: '/teams' },
    { label: 'Games', path: '/games' },
    { label: 'Seasons', path: '/seasons' }
  ];

  function handleSearch() {
    const q = searchQuery.trim();
    if (!q) return;
    const path = ENTITY_TYPES.find((e) => e.label === activeEntity)?.path ?? '/players';
    goto(resolve(`${path}?q=${encodeURIComponent(q)}`));
  }

  function handlePill(label: string, path: '/players' | '/teams' | '/games' | '/seasons') {
    const q = searchQuery.trim();
    if (q) {
      goto(resolve(`${path}?q=${encodeURIComponent(q)}`));
    } else {
      activeEntity = activeEntity === label ? null : label;
    }
  }

  const QUICK_LINKS = [
    { label: 'Players', path: '/players', hint: '/v1/players', desc: 'Career stats, batting, pitching, awards' },
    { label: 'Teams', path: '/teams', hint: '/v1/franchises', desc: 'Franchise history and team-season records' },
    { label: 'Games', path: '/games', hint: '/v1/games', desc: 'Game finder with advanced filters' },
    {
      label: 'Leaders',
      path: '/leaders',
      hint: '/v1/seasons/{year}/leaders/{type}',
      desc: 'Stat leaderboards by season or career'
    },
    {
      label: 'Seasons',
      path: '/seasons',
      hint: '/v1/seasons/{year}/teams',
      desc: 'Season hub: teams, schedule, date explorer'
    }
  ] as const;

  const FEATURED_QUERIES: { title: string; endpoint: string }[] = [
    { title: 'HR leaders in 1927', endpoint: '/v1/seasons/1927/leaders/batting?stat=hr' },
    { title: 'Career HR leaders (min 3000 AB)', endpoint: '/v1/stats/batting?sort_by=hr&min_ab=3000' },
    { title: 'Extra-inning games in 2023', endpoint: '/v1/games?season=2023&min_innings=10' },
    { title: '1998 New York Yankees season', endpoint: '/v1/teams/NYA?year=1998' },
    { title: 'ERA leaders — Year of the Pitcher (1968)', endpoint: '/v1/seasons/1968/leaders/pitching?stat=era' },
    { title: 'Most saves in a season (all-time)', endpoint: '/v1/stats/pitching?sort_by=sv&sort_order=desc&min_ip=1' }
  ];

  const ALL_ENDPOINTS = [
    '/v1/players',
    '/v1/players/{id}/seasons',
    '/v1/players/{id}/awards',
    '/v1/teams',
    '/v1/franchises',
    '/v1/games',
    '/v1/seasons/{year}/teams',
    '/v1/seasons/{year}/leaders/{type}',
    '/v1/stats/batting',
    '/v1/stats/pitching',
    '/v1/meta'
  ] as const;

  const SOURCE_COLORS: Record<string, string> = { lahman: '#3b82f6', retrosheet: '#10b981' };

  let coverageChartConfig = $derived.by((): ChartConfiguration => {
    const coverage = meta.coverage;
    const decades = Array.from({ length: 16 }, (_, i) => 1870 + i * 10);

    const datasets = Object.entries(coverage).map(([key, cov]) => ({
      label: key.charAt(0).toUpperCase() + key.slice(1),
      backgroundColor: SOURCE_COLORS[key] ?? '#6b7280',
      borderRadius: 2,
      data: decades.map((d) => {
        const from = cov.from ?? 9999;
        const to = cov.to ?? 0;
        return d >= from - 5 && d <= to ? 1 : 0;
      })
    }));

    if (!datasets.length) {
      datasets.push(
        { label: 'Lahman', backgroundColor: '#3b82f6', borderRadius: 2, data: decades.map((d) => (d >= 1871 ? 1 : 0)) },
        {
          label: 'Retrosheet',
          backgroundColor: '#10b981',
          borderRadius: 2,
          data: decades.map((d) => (d >= 1871 ? 1 : 0))
        }
      );
    }

    return {
      type: 'bar',
      data: { labels: decades.map((d) => `'${String(d).slice(2)}`), datasets },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false }, tooltip: { enabled: false } },
        scales: {
          x: {
            stacked: true,
            grid: { display: false },
            border: { display: false },
            ticks: { color: '#6b7280', font: { size: 9 }, maxRotation: 0 }
          },
          y: { stacked: true, display: false, max: Math.max(2, Object.keys(coverage).length) + 0.5 }
        },
        animation: { duration: 600 }
      }
    };
  });

  let coverageBars = $derived.by(() => {
    if (meta.datasets.length > 0) {
      return meta.datasets.map((ds) => ({
        label: ds.name,
        range: `${ds.coverage_from ?? '?'} – ${ds.coverage_to ?? '?'}`,
        percent:
          ds.coverage_from && ds.coverage_to
            ? Math.round(((ds.coverage_to - ds.coverage_from) / (2024 - 1871)) * 100)
            : 80,
        variant: (ds.source === 'lahman' ? 'primary' : ds.source === 'retrosheet' ? 'secondary' : 'warning') as
          | 'primary'
          | 'secondary'
          | 'warning'
      }));
    }
    return [
      { label: 'Lahman Database', range: '1871 – 2023', percent: 98, variant: 'primary' as const },
      { label: 'Retrosheet', range: '1871 – 2023', percent: 95, variant: 'secondary' as const },
      { label: 'MLB StatsAPI', range: '2000 – present', percent: 75, variant: 'warning' as const }
    ];
  });
</script>

<main class="min-h-[calc(100vh-3.5rem)] bg-mantle pb-0">
  <!-- Search hero -->
  <section class="mx-auto max-w-3xl px-8 pt-14 pb-10 text-center">
    <h1 class="mb-3 font-display text-3xl font-bold text-foreground">Baseball API</h1>
    <p class="mb-8 text-[0.9rem] text-muted">
      Historical baseball data from 1871 to present — players, teams, games, and stats.
      <br />Lahman · Retrosheet · MLB StatsAPI
    </p>
    <SearchInput bind:value={searchQuery} placeholder="Search players, teams, games…" onsubmit={handleSearch} />
    <div class="mt-4 flex flex-wrap justify-center gap-2">
      {#each ENTITY_TYPES as { label, path } (label)}
        <Pill {label} active={activeEntity === label} onclick={() => handlePill(label, path)} />
      {/each}
    </div>
  </section>

  <!-- Dashboard grid -->
  <div class="mx-auto max-w-6xl px-8 pb-6">
    <div class="grid grid-cols-3 gap-px overflow-hidden rounded-lg bg-outline">
      <!-- Row 1 -->

      <!-- Quick Links -->
      <div class="bg-crust p-5">
        <div class="panel-label">Quick links</div>
        <ul class="space-y-1">
          {#each QUICK_LINKS as link (link.label)}
            <li>
              <a
                href={resolve(link.path)}
                class="group flex items-start gap-3 rounded-md px-2 py-2.5 no-underline transition-colors hover:bg-surface">
                <div class="min-w-0 flex-1">
                  <div class="text-[0.82rem] font-medium text-foreground transition-colors group-hover:text-primary">
                    {link.label}
                  </div>
                  <div class="truncate font-monospace text-[0.65rem] text-muted">{link.hint}</div>
                </div>
                <span class="mt-0.5 text-muted transition-colors group-hover:text-primary">→</span>
              </a>
            </li>
          {/each}
        </ul>
      </div>

      <!-- Featured Queries -->
      <div class="bg-crust p-5">
        <div class="panel-label">Featured queries</div>
        <ul class="space-y-1">
          {#each FEATURED_QUERIES as q (q.title)}
            <li>
              <a
                href={resolve(`/explorer?endpoint=${encodeURIComponent(q.endpoint)}`)}
                class="group block rounded-md px-2 py-2 no-underline transition-colors hover:bg-surface">
                <div class="mb-0.5 text-[0.82rem] text-foreground transition-colors group-hover:text-primary">
                  {q.title}
                </div>
                <div class="truncate font-monospace text-[0.65rem] text-muted">{q.endpoint}</div>
              </a>
            </li>
          {/each}
        </ul>
      </div>

      <!-- API Health -->
      <div class="bg-crust p-5">
        <div class="panel-label">API health</div>
        {#if meta.loading}
          <div class="grid grid-cols-2 gap-x-4 gap-y-4">
            {#each { length: 6 } as i (i)}
              <div>
                <div class="mb-1 h-2 w-12 rounded bg-outline"></div>
                <div class="h-4 w-20 rounded bg-surface"></div>
              </div>
            {/each}
          </div>
        {:else if meta.error}
          <div class="rounded-md bg-surface px-3 py-2 font-monospace text-[0.72rem] text-warning">
            {meta.error}
          </div>
        {:else}
          <div class="grid grid-cols-2 gap-x-4 gap-y-4">
            <div>
              <div class="mb-0.5 font-monospace text-[0.6rem] tracking-wider text-muted uppercase">status</div>
              <div
                class="font-monospace text-sm {meta.status === 'online'
                  ? 'text-secondary'
                  : meta.status === 'loading'
                    ? 'text-warning'
                    : 'text-red-400'}">
                {meta.status}
              </div>
            </div>
            <div>
              <div class="mb-0.5 font-monospace text-[0.6rem] tracking-wider text-muted uppercase">version</div>
              <div class="font-monospace text-sm text-foreground">{meta.version}</div>
            </div>
            <div>
              <div class="mb-0.5 font-monospace text-[0.6rem] tracking-wider text-muted uppercase">data from</div>
              <div class="font-monospace text-sm text-foreground">{meta.dataFrom}</div>
            </div>
            <div>
              <div class="mb-0.5 font-monospace text-[0.6rem] tracking-wider text-muted uppercase">data to</div>
              <div class="font-monospace text-sm text-foreground">{meta.dataTo}</div>
            </div>
            <div>
              <div class="mb-0.5 font-monospace text-[0.6rem] tracking-wider text-muted uppercase">sources</div>
              <div class="font-monospace text-sm text-foreground">
                {meta.datasets.length || '—'}
              </div>
            </div>
            <div>
              <div class="mb-0.5 font-monospace text-[0.6rem] tracking-wider text-muted uppercase">generated</div>
              <div class="font-monospace text-sm text-muted">{meta.generatedAt}</div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Row 2 -->

      <!-- Dataset Coverage (col-span-2) -->
      <div class="col-span-2 bg-crust p-5">
        <div class="panel-label">Dataset coverage</div>
        <div class="mb-4 space-y-2">
          {#each coverageBars as bar (bar.label)}
            <CoverageBar label={bar.label} range={bar.range} percent={bar.percent} variant={bar.variant} />
          {/each}
        </div>
        <div class="h-20">
          <Chart config={coverageChartConfig} height={80} />
        </div>
      </div>

      <!-- Endpoints -->
      <div class="bg-crust p-5">
        <div class="panel-label">Endpoints</div>
        <ul class="space-y-1.5">
          {#each ALL_ENDPOINTS as ep (ep)}
            <li>
              <a
                href={resolve(`/explorer?endpoint=${encodeURIComponent(ep)}`)}
                class="block font-monospace text-[0.72rem] text-primary no-underline opacity-80 transition-opacity hover:opacity-100">
                {ep}
              </a>
            </li>
          {/each}
        </ul>
      </div>
    </div>
  </div>

  <!-- API Mirror Strip -->
  <div class="mx-auto max-w-6xl px-8 pb-8">
    <ApiMirrorStrip url="/v1/meta" />
  </div>
</main>
