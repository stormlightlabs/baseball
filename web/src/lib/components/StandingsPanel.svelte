<script lang="ts">
  import { resolve } from '$app/paths';
  import { apiFetch } from '$lib/api';
  import type { LeagueFilter } from '$lib/common/types';
  import SegmentControl from '$lib/components/SegmentControl.svelte';
  import { EP } from '$lib/endpoints';
  import {
    buildStandingsRowsFromSeasonStandingsPayload,
    sortStandingsRows,
    type SeasonStandingsPayload,
    type StandingsRow,
    type StandingsSortKey
  } from '$lib/mlb/standings';
  import { teamPrimaryHexFor } from '$lib/mlb/team-branding';
  import { SvelteURLSearchParams } from 'svelte/reactivity';

  type TeamHref = `/teams/${string}/overview?${string}` | `/teams?${string}`;

  let {
    season: seasonProp = new Date().getFullYear(),
    title = 'Current standings',
    showEndpointHint = true
  }: { season?: number; title?: string; showEndpointHint?: boolean } = $props();

  const LEAGUE_OPTIONS = [
    { id: 'AL', label: 'AL' },
    { id: 'NL', label: 'NL' },
    { id: 'both', label: 'Both' }
  ];

  let loading = $state(true);
  let refreshing = $state(false);
  let errorMessage = $state<string | null>(null);
  let season = $state(new Date().getFullYear());
  let lastUpdated = $state<string | null>(null);
  let leagueFilter = $state<LeagueFilter>('both');
  let sort = $state<{ key: StandingsSortKey; dir: 'asc' | 'desc' }>({ key: 'rank', dir: 'asc' });
  let rows = $state<StandingsRow[]>([]);

  let lastRequestKey = '';

  const selectedSeason = $derived.by(() => {
    const parsed = Number(seasonProp);
    if (Number.isFinite(parsed) && parsed > 0) return Math.round(parsed);
    return new Date().getFullYear();
  });

  const standingsHint = $derived(`/v1${EP.standings}?season=${selectedSeason}`);

  const filteredRows = $derived.by(() => {
    if (leagueFilter === 'both') return rows;
    return rows.filter((row) => row.league === leagueFilter);
  });

  const lastUpdatedLabel = $derived.by(() => {
    if (!lastUpdated) return null;
    const parsed = new Date(lastUpdated);
    if (Number.isNaN(parsed.getTime())) return null;
    return parsed.toLocaleString(undefined, { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
  });

  type DivisionGroup = { division: string; rows: StandingsRow[] };

  const divisionGroups = $derived.by((): DivisionGroup[] => {
    const byDivision: Record<string, StandingsRow[]> = {};
    const orderedDivisions: string[] = [];

    for (const row of filteredRows) {
      const key = `${row.league}:${row.division}`;
      const existing = byDivision[key];
      if (existing) {
        existing.push(row);
      } else {
        byDivision[key] = [row];
        orderedDivisions.push(key);
      }
    }

    return orderedDivisions.map((key) => {
      const groupRows = byDivision[key] ?? [];
      const first = groupRows[0];
      const divisionName = first ? first.division : key;
      return { division: divisionName, rows: sortStandingsRows(groupRows, sort.key, sort.dir) };
    });
  });

  $effect(() => {
    void refreshStandings('initial');
  });

  async function refreshStandings(mode: 'initial' | 'manual'): Promise<void> {
    const requestSeason = selectedSeason;
    const requestKey = String(requestSeason);
    if (mode === 'initial' && requestKey === lastRequestKey) return;
    lastRequestKey = requestKey;

    if (mode === 'initial') {
      loading = true;
    } else {
      refreshing = true;
    }

    try {
      const payload = await apiFetch<unknown>(EP.standings, { season: requestSeason });
      const normalized: SeasonStandingsPayload = buildStandingsRowsFromSeasonStandingsPayload(payload);
      rows = normalized.rows;
      season = normalized.season || requestSeason;
      lastUpdated = normalized.lastUpdated ?? null;
      errorMessage = null;
    } catch (error) {
      const message = error instanceof Error && error.message ? error.message : 'Failed to load season standings.';
      errorMessage = message;
      rows = [];
      season = requestSeason;
      lastUpdated = null;
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

    let defaultDir: 'asc' | 'desc' = 'desc';
    if (key === 'rank' || key === 'gb' || key === 'wcGb' || key === 'losses') {
      defaultDir = 'asc';
    }
    sort = { key, dir: defaultDir };
  }

  function arrowFor(key: StandingsSortKey): string {
    if (sort.key !== key) return '↕';
    if (sort.dir === 'asc') return '↑';
    return '↓';
  }

  function teamHref(row: StandingsRow): TeamHref {
    const query = new SvelteURLSearchParams({ year: String(season) });
    if (row.localFranchiseID) {
      query.set('franchise_id', row.localFranchiseID);
      const encoded = encodeURIComponent(row.localFranchiseID);
      return `/teams/${encoded}/overview?${query.toString()}`;
    }

    const lookup = row.teamAbbr && row.teamAbbr !== '—' ? row.teamAbbr : row.teamName;
    const params = new URLSearchParams({ q: lookup, year: String(season) });
    return `/teams?${params.toString()}`;
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
      <h2 class="font-mono text-[0.74rem] tracking-[0.08em] text-muted uppercase">{title}</h2>
      <p class="text-xs text-muted">Season {season}</p>
      {#if showEndpointHint}
        <p class="font-mono text-[0.63rem] text-muted">{standingsHint}</p>
      {/if}
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
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('last10')}>
                      L10 {arrowFor('last10')}
                    </button>
                  </th>
                  <th class="px-2 py-1.5 text-right font-mono text-[0.62rem] text-muted uppercase">
                    <button type="button" class="hover:text-foreground" onclick={() => toggleSort('runDiff')}>
                      Run Diff {arrowFor('runDiff')}
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
                        <a
                          href={resolve(teamHref(row) as `/teams/${string}/overview`)}
                          class="text-primary no-underline hover:underline">{row.teamName}</a>
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
                    <td class="px-2 py-1.5 text-right font-mono text-[0.72rem] text-foreground">{row.last10}</td>
                    <td class="px-2 py-1.5 text-right font-mono text-[0.72rem] text-foreground"
                      >{runDiffLabel(row.runDiff)}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </details>
      {/each}
    </div>
  {/if}

  {#if lastUpdatedLabel}
    <p class="mt-3 font-mono text-[0.63rem] text-muted">Updated {lastUpdatedLabel}</p>
  {/if}
</section>
