<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, apiUrl, fetchPaginated } from '$lib/api';
  import AdvancedTabPanel from '$lib/components/AdvancedTabPanel.svelte';
  import ApiPanel from '$lib/components/ApiPanel.svelte';
  import Chart from '$lib/components/Chart.svelte';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import EraRangeChip from '$lib/components/EraRangeChip.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import SearchInput from '$lib/components/SearchInput.svelte';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import TabRow from '$lib/components/TabRow.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear, erasInRange, STATIC_ERAS } from '$lib/eras';
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
  import {
    type ApiListPayload,
    type Award,
    type BattingSeason,
    type GameLog,
    type HofEntry,
    type PitchingSeason,
    type PlayerProfile,
    type PlayerResult,
    type PlayerTeam,
    type Relative,
    type Salary,
    normalizeApiList,
    rowColumns
  } from '$lib/players/types';
  import { PlayersUiController } from '$lib/players.svelte';
  import { AsyncListResource, AsyncPaginatedListResource, AsyncValueResource } from '$lib/players/resources.svelte';
  import { intParam, setUrlParams } from '$lib/url-state.svelte';
  import { type ChartConfiguration } from 'chart.js';
  import { onMount } from 'svelte';

  let q = $derived(page.url.searchParams.get('q') ?? '');
  let playerId = $derived(page.url.searchParams.get('id') ?? '');
  let activeTab = $derived(page.url.searchParams.get('tab') ?? 'batting');
  let gameLogType = $derived(page.url.searchParams.get('log') ?? 'batting');
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 20));
  let batStat = $derived(page.url.searchParams.get('stat') ?? 'hr');
  let showAdvanced = $derived(page.url.searchParams.get('advanced') === '1');

  const ui = new PlayersUiController();
  let mounted = false;

  onMount(() => {
    ui.syncFromUrl({ q, tab: activeTab, page: currentPage, perPage });
    mounted = true;
  });

  $effect(() => {
    const nextQ = q;
    const nextTab = activeTab;
    const nextPage = currentPage;
    const nextPerPage = perPage;
    if (!mounted) return;
    ui.syncFromUrl({ q: nextQ, tab: nextTab, page: nextPage, perPage: nextPerPage });
  });

  $effect(() => {
    const params = ui.nextUrlParams;
    if (!params) return;
    void setUrlParams(Object.fromEntries(params));
  });

  const searchResource = new AsyncPaginatedListResource<PlayerResult>();
  const profileResource = new AsyncValueResource<PlayerProfile>();
  const battingResource = new AsyncPaginatedListResource<BattingSeason>();
  const pitchingResource = new AsyncPaginatedListResource<PitchingSeason>();
  const awardsResource = new AsyncListResource<Award>();
  const hofResource = new AsyncListResource<HofEntry>();
  const teamsResource = new AsyncListResource<PlayerTeam>();
  const salariesResource = new AsyncListResource<Salary>();
  const relativesResource = new AsyncListResource<Relative>();
  const gameLogsResource = new AsyncPaginatedListResource<GameLog>();

  let searchResults = $derived(searchResource.items);
  let searchLoading = $derived(searchResource.loading);

  let profile = $derived(profileResource.value);
  let profileLoading = $derived(profileResource.loading);
  let profileError = $derived(profileResource.error);

  let battingSeasons = $derived(battingResource.items);
  let battingTotal = $derived(battingResource.total);
  let battingLoading = $derived(battingResource.loading);

  let pitchingSeasons = $derived(pitchingResource.items);
  let pitchingTotal = $derived(pitchingResource.total);
  let pitchingLoading = $derived(pitchingResource.loading);

  let awards = $derived(awardsResource.items);
  let awardsLoading = $derived(awardsResource.loading);

  let hofEntries = $derived(hofResource.items);
  let hofLoading = $derived(hofResource.loading);

  let teams = $derived(teamsResource.items);
  let teamsLoading = $derived(teamsResource.loading);

  let salaries = $derived(salariesResource.items);
  let salariesLoading = $derived(salariesResource.loading);

  let relatives = $derived(relativesResource.items);
  let relativesLoading = $derived(relativesResource.loading);

  let gameLogs = $derived(gameLogsResource.items);
  let gameLogsTotal = $derived(gameLogsResource.total);
  let gameLogsLoading = $derived(gameLogsResource.loading);

  async function fetchList<T>(endpoint: string): Promise<T[]> {
    const payload = await apiFetch<ApiListPayload<T>>(endpoint);
    return normalizeApiList(payload);
  }

  // ── Fetch: search ──────────────────────────────────────────────────────────

  $effect(() => {
    const thisQ = q;
    if (!thisQ) {
      searchResource.clear();
      return;
    }
    void searchResource.load(() => fetchPaginated<PlayerResult>(EP.searchPlayers, { q: thisQ, per_page: 20 }));
  });

  // ── Fetch: player profile ──────────────────────────────────────────────────

  $effect(() => {
    const id = playerId;
    if (!id) {
      profileResource.clear();
      return;
    }
    void profileResource.load(() => apiFetch<PlayerProfile>(EP.player(id)));
  });

  // ── Fetch: batting ─────────────────────────────────────────────────────────

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    const p = currentPage;
    const pp = perPage;
    if (!id || tab !== 'batting') return;
    void battingResource.load(() =>
      fetchPaginated<BattingSeason>(EP.playerStatsBatting(id), { page: p, per_page: pp })
    );
  });

  // ── Fetch: pitching ────────────────────────────────────────────────────────

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    const p = currentPage;
    const pp = perPage;
    if (!id || tab !== 'pitching') return;
    void pitchingResource.load(() =>
      fetchPaginated<PitchingSeason>(EP.playerStatsPitching(id), { page: p, per_page: pp })
    );
  });

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    if (!id || tab !== 'awards') return;
    void awardsResource.load(() => fetchList<Award>(EP.playerAwards(id)));
  });

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    if (!id || tab !== 'hof') return;
    void hofResource.load(() => fetchList<HofEntry>(EP.playerHallOfFame(id)));
  });

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    if (!id || tab !== 'teams') return;
    void teamsResource.load(() => fetchList<PlayerTeam>(EP.playerTeams(id)));
  });

  // ── Fetch: salaries ───────────────────────────────────────────────────────

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    if (!id || tab !== 'salaries') return;
    void salariesResource.load(() => fetchList<Salary>(EP.playerSalaries(id)));
  });

  // ── Fetch: relatives ──────────────────────────────────────────────────────

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    if (!id || tab !== 'relatives') return;
    void relativesResource.load(() => fetchList<Relative>(EP.playerRelatives(id)));
  });

  // ── Fetch: game logs ──────────────────────────────────────────────────────

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    const logType = gameLogType;
    const p = currentPage;
    const pp = perPage;
    if (!id || tab !== 'game-logs') return;
    const logEP =
      logType === 'pitching'
        ? EP.playerGameLogsPitching(id)
        : logType === 'fielding'
          ? EP.playerGameLogsFielding(id)
          : EP.playerGameLogsBatting(id);
    void gameLogsResource.load(() => fetchPaginated<GameLog>(logEP, { page: p, per_page: pp }));
  });

  // ── Derived: career era span ───────────────────────────────────────────────

  let careerEras = $derived.by(() => {
    if (!profile?.debut_year || !profile?.final_year) return [];
    return erasInRange(profile.debut_year, profile.final_year);
  });

  // ── Batting / pitching rows with team display ─────────────────────────────

  let battingRows = $derived(battingSeasons.map((s) => ({ ...s, team: s.team ?? s.team_id ?? '?' })));

  let pitchingRows = $derived(pitchingSeasons.map((s) => ({ ...s, team: s.team ?? s.team_id ?? '?' })));
  let gameLogColumns = $derived(rowColumns(gameLogs));

  // ── Chart configs ─────────────────────────────────────────────────────────

  const BATTING_STATS = [
    { value: 'hr', label: 'Home Runs (HR)' },
    { value: 'avg', label: 'Batting Avg (AVG)' },
    { value: 'rbi', label: 'RBI' },
    { value: 'sb', label: 'Stolen Bases (SB)' },
    { value: 'obp', label: 'OBP' },
    { value: 'slg', label: 'SLG' },
    { value: 'ops', label: 'OPS' }
  ] as const;

  const CHART_SCALES = {
    x: { ticks: { color: '#6b7280', font: { size: 9 } }, grid: { color: '#252934' } },
    y: { ticks: { color: '#6b7280', font: { size: 9 } }, grid: { color: '#252934' } }
  } as const;

  let battingChartConfig = $derived.by((): ChartConfiguration => {
    const isRate = ['avg', 'obp', 'slg', 'ops'].includes(batStat);
    const data = battingRows.map((s) =>
      isRate
        ? Math.round(Number(s[batStat as keyof BattingSeason] ?? 0) * 1000)
        : Number(s[batStat as keyof BattingSeason] ?? 0)
    );
    return {
      type: 'line',
      data: {
        labels: battingRows.map((s) => String(s.year)),
        datasets: [
          {
            label: BATTING_STATS.find((x) => x.value === batStat)?.label ?? batStat.toUpperCase(),
            data,
            borderColor: '#3b82f6',
            backgroundColor: 'rgba(59,130,246,0.1)',
            pointRadius: 3,
            tension: 0.3,
            fill: true
          }
        ]
      },
      options: {
        responsive: true,
        plugins: { legend: { labels: { color: '#6b7280', font: { family: 'Inter', size: 10 } } } },
        scales: CHART_SCALES
      }
    };
  });

  let pitchingChartConfig = $derived.by(
    (): ChartConfiguration => ({
      type: 'line',
      data: {
        labels: pitchingRows.map((s) => String(s.year)),
        datasets: [
          {
            label: 'ERA',
            data: pitchingRows.map((s) => Number(s.era ?? 0)),
            borderColor: '#10b981',
            backgroundColor: 'rgba(16,185,129,0.1)',
            pointRadius: 3,
            tension: 0.3,
            fill: true
          }
        ]
      },
      options: {
        responsive: true,
        plugins: { legend: { labels: { color: '#6b7280', font: { family: 'Inter', size: 10 } } } },
        scales: CHART_SCALES
      }
    })
  );

  let hofChartConfig = $derived.by(
    (): ChartConfiguration => ({
      type: 'bar',
      data: {
        labels: hofEntries.map((e) => String(e.year_inducted ?? '?')),
        datasets: [
          { label: 'Vote %', data: hofEntries.map((e) => Number(e.pct ?? 0)), backgroundColor: 'rgba(16,185,129,0.6)' }
        ]
      },
      options: {
        responsive: true,
        plugins: { legend: { labels: { color: '#6b7280', font: { size: 10 } } } },
        scales: {
          x: { ticks: { color: '#6b7280' }, grid: { color: '#252934' } },
          y: { min: 0, max: 100, ticks: { color: '#6b7280', callback: (v) => v + '%' }, grid: { color: '#252934' } }
        }
      }
    })
  );

  // ── Active endpoint (for API panel) ───────────────────────────────────────

  let activeEndpoint = $derived.by(() => {
    if (!playerId) return EP.searchPlayers;
    switch (activeTab) {
      case 'batting':
        return EP.playerStatsBatting(playerId);
      case 'pitching':
        return EP.playerStatsPitching(playerId);
      case 'awards':
        return EP.playerAwards(playerId);
      case 'hof':
        return EP.playerHallOfFame(playerId);
      case 'teams':
        return EP.playerTeams(playerId);
      case 'salaries':
        return EP.playerSalaries(playerId);
      case 'relatives':
        return EP.playerRelatives(playerId);
      case 'war':
        return EP.playerStatsWar(playerId);
      case 'batting-adv':
        return EP.playerStatsBattingAdv(playerId);
      case 'pitching-adv':
        return EP.playerStatsPitchingAdv(playerId);
      case 'splits':
        return EP.playerSplits(playerId);
      case 'streaks':
        return EP.playerStreaks(playerId);
      case 'game-logs':
        return gameLogType === 'pitching'
          ? EP.playerGameLogsPitching(playerId)
          : gameLogType === 'fielding'
            ? EP.playerGameLogsFielding(playerId)
            : EP.playerGameLogsBatting(playerId);
      default:
        return EP.player(playerId);
    }
  });

  let activeUrl = $derived(apiUrl(activeEndpoint));

  const MAIN_TABS = [
    { id: 'batting', label: 'Batting' },
    { id: 'pitching', label: 'Pitching' },
    { id: 'game-logs', label: 'Game Logs' },
    { id: 'awards', label: 'Awards' },
    { id: 'hof', label: 'Hall of Fame' },
    { id: 'teams', label: 'Teams' },
    { id: 'salaries', label: 'Salaries' },
    { id: 'relatives', label: 'Relatives' }
  ];

  const ADV_TABS = [
    { id: 'batting-adv', label: 'Batting Adv.' },
    { id: 'pitching-adv', label: 'Pitching Adv.' },
    { id: 'war', label: 'WAR' },
    { id: 'splits', label: 'Splits' },
    { id: 'streaks', label: 'Streaks' }
  ];

  let allTabs = $derived(showAdvanced ? [...MAIN_TABS, ...ADV_TABS] : MAIN_TABS);

  const teamsColumns = [
    { key: 'year', label: 'Year', sortable: true },
    { key: 'team', label: 'Team', sortable: true },
    { key: 'league', label: 'League', sortable: true },
    { key: 'g', label: 'G', sortable: true }
  ];

  const salariesColumns = [
    { key: 'year', label: 'Year', sortable: true },
    { key: 'team', label: 'Team', sortable: true },
    {
      key: 'salary',
      label: 'Salary',
      sortable: true,
      rank: true,
      format: (v: unknown) => `$${Number(v).toLocaleString()}`
    }
  ];

  function handleSearch(value = ui.search) {
    const trimmed = value.trim();
    if (!trimmed) return;
    void setUrlParams({ q: trimmed, id: null, tab: 'batting', page: 1 });
  }

  function selectPlayer(id: string) {
    void setUrlParams({ id, tab: 'batting', page: 1, stat: null });
  }

  function fmtAvg(v: number | undefined) {
    return v != null ? Number(v).toFixed(3) : '—';
  }

  function fmtNum(v: number | undefined) {
    return v != null ? String(v) : '—';
  }
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <!-- Search -->
    <div class="panel-label">Player search</div>
    <SearchInput mini bind:value={ui.search} placeholder="Name, ID…" onsubmit={handleSearch} />

    <!-- Bio card when a player is selected -->
    {#if playerId}
      <div class="mt-4 rounded-lg border border-outline bg-surface p-4">
        {#if profileLoading}
          <p class="font-monospace text-[0.72rem] text-muted">Loading…</p>
        {:else if profileError}
          <p class="font-monospace text-[0.72rem] text-warning">{profileError}</p>
        {:else if profile}
          <div class="mb-0.5 text-center font-monospace text-[0.6rem] tracking-wider text-muted uppercase">
            {profile.primary_position ?? profile.positions?.[0] ?? ''}
          </div>
          <div class="mb-1 text-center font-display text-[0.95rem] font-medium text-foreground">
            {profile.name}
          </div>
          {#if profile.birth_date || profile.birth_city}
            <div class="mb-1 text-center font-monospace text-[0.68rem] text-muted">
              {profile.birth_date ?? ''}
              {#if profile.birth_city}
                · {profile.birth_city}{#if profile.birth_state}, {profile.birth_state}{/if}{/if}
            </div>
          {/if}
          {#if profile.bats || profile.throws || profile.debut_year}
            <div class="mb-3 text-center font-monospace text-[0.68rem] text-muted">
              {#if profile.bats}Bats: {profile.bats}{/if}
              {#if profile.bats && profile.throws}
                ·
              {/if}
              {#if profile.throws}Throws: {profile.throws}{/if}
              {#if profile.debut_year}
                · {profile.debut_year}–{profile.final_year ?? 'pres.'}{/if}
            </div>
          {/if}
          <!-- Career quick stats -->
          {#if profile.career_hr != null || profile.career_avg != null || profile.career_rbi != null}
            <div class="mb-3 grid grid-cols-3 gap-1 text-center">
              {#if profile.career_hr != null}
                <div>
                  <div class="font-display text-[0.95rem] font-semibold text-foreground">{profile.career_hr}</div>
                  <div class="font-monospace text-[0.58rem] text-muted uppercase">HR</div>
                </div>
              {/if}
              {#if profile.career_avg != null}
                <div>
                  <div class="font-display text-[0.95rem] font-semibold text-foreground">
                    {fmtAvg(profile.career_avg)}
                  </div>
                  <div class="font-monospace text-[0.58rem] text-muted uppercase">AVG</div>
                </div>
              {/if}
              {#if profile.career_rbi != null}
                <div>
                  <div class="font-display text-[0.95rem] font-semibold text-foreground">{profile.career_rbi}</div>
                  <div class="font-monospace text-[0.58rem] text-muted uppercase">RBI</div>
                </div>
              {/if}
            </div>
          {/if}
          <!-- Career era span -->
          {#if careerEras.length > 0}
            <div class="border-t border-outline pt-3">
              <div class="mb-1.5 font-monospace text-[0.6rem] tracking-wider text-muted uppercase">Career eras</div>
              <div class="flex flex-wrap gap-1">
                {#each careerEras as era (era.code)}
                  <EraRangeChip {era} />
                {/each}
              </div>
            </div>
          {/if}
        {/if}
      </div>
    {/if}

    <!-- Search results -->
    {#if searchResults.length > 0}
      <div class="mt-4 panel-label">Results</div>
      <div class="flex flex-col gap-0.5">
        {#each searchResults as result (result.id)}
          <button
            onclick={() => selectPlayer(result.id)}
            class="rounded-md px-3 py-2 text-left transition-colors hover:bg-surface {playerId === result.id
              ? 'bg-surface'
              : ''}">
            <div class="font-display text-[0.8rem] text-foreground">{result.name}</div>
            <div class="font-monospace text-[0.68rem] text-muted">
              {result.id}
              {#if result.primary_position ?? result.position}
                · {result.primary_position ?? result.position}{/if}
              {#if result.debut_year}
                · {result.debut_year}–{result.final_year ?? 'pres.'}{/if}
            </div>
          </button>
        {/each}
      </div>
    {:else if searchLoading}
      <p class="mt-3 font-monospace text-[0.72rem] text-muted">Searching…</p>
    {:else if !q && !playerId}
      <p class="mt-4 font-monospace text-[0.72rem] text-muted">Search for a player to begin.</p>
    {/if}
  {/snippet}

  {#snippet center()}
    {#if !playerId}
      <!-- Empty state -->
      <div class="flex h-full flex-col items-center justify-center gap-3 text-center">
        <div class="font-display text-[1.1rem] text-muted">Player Explorer</div>
        <p class="max-w-xs font-monospace text-[0.78rem] text-muted/60">
          Search for a player in the sidebar to explore career stats, awards, game logs, and more.
        </p>
        <div class="mt-2 flex flex-wrap justify-center gap-1.5">
          {#each STATIC_ERAS as era (era.code)}
            <EraRangeChip {era} />
          {/each}
        </div>
      </div>
    {:else}
      <!-- Tab row + advanced toggle -->
      <div class="mb-1 flex flex-wrap items-center gap-2">
        <TabRow tabs={allTabs} bind:active={ui.tab} />
        <button
          onclick={() =>
            setUrlParams({ advanced: showAdvanced ? null : '1', tab: showAdvanced ? 'batting' : activeTab })}
          class="ml-auto shrink-0 rounded border px-2.5 py-1 font-monospace text-[0.68rem] transition-colors {showAdvanced
            ? 'border-primary/40 text-primary hover:border-primary'
            : 'border-outline text-muted hover:border-primary hover:text-foreground'}">
          {showAdvanced ? 'Hide advanced' : 'Show advanced'}
        </button>
      </div>

      <!-- ── Batting ──────────────────────────────────────────────────────── -->
      {#if activeTab === 'batting'}
        {#if battingLoading}
          <p class="mt-4 font-monospace text-[0.78rem] text-muted">Loading…</p>
        {:else if battingSeasons.length === 0}
          <p class="mt-4 font-monospace text-[0.78rem] text-muted">No batting data found for this player.</p>
        {:else}
          <!-- Chart -->
          <div class="mb-4 rounded-lg border border-outline bg-crust p-4">
            <div class="mb-3 flex items-center gap-3">
              <span class="panel-label">Career batting</span>
              <select
                value={batStat}
                onchange={(e) => setUrlParams({ stat: (e.target as HTMLSelectElement).value })}
                class="ml-auto rounded border border-outline bg-surface px-2 py-1 font-monospace text-[0.72rem] text-muted focus:outline-none">
                {#each BATTING_STATS as s (s.value)}
                  <option value={s.value}>{s.label}</option>
                {/each}
              </select>
            </div>
            <Chart config={battingChartConfig} height={110} />
          </div>
          <!-- Season table -->
          <div class="rounded-lg border border-outline bg-crust p-4">
            <div class="panel-label mb-3">Season log</div>
            <div class="overflow-x-auto">
              <table class="w-full border-collapse text-[0.75rem]">
                <thead>
                  <tr>
                    {#each ['Year', 'Team', 'Era', 'G', 'AB', 'H', 'HR', 'RBI', 'AVG', 'SB', 'OBP', 'SLG'] as col (col)}
                      <th
                        class="border-b border-outline px-2 py-1.5 text-left font-body text-[0.72rem] font-medium whitespace-nowrap text-muted">
                        {col}
                      </th>
                    {/each}
                  </tr>
                </thead>
                <tbody>
                  {#each battingRows as row (`${row.year}-${row.team}`)}
                    {@const era = eraForYear(row.year)}
                    <tr class="border-b border-outline last:border-b-0 hover:[&>td]:bg-surface">
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">
                        <a href={resolve(`/seasons?year=${row.year}`)} class="text-primary hover:underline">
                          {row.year}
                        </a>
                      </td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{row.team}</td>
                      <td class="px-2 py-1.5">
                        {#if era}<EraBadge {era} size="xs" />{:else}<span class="text-muted">—</span>{/if}
                      </td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{row.g}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{row.ab}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{row.h}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{row.hr}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{row.rbi}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{fmtAvg(row.avg)}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{fmtNum(row.sb)}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{fmtAvg(row.obp)}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{fmtAvg(row.slg)}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            {#if battingTotal > perPage}
              <div class="mt-4">
                <Pagination bind:page={ui.page} bind:perPage={ui.perPage} total={battingTotal} />
              </div>
            {/if}
          </div>
        {/if}

        <!-- ── Pitching ─────────────────────────────────────────────────────── -->
      {:else if activeTab === 'pitching'}
        {#if pitchingLoading}
          <p class="mt-4 font-monospace text-[0.78rem] text-muted">Loading…</p>
        {:else if pitchingSeasons.length === 0}
          <p class="mt-4 font-monospace text-[0.78rem] text-muted">No pitching data found for this player.</p>
        {:else}
          <div class="mb-4 rounded-lg border border-outline bg-crust p-4">
            <div class="panel-label mb-3">Career ERA</div>
            <Chart config={pitchingChartConfig} height={110} />
          </div>
          <div class="rounded-lg border border-outline bg-crust p-4">
            <div class="panel-label mb-3">Season log</div>
            <div class="overflow-x-auto">
              <table class="w-full border-collapse text-[0.75rem]">
                <thead>
                  <tr>
                    {#each ['Year', 'Team', 'Era', 'G', 'GS', 'W', 'L', 'SV', 'IP', 'SO', 'BB', 'ERA', 'WHIP'] as col (col)}
                      <th
                        class="border-b border-outline px-2 py-1.5 text-left font-body text-[0.72rem] font-medium whitespace-nowrap text-muted">
                        {col}
                      </th>
                    {/each}
                  </tr>
                </thead>
                <tbody>
                  {#each pitchingRows as row (`${row.year}-${row.team}`)}
                    {@const era = eraForYear(row.year)}
                    <tr class="border-b border-outline last:border-b-0 hover:[&>td]:bg-surface">
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">
                        <a href={resolve(`/seasons?year=${row.year}`)} class="text-primary hover:underline">
                          {row.year}
                        </a>
                      </td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{row.team}</td>
                      <td class="px-2 py-1.5">
                        {#if era}<EraBadge {era} size="xs" />{:else}<span class="text-muted">—</span>{/if}
                      </td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{row.g}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{fmtNum(row.gs)}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{row.w}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{row.l}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{fmtNum(row.sv)}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{row.ip}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{fmtNum(row.so)}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">{fmtNum(row.bb)}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground"
                        >{Number(row.era).toFixed(2)}</td>
                      <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">
                        {row.whip != null ? Number(row.whip).toFixed(2) : '—'}
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            {#if pitchingTotal > perPage}
              <div class="mt-4">
                <Pagination bind:page={ui.page} bind:perPage={ui.perPage} total={pitchingTotal} />
              </div>
            {/if}
          </div>
        {/if}

        <!-- ── Game Logs ────────────────────────────────────────────────────── -->
      {:else if activeTab === 'game-logs'}
        <!-- Sub-type selector -->
        <div class="mb-4 flex gap-1">
          {#each ['batting', 'pitching', 'fielding'] as type (type)}
            <button
              onclick={() => setUrlParams({ log: type, page: 1 })}
              class="rounded border px-3 py-1.25 font-display text-[0.78rem] transition-all {gameLogType === type
                ? 'border-outline bg-surface text-foreground'
                : 'border-transparent text-muted hover:text-foreground'}">
              {type.charAt(0).toUpperCase() + type.slice(1)}
            </button>
          {/each}
        </div>
        <div class="rounded-lg border border-outline bg-crust p-4">
          {#if gameLogsLoading}
            <p class="font-monospace text-[0.72rem] text-muted">Loading…</p>
          {:else if gameLogs.length === 0}
            <p class="font-monospace text-[0.78rem] text-muted">No {gameLogType} game logs found.</p>
          {:else}
            <div class="panel-label mb-3">
              {gameLogType.charAt(0).toUpperCase() + gameLogType.slice(1)} game logs
            </div>
            <SortableTable columns={gameLogColumns} rows={gameLogs} />
            {#if gameLogsTotal > perPage}
              <div class="mt-4">
                <Pagination bind:page={ui.page} bind:perPage={ui.perPage} total={gameLogsTotal} />
              </div>
            {/if}
          {/if}
        </div>

        <!-- ── Awards ──────────────────────────────────────────────────────── -->
      {:else if activeTab === 'awards'}
        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-3">Awards timeline</div>
          {#if awardsLoading}
            <p class="font-monospace text-[0.72rem] text-muted">Loading…</p>
          {:else if awards.length === 0}
            <p class="font-monospace text-[0.78rem] text-muted">No awards on record.</p>
          {:else}
            <div class="flex flex-col gap-2">
              {#each awards as award (`${award.year}-${award.name ?? award.award_id}`)}
                {@const era = eraForYear(award.year)}
                <div class="flex items-center gap-3 rounded-md bg-surface px-3 py-2.5">
                  <a
                    href={resolve(`/seasons?year=${award.year}`)}
                    class="min-w-11 font-monospace text-[0.72rem] text-muted hover:text-primary">
                    {award.year}
                  </a>
                  {#if era}<EraBadge {era} size="xs" />{/if}
                  <span class="font-display text-[0.82rem] text-foreground">{award.name ?? award.award_id}</span>
                  {#if award.league}
                    <span class="ml-auto font-monospace text-[0.68rem] text-muted">{award.league}</span>
                  {/if}
                  {#if award.notes}
                    <span class="font-monospace text-[0.68rem] text-muted">{award.notes}</span>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>

        <!-- ── Hall of Fame ────────────────────────────────────────────────── -->
      {:else if activeTab === 'hof'}
        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-3">Hall of Fame voting history</div>
          {#if hofLoading}
            <p class="font-monospace text-[0.72rem] text-muted">Loading…</p>
          {:else if hofEntries.length === 0}
            <p class="font-monospace text-[0.78rem] text-muted">No Hall of Fame data on record.</p>
          {:else}
            {#if hofEntries.some((e) => e.pct != null)}
              <div class="mb-4">
                <Chart config={hofChartConfig} height={100} />
              </div>
            {/if}
            <SortableTable
              columns={[
                { key: 'year_inducted', label: 'Year', sortable: true },
                { key: 'inducted', label: 'Inducted', format: (v) => (v ? 'Yes' : 'No') },
                { key: 'votes', label: 'Votes', sortable: true },
                { key: 'ballots', label: 'Ballots', sortable: true },
                {
                  key: 'pct',
                  label: 'Vote %',
                  sortable: true,
                  format: (v) => (v != null ? Number(v).toFixed(1) + '%' : '—')
                },
                { key: 'category', label: 'Category' }
              ]}
              rows={hofEntries} />
          {/if}
        </div>

        <!-- ── Teams ───────────────────────────────────────────────────────── -->
      {:else if activeTab === 'teams'}
        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-3">Teams</div>
          {#if teamsLoading}
            <p class="font-monospace text-[0.72rem] text-muted">Loading…</p>
          {:else if teams.length === 0}
            <p class="font-monospace text-[0.78rem] text-muted">No team data on record.</p>
          {:else}
            <SortableTable
              columns={teamsColumns}
              rows={teams.map((t) => ({ ...t, team: t.team ?? t.team_id ?? '?' }))} />
          {/if}
        </div>

        <!-- ── Salaries ────────────────────────────────────────────────────── -->
      {:else if activeTab === 'salaries'}
        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-3">Salaries</div>
          {#if salariesLoading}
            <p class="font-monospace text-[0.72rem] text-muted">Loading…</p>
          {:else if salaries.length === 0}
            <p class="font-monospace text-[0.78rem] text-muted">No salary data on record.</p>
          {:else}
            <SortableTable
              columns={salariesColumns}
              rows={salaries.map((s) => ({ ...s, team: s.team ?? s.team_id ?? '?' }))} />
          {/if}
        </div>

        <!-- ── Relatives ───────────────────────────────────────────────────── -->
      {:else if activeTab === 'relatives'}
        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-3">Relatives</div>
          {#if relativesLoading}
            <p class="font-monospace text-[0.72rem] text-muted">Loading…</p>
          {:else if relatives.length === 0}
            <p class="font-monospace text-[0.78rem] text-muted">No relatives on record.</p>
          {:else}
            <div class="flex flex-col gap-2">
              {#each relatives as rel (rel.player_id ?? rel.name)}
                <div class="flex items-center gap-3 rounded-md bg-surface px-3 py-2.5">
                  {#if rel.player_id}
                    <button
                      onclick={() => selectPlayer(rel.player_id!)}
                      class="font-display text-[0.82rem] text-primary hover:underline">
                      {rel.name ?? rel.player_id}
                    </button>
                  {:else}
                    <span class="font-display text-[0.82rem] text-foreground">{rel.name ?? '?'}</span>
                  {/if}
                  {#if rel.relationship}
                    <span class="ml-auto font-monospace text-[0.68rem] text-muted">{rel.relationship}</span>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>

        <!-- ── Advanced tabs ───────────────────────────────────────────────── -->
      {:else if activeTab === 'batting-adv'}
        <AdvancedTabPanel endpoint={EP.playerStatsBattingAdv(playerId)} label="Advanced Batting" />
      {:else if activeTab === 'pitching-adv'}
        <AdvancedTabPanel endpoint={EP.playerStatsPitchingAdv(playerId)} label="Advanced Pitching" />
      {:else if activeTab === 'war'}
        <AdvancedTabPanel endpoint={EP.playerStatsWar(playerId)} label="WAR" />
      {:else if activeTab === 'splits'}
        <AdvancedTabPanel endpoint={EP.playerSplits(playerId)} label="Splits" />
      {:else if activeTab === 'streaks'}
        <AdvancedTabPanel endpoint={EP.playerStreaks(playerId)} label="Streaks" />
      {/if}
    {/if}
  {/snippet}

  {#snippet panel()}
    <ApiPanel endpoint={`/v1${activeEndpoint}`} url={activeUrl} />
  {/snippet}
</ThreeColLayout>
