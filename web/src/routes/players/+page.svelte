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
  import { PlayersUiController } from '$lib/players.svelte';
  import type { ApiPlayerPayload, PlayerStatsPayload } from '$lib/players/api-payloads';
  import {
    createBattingChartConfig,
    createHallOfFameChartConfig,
    createPitchingChartConfig
  } from '$lib/players/charts';
  import { ADV_TABS, BATTING_STATS, MAIN_TABS, SALARIES_COLUMNS, TEAMS_COLUMNS } from '$lib/players/constants';
  import {
    normalizeHallOfFamePayload,
    normalizePlayerProfile,
    normalizePlayerStatsPage,
    normalizeSearchPlayersPage
  } from '$lib/players/normalizers';
  import { AsyncListResource, AsyncPaginatedListResource, AsyncValueResource } from '$lib/players/resources.svelte';
  import type {
    ApiListPayload,
    Award,
    BattingSeason,
    GameLog,
    HofEntry,
    PitchingSeason,
    PlayerProfile,
    PlayerResult,
    PlayerTeam,
    Relative,
    Salary
  } from '$lib/players/types';
  import { normalizeApiList, rowColumns } from '$lib/players/types';
  import { intParam, setUrlParams } from '$lib/url-state.svelte';
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
  let mounted = $state(false);

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

  async function fetchList<T>(endpoint: string): Promise<T[]> {
    const payload = await apiFetch<ApiListPayload<T>>(endpoint);
    return normalizeApiList(payload);
  }

  let careerEras = $derived.by(() => {
    if (!profileResource.value?.debut_year || !profileResource.value?.final_year) return [];
    return erasInRange(profileResource.value.debut_year, profileResource.value.final_year);
  });

  let battingRows = $derived(battingResource.items.map((s) => ({ ...s, team: s.team ?? s.team_id ?? '?' })));
  let pitchingRows = $derived(pitchingResource.items.map((s) => ({ ...s, team: s.team ?? s.team_id ?? '?' })));

  let gameLogColumns = $derived(rowColumns(gameLogsResource.items));

  let battingChartConfig = $derived(createBattingChartConfig(battingRows, batStat));
  let pitchingChartConfig = $derived(createPitchingChartConfig(pitchingRows));
  let hofChartConfig = $derived(createHallOfFameChartConfig(hofResource.items));

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
        return getLogEndpoint(gameLogType, playerId);
      default:
        return EP.player(playerId);
    }
  });

  let activeUrl = $derived(apiUrl(activeEndpoint));
  let allTabs = $derived(showAdvanced ? [...MAIN_TABS, ...ADV_TABS] : MAIN_TABS);

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

  $effect(() => {
    const thisQ = q;
    if (!thisQ) {
      searchResource.clear();
      return;
    }
    void searchResource.load(async () => {
      const payload = await fetchPaginated<ApiPlayerPayload>(EP.searchPlayers, { q: thisQ, per_page: 20 });
      return normalizeSearchPlayersPage(payload);
    });
  });

  $effect(() => {
    const id = playerId;
    if (!id) {
      profileResource.clear();
      return;
    }
    void profileResource.load(async () => {
      const payload = await apiFetch<ApiPlayerPayload>(EP.player(id));
      return normalizePlayerProfile(payload);
    });
  });

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    const p = currentPage;
    const pp = perPage;
    if (!id || tab !== 'batting') return;
    void battingResource.load(async () => {
      const payload = await apiFetch<PlayerStatsPayload<BattingSeason>>(EP.playerStatsBatting(id));
      return normalizePlayerStatsPage(payload, p, pp);
    });
  });

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    const p = currentPage;
    const pp = perPage;
    if (!id || tab !== 'pitching') return;
    void pitchingResource.load(async () => {
      const payload = await apiFetch<PlayerStatsPayload<PitchingSeason>>(EP.playerStatsPitching(id));
      return normalizePlayerStatsPage(payload, p, pp);
    });
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
    void hofResource.load(async () => {
      const payload = await apiFetch<{ records?: HofEntry[] | null; data?: HofEntry[] }>(EP.playerHallOfFame(id));
      return normalizeHallOfFamePayload(payload);
    });
  });

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    if (!id || tab !== 'teams') return;
    void teamsResource.load(() => fetchList<PlayerTeam>(EP.playerTeams(id)));
  });

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    if (!id || tab !== 'salaries') return;
    void salariesResource.load(() => fetchList<Salary>(EP.playerSalaries(id)));
  });

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    if (!id || tab !== 'relatives') return;
    void relativesResource.load(() => fetchList<Relative>(EP.playerRelatives(id)));
  });

  $effect(() => {
    const id = playerId;
    const tab = activeTab;
    const logType = gameLogType;
    const p = currentPage;
    const pp = perPage;
    if (!id || tab !== 'game-logs') return;

    const logEP = getLogEndpoint(logType, id);
    void gameLogsResource.load(() => fetchPaginated<GameLog>(logEP, { page: p, per_page: pp }));
  });

  function getLogEndpoint(kind: string, id: string): string {
    switch (kind) {
      case 'pitching':
        return EP.playerGameLogsPitching(id);
      case 'fielding':
        return EP.playerGameLogsFielding(id);
      default:
        return EP.playerGameLogsBatting(id);
    }
  }

  function handleSearch(value = ui.search) {
    const trimmed = value.trim();
    if (!trimmed) return;
    void setUrlParams({ q: trimmed, id: null, tab: 'batting', page: 1 });
  }

  function selectPlayer(id: string) {
    void setUrlParams({ id, tab: 'batting', page: 1, stat: null });
  }

  const fmtAvg = (v: number | undefined) => (v != null ? Number(v).toFixed(3) : '—');

  const fmtNum = (v: number | undefined) => (v != null ? String(v) : '—');
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Player search</div>
    <SearchInput mini bind:value={ui.search} placeholder="Name, ID…" onsubmit={handleSearch} />

    {#if playerId}
      <div class="mt-4 rounded-lg border border-outline bg-surface p-4">
        {#if profileResource.loading}
          <p class="font-mono text-xs text-muted">Loading…</p>
        {:else if profileResource.error}
          <p class="font-mono text-xs text-warning">{profileResource.error}</p>
        {:else if profileResource.value}
          <div class="mb-0.5 text-center font-mono text-[0.6rem] tracking-wider text-muted uppercase">
            {profileResource.value.primary_position ?? profileResource.value.positions?.[0] ?? ''}
          </div>
          <div class="mb-1 text-center font-display text-[0.95rem] font-medium text-foreground">
            {profileResource.value.name}
          </div>
          {#if profileResource.value.birth_date || profileResource.value.birth_city}
            <div class="mb-1 text-center font-mono text-[0.68rem] text-muted">
              {profileResource.value.birth_date ?? ''}
              {#if profileResource.value.birth_city}
                · {profileResource.value.birth_city}{#if profileResource.value.birth_state}, {profileResource.value
                    .birth_state}{/if}{/if}
            </div>
          {/if}
          {#if profileResource.value.bats || profileResource.value.throws || profileResource.value.debut_year}
            <div class="mb-3 text-center font-mono text-[0.68rem] text-muted">
              {#if profileResource.value.bats}Bats: {profileResource.value.bats}{/if}
              {#if profileResource.value.bats && profileResource.value.throws}
                ·
              {/if}
              {#if profileResource.value.throws}Throws: {profileResource.value.throws}{/if}
              {#if profileResource.value.debut_year}
                · {profileResource.value.debut_year}–{profileResource.value.final_year ?? 'pres.'}{/if}
            </div>
          {/if}
          {#if profileResource.value.career_hr != null || profileResource.value.career_avg != null || profileResource.value.career_rbi != null}
            <div class="mb-3 grid grid-cols-3 gap-1 text-center">
              {#if profileResource.value.career_hr != null}
                <div>
                  <div class="font-display text-[0.95rem] font-semibold text-foreground">
                    {profileResource.value.career_hr}
                  </div>
                  <div class="font-mono text-[0.58rem] text-muted uppercase">HR</div>
                </div>
              {/if}
              {#if profileResource.value.career_avg != null}
                <div>
                  <div class="font-display text-[0.95rem] font-semibold text-foreground">
                    {fmtAvg(profileResource.value.career_avg)}
                  </div>
                  <div class="font-mono text-[0.58rem] text-muted uppercase">AVG</div>
                </div>
              {/if}
              {#if profileResource.value.career_rbi != null}
                <div>
                  <div class="font-display text-[0.95rem] font-semibold text-foreground">
                    {profileResource.value.career_rbi}
                  </div>
                  <div class="font-mono text-[0.58rem] text-muted uppercase">RBI</div>
                </div>
              {/if}
            </div>
          {/if}

          {#if careerEras.length > 0}
            <div class="border-t border-outline pt-3">
              <div class="mb-1.5 font-mono text-[0.6rem] tracking-wider text-muted uppercase">Career eras</div>
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

    {#if searchResource.items.length > 0}
      <div class="mt-4 panel-label">Results</div>
      <div class="flex flex-col gap-0.5">
        {#each searchResource.items as result (result.id)}
          <button
            onclick={() => selectPlayer(result.id)}
            class="rounded-md px-3 py-2 text-left transition-colors hover:bg-surface {playerId === result.id
              ? 'bg-surface'
              : ''}">
            <div class="font-display text-[0.8rem] text-foreground">{result.name}</div>
            <div class="font-mono text-[0.68rem] text-muted">
              {result.id}
              {#if result.primary_position ?? result.position}
                · {result.primary_position ?? result.position}{/if}
              {#if result.debut_year}
                · {result.debut_year}–{result.final_year ?? 'pres.'}{/if}
            </div>
          </button>
        {/each}
      </div>
    {:else if searchResource.loading}
      <p class="mt-3 font-mono text-xs text-muted">Searching…</p>
    {:else if !q && !playerId}
      <p class="mt-4 font-mono text-xs text-muted">Search for a player to begin.</p>
    {/if}
  {/snippet}

  {#snippet center()}
    {#if playerId}
      <div class="mb-1 flex flex-wrap items-center gap-2">
        <TabRow tabs={allTabs} bind:active={ui.tab} />
        <button
          onclick={() =>
            setUrlParams({ advanced: showAdvanced ? null : '1', tab: showAdvanced ? 'batting' : activeTab })}
          class="ml-auto shrink-0 rounded border px-2.5 py-1 font-mono text-[0.68rem] transition-colors {showAdvanced
            ? 'border-primary/40 text-primary hover:border-primary'
            : 'border-outline text-muted hover:border-primary hover:text-foreground'}">
          {showAdvanced ? 'Hide advanced' : 'Show advanced'}
        </button>
      </div>
      {#if activeTab === 'batting'}
        {#if battingResource.loading}
          <p class="mt-4 font-mono text-[0.78rem] text-muted">Loading…</p>
        {:else if battingResource.items.length === 0}
          <p class="mt-4 font-mono text-[0.78rem] text-muted">No batting data found for this player.</p>
        {:else}
          <div class="mb-4 rounded-lg border border-outline bg-crust p-4">
            <div class="mb-3 flex items-center gap-3">
              <span class="panel-label">Career batting</span>
              <select
                value={batStat}
                onchange={(e) => setUrlParams({ stat: (e.target as HTMLSelectElement).value })}
                class="ml-auto rounded border border-outline bg-surface px-2 py-1 font-mono text-xs text-muted focus:outline-none">
                {#each BATTING_STATS as s (s.value)}
                  <option value={s.value}>{s.label}</option>
                {/each}
              </select>
            </div>
            <Chart config={battingChartConfig} height={110} />
          </div>
          <div class="rounded-lg border border-outline bg-crust p-4">
            <div class="panel-label mb-3">Season log</div>
            <div class="overflow-x-auto">
              <table class="w-full border-collapse text-[0.75rem]">
                <thead>
                  <tr>
                    {#each ['Year', 'Team', 'Era', 'G', 'AB', 'H', 'HR', 'RBI', 'AVG', 'SB', 'OBP', 'SLG'] as col (col)}
                      <th
                        class="border-b border-outline px-2 py-1.5 text-left font-sans text-xs font-medium whitespace-nowrap text-muted">
                        {col}
                      </th>
                    {/each}
                  </tr>
                </thead>
                <tbody>
                  {#each battingRows as row (`${row.year}-${row.team}`)}
                    {@const era = eraForYear(row.year)}
                    <tr class="border-b border-outline last:border-b-0 hover:[&>td]:bg-surface">
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">
                        <a href={resolve(`/seasons?year=${row.year}`)} class="text-primary hover:underline">
                          {row.year}
                        </a>
                      </td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.team}</td>
                      <td class="px-2 py-1.5">
                        {#if era}<EraBadge {era} size="xs" />{:else}<span class="text-muted">—</span>{/if}
                      </td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.g}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.ab}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.h}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.hr}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.rbi}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtAvg(row.avg)}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtNum(row.sb)}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtAvg(row.obp)}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtAvg(row.slg)}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            {#if battingResource.total > perPage}
              <div class="mt-4">
                <Pagination bind:page={ui.page} bind:perPage={ui.perPage} total={battingResource.total} />
              </div>
            {/if}
          </div>
        {/if}
      {:else if activeTab === 'pitching'}
        {#if pitchingResource.loading}
          <p class="mt-4 font-mono text-[0.78rem] text-muted">Loading…</p>
        {:else if pitchingResource.items.length === 0}
          <p class="mt-4 font-mono text-[0.78rem] text-muted">No pitching data found for this player.</p>
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
                        class="border-b border-outline px-2 py-1.5 text-left font-sans text-xs font-medium whitespace-nowrap text-muted">
                        {col}
                      </th>
                    {/each}
                  </tr>
                </thead>
                <tbody>
                  {#each pitchingRows as row (`${row.year}-${row.team}`)}
                    {@const era = eraForYear(row.year)}
                    <tr class="border-b border-outline last:border-b-0 hover:[&>td]:bg-surface">
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">
                        <a href={resolve(`/seasons?year=${row.year}`)} class="text-primary hover:underline">
                          {row.year}
                        </a>
                      </td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.team}</td>
                      <td class="px-2 py-1.5">
                        {#if era}<EraBadge {era} size="xs" />{:else}<span class="text-muted">—</span>{/if}
                      </td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.g}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtNum(row.gs)}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.w}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.l}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtNum(row.sv)}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.ip}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtNum(row.so)}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtNum(row.bb)}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">{Number(row.era).toFixed(2)}</td>
                      <td class="px-2 py-1.5 font-mono text-xs text-foreground">
                        {row.whip != null ? Number(row.whip).toFixed(2) : '—'}
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            {#if pitchingResource.total > perPage}
              <div class="mt-4">
                <Pagination bind:page={ui.page} bind:perPage={ui.perPage} total={pitchingResource.total} />
              </div>
            {/if}
          </div>
        {/if}
      {:else if activeTab === 'game-logs'}
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
          {#if gameLogsResource.loading}
            <p class="font-mono text-xs text-muted">Loading…</p>
          {:else if gameLogsResource.items.length === 0}
            <p class="font-mono text-[0.78rem] text-muted">No {gameLogType} game logs found.</p>
          {:else}
            <div class="panel-label mb-3">
              {gameLogType.charAt(0).toUpperCase() + gameLogType.slice(1)} game logs
            </div>
            <SortableTable columns={gameLogColumns} rows={gameLogsResource.items} />
            {#if gameLogsResource.total > perPage}
              <div class="mt-4">
                <Pagination bind:page={ui.page} bind:perPage={ui.perPage} total={gameLogsResource.total} />
              </div>
            {/if}
          {/if}
        </div>
      {:else if activeTab === 'awards'}
        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-3">Awards timeline</div>
          {#if awardsResource.loading}
            <p class="font-mono text-xs text-muted">Loading…</p>
          {:else if awardsResource.items.length === 0}
            <p class="font-mono text-[0.78rem] text-muted">No awards on record.</p>
          {:else}
            <div class="flex flex-col gap-2">
              {#each awardsResource.items as award, idx (`${award.year}-${award.name ?? award.award_id}-${award.league ?? ''}-${award.notes ?? ''}-${idx}`)}
                {@const era = eraForYear(award.year)}
                <div class="flex items-center gap-3 rounded-md bg-surface px-3 py-2.5">
                  <a
                    href={resolve(`/seasons?year=${award.year}`)}
                    class="min-w-11 font-mono text-xs text-muted hover:text-primary">
                    {award.year}
                  </a>
                  {#if era}<EraBadge {era} size="xs" />{/if}
                  <span class="font-display text-[0.82rem] text-foreground">{award.name ?? award.award_id}</span>
                  {#if award.league}
                    <span class="ml-auto font-mono text-[0.68rem] text-muted">{award.league}</span>
                  {/if}
                  {#if award.notes}
                    <span class="font-mono text-[0.68rem] text-muted">{award.notes}</span>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {:else if activeTab === 'hof'}
        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-3">Hall of Fame voting history</div>
          {#if hofResource.loading}
            <p class="font-mono text-xs text-muted">Loading…</p>
          {:else if hofResource.items.length === 0}
            <p class="font-mono text-xs text-muted">No Hall of Fame data on record.</p>
          {:else}
            {#if hofResource.items.some((e) => e.pct != null)}
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
              rows={hofResource.items} />
          {/if}
        </div>
      {:else if activeTab === 'teams'}
        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-3">Teams</div>
          {#if teamsResource.loading}
            <p class="font-mono text-xs text-muted">Loading…</p>
          {:else if teamsResource.items.length === 0}
            <p class="font-mono text-[0.78rem] text-muted">No team data on record.</p>
          {:else}
            <SortableTable
              columns={TEAMS_COLUMNS}
              rows={teamsResource.items.map((t) => ({ ...t, team: t.team ?? t.team_id ?? '?' }))} />
          {/if}
        </div>
      {:else if activeTab === 'salaries'}
        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-3">Salaries</div>
          {#if salariesResource.loading}
            <p class="font-mono text-xs text-muted">Loading…</p>
          {:else if salariesResource.items.length === 0}
            <p class="font-mono text-[0.78rem] text-muted">No salary data on record.</p>
          {:else}
            <SortableTable
              columns={SALARIES_COLUMNS}
              rows={salariesResource.items.map((s) => ({ ...s, team: s.team ?? s.team_id ?? '?' }))} />
          {/if}
        </div>
      {:else if activeTab === 'relatives'}
        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-3">Relatives</div>
          {#if relativesResource.loading}
            <p class="font-mono text-xs text-muted">Loading…</p>
          {:else if relativesResource.items.length === 0}
            <p class="font-mono text-[0.78rem] text-muted">No relatives on record.</p>
          {:else}
            <div class="flex flex-col gap-2">
              {#each relativesResource.items as rel (rel.player_id ?? rel.name)}
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
                    <span class="ml-auto font-mono text-[0.68rem] text-muted">{rel.relationship}</span>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
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
    {:else}
      <div class="flex h-full flex-col items-center justify-center gap-3 text-center">
        <div class="font-display text-[1.1rem] text-muted">Player Explorer</div>
        <p class="max-w-xs font-mono text-[0.78rem] text-muted/60">
          Search for a player in the sidebar to explore career stats, awards, game logs, and more.
        </p>
        <div class="mt-2 flex flex-wrap justify-center gap-1.5">
          {#each STATIC_ERAS as era (era.code)}
            <EraRangeChip {era} />
          {/each}
        </div>
      </div>
    {/if}
  {/snippet}

  {#snippet panel()}
    <ApiPanel endpoint={`/v1${activeEndpoint}`} url={activeUrl} />
  {/snippet}
</ThreeColLayout>
