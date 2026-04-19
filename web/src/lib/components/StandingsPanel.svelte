<script lang="ts">
  import { resolve } from '$app/paths';
  import { apiFetch } from '$lib/api';
  import SegmentControl from '$lib/components/SegmentControl.svelte';
  import { EP } from '$lib/endpoints';
  import {
    buildMlbTeamAbbrByIDFromDetails,
    buildStandingsRows,
    extractSeasonFromStandings,
    normalizeFranchiseIDByMlbTeamID,
    sortStandingsRows,
    type StandingsRow,
    type StandingsSortKey
  } from '$lib/mlb/standings';
  import { teamPrimaryHexFor } from '$lib/mlb/team-branding';
  import { onMount } from 'svelte';

  type LeagueFilter = 'both' | 'AL' | 'NL';

  const DEFAULT_SEASON = new Date().getFullYear();
  const STANDINGS_HINT = `/v1${EP.mlbStandings}?season=${DEFAULT_SEASON}&standingsTypes=regularSeason&include=details`;

  const LEAGUE_OPTIONS = [
    { id: 'both', label: 'Both' },
    { id: 'AL', label: 'AL' },
    { id: 'NL', label: 'NL' }
  ];

  let loading = $state(true);
  let refreshing = $state(false);
  let errorMessage = $state<string | null>(null);
  let season = $state(DEFAULT_SEASON);
  let leagueFilter = $state<LeagueFilter>('both');
  let sort = $state<{ key: StandingsSortKey; dir: 'asc' | 'desc' }>({ key: 'rank', dir: 'asc' });

  let rows = $state<StandingsRow[]>([]);

  const filteredRows = $derived.by(() => {
    if (leagueFilter === 'both') return rows;
    return rows.filter((row) => row.league === leagueFilter);
  });

  type DivisionGroup = { division: string; rows: StandingsRow[] };

  const divisionGroups = $derived.by((): DivisionGroup[] => {
    const byDivision: Record<string, StandingsRow[]> = {};
    const orderedDivisions: string[] = [];

    for (const row of filteredRows) {
      const existing = byDivision[row.division];
      if (existing) {
        existing.push(row);
      } else {
        byDivision[row.division] = [row];
        orderedDivisions.push(row.division);
      }
    }

    return orderedDivisions.map((division) => {
      const groupRows = byDivision[division] ?? [];
      return { division, rows: sortStandingsRows(groupRows, sort.key, sort.dir) };
    });
  });

  onMount(() => {
    void refreshStandings('initial');
  });

  async function refreshStandings(mode: 'initial' | 'manual'): Promise<void> {
    if (mode === 'initial') {
      loading = true;
    } else {
      refreshing = true;
    }

    try {
      const [standingsPayload, crosswalkPayload] = await Promise.all([
        apiFetch<unknown>(EP.mlbStandings, {
          season: DEFAULT_SEASON,
          standingsTypes: 'regularSeason',
          include: 'details'
        }),
        apiFetch<unknown>(EP.metaCrosswalkTeams, { season: DEFAULT_SEASON, include: 'details' })
      ]);

      const teamAbbrByID = buildMlbTeamAbbrByIDFromDetails(standingsPayload);
      const franchiseIDByMlbTeamID = normalizeFranchiseIDByMlbTeamID(crosswalkPayload);
      rows = buildStandingsRows(standingsPayload, teamAbbrByID, {}, franchiseIDByMlbTeamID);
      season = extractSeasonFromStandings(standingsPayload) ?? DEFAULT_SEASON;
      errorMessage = null;
    } catch (error) {
      const message = error instanceof Error && error.message ? error.message : 'Failed to load current standings.';
      errorMessage = message;
    } finally {
      loading = false;
      refreshing = false;
    }
  }

  function toggleSort(key: StandingsSortKey): void {
    if (sort.key === key) {
      sort = { key, dir: sort.dir === 'asc' ? 'desc' : 'asc' };
      return;
    }
    const defaultDir = key === 'rank' || key === 'gb' || key === 'wcGb' || key === 'losses' ? 'asc' : 'desc';
    sort = { key, dir: defaultDir };
  }

  function arrowFor(key: StandingsSortKey): string {
    if (sort.key !== key) return '↕';
    if (sort.dir === 'asc') return '↑';
    return '↓';
  }

  function teamHref(row: StandingsRow): string {
    if (row.localFranchiseID) {
      const encodedFranchiseID = encodeURIComponent(row.localFranchiseID);
      const params = new URLSearchParams({ franchise_id: row.localFranchiseID, year: String(season) }).toString();
      return `/teams/${encodedFranchiseID}/overview?${params}`;
    }
    const params = new URLSearchParams({ q: row.teamName }).toString();
    return `/teams?${params}`;
  }

  function teamColor(row: StandingsRow): string {
    return teamPrimaryHexFor(row.teamAbbr) ?? '#4B5563';
  }

  function runDiffLabel(value: number): string {
    if (value > 0) return `+${value}`;
    return String(value);
  }

  function streakToneClass(streak: string): string {
    if (streak.startsWith('W')) return 'text-secondary';
    if (streak.startsWith('L')) return 'text-warning';
    return 'text-muted';
  }

  function leaderRowClass(isLeader: boolean): string {
    if (isLeader) return 'bg-primary/8';
    return '';
  }
</script>

