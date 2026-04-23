<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { apiFetch, fetchPaginated, type PaginatedResponse } from '$lib/api';
  import { parseLeague } from '$lib/common/types';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import StandingsPanel from '$lib/components/StandingsPanel.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear } from '$lib/eras';
  import { emptyPage, fmtInt, fmtSigned, toErrorMessage } from '$lib/leaders/utils';
  import { normalizeSeasons, normalizeSeasonTeamsPage } from '$lib/seasons/normalizers';
  import type { SeasonSummary, SeasonTeam } from '$lib/seasons/types';
  import { onMount } from 'svelte';

  const CURRENT_YEAR = new Date().getFullYear();

  let routeYear = $derived(Number(page.params.year ?? ''));
  let selectedYear = $derived(Number.isFinite(routeYear) && routeYear > 0 ? routeYear : CURRENT_YEAR);
  let rawLeague = $derived(page.url.searchParams.get('league'));
  let leagueFilter = $derived(parseLeague(rawLeague));

  let seasons = $state<SeasonSummary[]>([]);
  let teamsPage = $state<PaginatedResponse<SeasonTeam>>(emptyPage());

  let seasonsLoading = $state(false);
  let teamsLoading = $state(false);
  let teamsError = $state<string | null>(null);

  let seasonsRequestVersion = 0;
  let teamsRequestVersion = 0;
  let lastTeamsKey = '';

  let selectedSeason = $derived(seasons.find((season) => season.year === selectedYear) ?? null);
  let selectedEra = $derived(eraForYear(selectedYear));
  let isCurrentSeason = $derived(selectedYear === CURRENT_YEAR);

  const filteredTeams = $derived.by(() => {
    if (leagueFilter === 'both') return teamsPage.data;
    const league = leagueFilter.toUpperCase();
    return teamsPage.data.filter((team) => (team.league ?? '').toUpperCase() === league);
  });

  const teamRows = $derived(
    filteredTeams
      .map((team) => {
        const pct = winningPct(team.wins, team.losses);
        const runDiff =
          team.runs_scored != null && team.runs_allowed != null ? team.runs_scored - team.runs_allowed : undefined;

        return {
          team_id: team.team_id,
          name: team.name ?? team.team_id,
          league: team.league ?? '—',
          division: team.division ?? '—',
          wins: team.wins,
          losses: team.losses,
          ties: team.ties,
          pct,
          runs_scored: team.runs_scored,
          runs_allowed: team.runs_allowed,
          run_diff: runDiff,
          games: team.games
        };
      })
      .toSorted((a, b) => {
        const pctA = a.pct ?? -1;
        const pctB = b.pct ?? -1;
        if (pctA !== pctB) return pctB - pctA;

        const winsA = a.wins ?? -1;
        const winsB = b.wins ?? -1;
        if (winsA !== winsB) return winsB - winsA;

        return a.team_id.localeCompare(b.team_id);
      })
  );

  const teamColumns = [
    {
      key: 'team_id',
      label: 'Team',
      sortable: true,
      href: (value: unknown) => {
        const teamId = String(value ?? '').trim();
        if (!teamId) return;
        const query = new URLSearchParams({ q: teamId, year: String(selectedYear) }).toString();
        return `/teams?${query}`;
      },
      tooltip: (_value: unknown, row: Record<string, unknown>) => {
        const teamName = String(row.name ?? '').trim();
        return teamName.length > 0 ? teamName : undefined;
      }
    },
    { key: 'league', label: 'Lg', sortable: true },
    { key: 'division', label: 'Div', sortable: true },
    { key: 'wins', label: 'W', sortable: true, format: (value: unknown) => fmtInt(value as number | undefined) },
    { key: 'losses', label: 'L', sortable: true, format: (value: unknown) => fmtInt(value as number | undefined) },
    { key: 'pct', label: 'Pct', sortable: true, format: (value: unknown) => fmtRate(value as number | undefined) },
    {
      key: 'runs_scored',
      label: 'RS',
      sortable: true,
      format: (value: unknown) => fmtInt(value as number | undefined)
    },
    {
      key: 'runs_allowed',
      label: 'RA',
      sortable: true,
      format: (value: unknown) => fmtInt(value as number | undefined)
    },
    {
      key: 'run_diff',
      label: 'Diff',
      sortable: true,
      format: (value: unknown) => fmtSigned(value as number | undefined)
    }
  ];

  onMount(() => {
    void loadSeasons(true);
    void refreshTeams(true);
  });

  afterNavigate(() => {
    void loadSeasons();
    void refreshTeams();
  });

  async function loadSeasons(force = false): Promise<void> {
    if (!force && (seasonsLoading || seasons.length > 0)) return;

    const requestVersion = ++seasonsRequestVersion;
    seasonsLoading = true;

    try {
      const payload = await apiFetch<unknown>(EP.seasons);
      if (requestVersion !== seasonsRequestVersion) return;
      seasons = normalizeSeasons(payload);
    } catch {
      if (requestVersion !== seasonsRequestVersion) return;
    } finally {
      if (requestVersion === seasonsRequestVersion) seasonsLoading = false;
    }
  }

  async function refreshTeams(force = false): Promise<void> {
    const key = `${selectedYear}|${leagueFilter}`;
    if (!force && key === lastTeamsKey) return;
    lastTeamsKey = key;

    const requestVersion = ++teamsRequestVersion;
    teamsLoading = true;
    teamsError = null;

    try {
      const params: Record<string, string | number> = { page: 1, per_page: 80 };
      if (leagueFilter !== 'both') params.league = leagueFilter.toUpperCase();

      const payload = await fetchPaginated<unknown>(EP.seasonTeams(selectedYear), params);
      if (requestVersion !== teamsRequestVersion) return;
      teamsPage = normalizeSeasonTeamsPage(payload);
    } catch (error) {
      if (requestVersion !== teamsRequestVersion) return;
      teamsError = toErrorMessage(error, 'Failed to load season teams.');
      teamsPage = emptyPage();
    } finally {
      if (requestVersion === teamsRequestVersion) teamsLoading = false;
    }
  }

  function winningPct(wins: number | undefined, losses: number | undefined): number | undefined {
    if (wins == null || losses == null) return undefined;
    const decisions = wins + losses;
    if (decisions <= 0) return undefined;
    return wins / decisions;
  }

  function fmtRate(value: number | undefined): string {
    if (value == null) return '—';
    return value.toFixed(3);
  }
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="flex flex-wrap items-center gap-4">
    <div class="font-display text-[1.05rem] text-foreground">{selectedYear} Season</div>
    {#if selectedEra}
      <EraBadge era={selectedEra} size="xs" />
    {/if}
    <div class="font-mono text-[0.7rem] text-muted">
      {#if selectedSeason}
        {selectedSeason.leagues.length > 0 ? selectedSeason.leagues.join(' + ') : '—'}
        · {selectedSeason.team_count ?? '—'} teams · {selectedSeason.game_count != null
          ? selectedSeason.game_count.toLocaleString()
          : '—'} games
      {:else}
        Coverage metadata unavailable for this year.
      {/if}
    </div>
  </div>
</div>

{#if isCurrentSeason}
  <div class="my-2 rounded-lg border border-primary/35 bg-primary/10 px-3 py-2 text-xs text-primary">
    Data refreshes every 4 hours for season stats. Standings and schedule update hourly.
  </div>
{/if}

<StandingsPanel season={selectedYear} title="Standings" showEndpointHint={false} />

<section class="mt-3 rounded-lg border border-outline bg-crust p-4">
  <div class="mb-3 flex items-center justify-between gap-2">
    <div class="panel-label mb-0 border-0 p-0">Teams</div>
    <div class="font-mono text-xxs tracking-[0.08em] text-muted uppercase">
      {leagueFilter === 'both' ? 'AL + NL' : leagueFilter.toUpperCase()}
    </div>
  </div>

  {#if teamsError}
    <p class="mb-3 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
      {teamsError}
    </p>
  {/if}

  {#if teamsLoading}
    <p class="font-mono text-xs text-muted">Loading season teams...</p>
  {:else if teamRows.length === 0}
    <p class="font-mono text-xs text-muted">No team rows for this season and league filter.</p>
  {:else}
    <SortableTable columns={teamColumns} rows={teamRows} />
  {/if}
</section>