<section class="rounded-lg border border-outline bg-crust p-4">
  <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
    <div>
      <h2 class="font-mono text-[0.74rem] tracking-[0.08em] text-muted uppercase">Current standings</h2>
      <p class="text-[0.75rem] text-muted">Season {season}</p>
      <p class="font-mono text-[0.63rem] text-muted">{STANDINGS_HINT}</p>
    </div>
    <button
      type="button"
      class="rounded border border-outline px-2.5 py-1 font-mono text-[0.64rem] text-foreground transition-colors hover:bg-surface"
      onclick={() => void refreshStandings('manual')}
      disabled={loading || refreshing}>
      {refreshing ? 'Refreshing…' : 'Refresh'}
    </button>
  </div>

  <div class="mb-3 max-w-xs">
    <SegmentControl options={LEAGUE_OPTIONS} bind:value={leagueFilter} />
  </div>

  {#if loading}
    <div class="space-y-2">
      {#each Array.from({ length: 6 }) as _, index (index)}
        <div class="h-9 rounded-md border border-outline bg-surface/30"></div>
      {/each}
    </div>
  {:else if errorMessage}
    <div class="rounded-md border border-warning/35 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
      {errorMessage}
    </div>
  {:else if divisionGroups.length === 0}
    <div class="rounded-md border border-outline bg-surface/30 px-3 py-3 text-[0.78rem] text-muted">
      No standings available for the selected league.
    </div>
  {:else}
    <div class="space-y-3">
      {#each divisionGroups as group (group.division)}
        <details class="overflow-hidden rounded-md border border-outline/80 bg-surface/25" open>
          <summary
            class="cursor-pointer list-none px-3 py-2 font-mono text-[0.68rem] tracking-[0.08em] text-muted uppercase">
            <div class="flex items-center justify-between gap-2">
              <span>{group.division}</span>
              <span>{group.rows.length} teams</span>
            </div>
          </summary>
          <div class="overflow-x-auto border-t border-outline/60">
            <table class="min-w-full border-collapse">
              <thead>
                <tr class="border-b border-outline/60 bg-mantle/50">
                  <th class="px-2 py-1.5 text-left font-mono text-[0.62rem] text-muted uppercase">
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('rank')}>
                      Rank {arrowFor('rank')}
                    </button>
                  </th>
                  <th class="px-2 py-1.5 text-left font-mono text-[0.62rem] text-muted uppercase">
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('team')}>
                      Team {arrowFor('team')}
                    </button>
                  </th>
                  <th class="px-2 py-1.5 text-right font-mono text-[0.62rem] text-muted uppercase">
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('wins')}>
                      W {arrowFor('wins')}
                    </button>
                  </th>
                  <th class="px-2 py-1.5 text-right font-mono text-[0.62rem] text-muted uppercase">
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('losses')}>
                      L {arrowFor('losses')}
                    </button>
                  </th>
                  <th class="px-2 py-1.5 text-right font-mono text-[0.62rem] text-muted uppercase">
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('pct')}>
                      PCT {arrowFor('pct')}
                    </button>
                  </th>
                  <th class="px-2 py-1.5 text-right font-mono text-[0.62rem] text-muted uppercase">
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('gb')}>
                      GB {arrowFor('gb')}
                    </button>
                  </th>
                  <th class="px-2 py-1.5 text-right font-mono text-[0.62rem] text-muted uppercase">
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('wcGb')}>
                      WC GB {arrowFor('wcGb')}
                    </button>
                  </th>
                  <th class="px-2 py-1.5 text-right font-mono text-[0.62rem] text-muted uppercase">
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('streak')}>
                      Streak {arrowFor('streak')}
                    </button>
                  </th>
                  <th class="px-2 py-1.5 text-right font-mono text-[0.62rem] text-muted uppercase">
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('runDiff')}>
                      Run Diff {arrowFor('runDiff')}
                    </button>
                  </th>
                  <th class="px-2 py-1.5 text-right font-mono text-[0.62rem] text-muted uppercase">
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('last10')}>
                      L10 {arrowFor('last10')}
                    </button>
                  </th>
                </tr>
              </thead>
              <tbody>
                {#each group.rows as row (`${group.division}:${row.teamName}`)}
                  <tr class="border-b border-outline/50 last:border-b-0 {leaderRowClass(row.divisionLeader)}">
                    <td class="px-2 py-1.5 text-right font-mono text-[0.72rem] text-foreground">{row.rank}</td>
                    <td class="px-2 py-1.5 font-mono text-[0.72rem]">
                      <div class="flex items-center gap-2">
                        <span class="h-2 w-2 rounded-full" style={`background-color:${teamColor(row)}`}></span>
                        <a href={resolve(teamHref(row) as '/teams')} class="text-primary no-underline hover:underline"
                          >{row.teamName}</a>
                        <span class="text-[0.64rem] text-muted">{row.teamAbbr}</span>
                      </div>
                    </td>
                    <td class="px-2 py-1.5 text-right font-mono text-[0.72rem] text-foreground">{row.wins}</td>
                    <td class="px-2 py-1.5 text-right font-mono text-[0.72rem] text-foreground">{row.losses}</td>
                    <td class="px-2 py-1.5 text-right font-mono text-[0.72rem] text-foreground">{row.pct}</td>
                    <td class="px-2 py-1.5 text-right font-mono text-[0.72rem] text-foreground">{row.gamesBack}</td>
                    <td class="px-2 py-1.5 text-right font-mono text-[0.72rem] text-foreground"
                      >{row.wildCardGamesBack}</td>
                    <td class="px-2 py-1.5 text-right font-mono text-[0.72rem] {streakToneClass(row.streak)}"
                      >{row.streak}</td>
                    <td class="px-2 py-1.5 text-right font-mono text-[0.72rem] text-foreground"
                      >{runDiffLabel(row.runDiff)}</td>
                    <td class="px-2 py-1.5 text-right font-mono text-[0.72rem] text-foreground">{row.last10}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </details>
      {/each}
    </div>
  {/if}
</section>
