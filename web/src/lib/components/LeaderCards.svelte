<script lang="ts">
  import { resolve } from '$app/paths';
  import { apiFetch } from '$lib/api';
  import { toObject, toString } from '$lib/common/converters';
  import LiveHomeCard from '$lib/components/LiveHomeCard.svelte';
  import { EP } from '$lib/endpoints';
  import {
    buildLeaderBoardByCategory,
    buildLocalLeaderBoardByCategory,
    LEADER_CATEGORIES,
    normalizeMlbTeamsAbbrByIDFromDetails,
    type LeaderCategory,
    type LeaderRow
  } from '$lib/home/leaders';
  import { onMount } from 'svelte';
  import { SvelteSet } from 'svelte/reactivity';

  type CategoryID = LeaderCategory['id'];
  type SourceMode = 'local' | 'mlb';

  type LocalBattingLeader = {
    player_id?: string;
    team_id?: string;
    hr?: number;
    avg?: number;
    ops?: number;
    rbi?: number;
    sb?: number;
  };

  type LocalPitchingLeader = {
    player_id?: string;
    team_id?: string;
    era?: number;
    so?: number;
    w?: number;
    sv?: number;
    whip?: number;
  };

  type SeasonLeadersPayload<T> = { leaders?: T[] };

  type PlayerHref = `/players/${string}/batting` | `/players/${string}/pitching` | `/players?${string}`;

  const CURRENT_SEASON = new Date().getFullYear();
  const MLB_STATS_HINT = `/v1${EP.mlbStats}?stats=season&group={hitting|pitching}&season=${CURRENT_SEASON}&playerPool=qualified&include=details`;
  const LOCAL_STATS_HINT = `/v1${EP.seasonLeadersBatting(CURRENT_SEASON)} + /v1${EP.seasonLeadersPitching(CURRENT_SEASON)}`;

  let loading = $state(true);
  let refreshing = $state(false);
  let errorMessage = $state<string | null>(null);
  let activeCategory = $state<CategoryID>('HR');
  let sourceMode = $state<SourceMode>('local');
  let localFallbackUsed = $state(false);

  let board = $state<Record<CategoryID, LeaderRow[]>>({
    HR: [],
    AVG: [],
    OPS: [],
    RBI: [],
    SB: [],
    ERA: [],
    SO: [],
    W: [],
    SV: [],
    WHIP: []
  });

  let crosswalkByMLBID = $state<Record<number, string>>({});

  const rows = $derived(board[activeCategory] ?? []);
  const activeStatsEndpoint = $derived(sourceMode === 'local' ? LOCAL_STATS_HINT : MLB_STATS_HINT);

  onMount(() => {
    void refreshLeaders('initial');
  });

  async function refreshLeaders(mode: 'initial' | 'manual'): Promise<void> {
    if (mode === 'initial') {
      loading = true;
    } else {
      refreshing = true;
    }

    try {
      let localLoaded = false;
      try {
        localLoaded = await refreshFromLocalLeaders();
      } catch {
        localLoaded = false;
      }

      if (!localLoaded) {
        localFallbackUsed = true;
        await refreshFromMLBProxy();
      } else {
        localFallbackUsed = false;
      }

      errorMessage = null;
    } catch (error) {
      const message = error instanceof Error && error.message ? error.message : 'Failed to load today’s leaders.';
      errorMessage = message;
    } finally {
      loading = false;
      refreshing = false;
    }
  }

  async function refreshFromLocalLeaders(): Promise<boolean> {
    const [battingPayload, pitchingPayload] = await Promise.all([
      apiFetch<SeasonLeadersPayload<LocalBattingLeader>>(EP.seasonLeadersBatting(CURRENT_SEASON), {
        stat: 'hr',
        page: 1,
        per_page: 3000
      }),
      apiFetch<SeasonLeadersPayload<LocalPitchingLeader>>(EP.seasonLeadersPitching(CURRENT_SEASON), {
        stat: 'era',
        page: 1,
        per_page: 3000
      })
    ]);

    const battingRows = battingPayload.leaders ?? [];
    const pitchingRows = pitchingPayload.leaders ?? [];
    if (battingRows.length === 0 && pitchingRows.length === 0) return false;

    const localBoard = buildLocalLeaderBoardByCategory(
      battingRows as Record<string, unknown>[],
      pitchingRows as Record<string, unknown>[],
      5
    );

    const hasRows = Object.values(localBoard).some((entries) => entries.length > 0);
    if (!hasRows) return false;

    const localIDs = new SvelteSet<string>();
    for (const entries of Object.values(localBoard)) {
      for (const row of entries) {
        if (row.localPlayerID) {
          localIDs.add(row.localPlayerID);
        }
      }
    }

    const localNames = await resolveLocalPlayerNames([...localIDs]);
    board = mapLeaderNames(localBoard, localNames);
    crosswalkByMLBID = {};
    sourceMode = 'local';
    return true;
  }

  async function refreshFromMLBProxy(): Promise<void> {
    const [hittingPayload, pitchingPayload] = await Promise.all([
      apiFetch<unknown>(EP.mlbStats, {
        stats: 'season',
        group: 'hitting',
        season: CURRENT_SEASON,
        playerPool: 'qualified',
        include: 'details'
      }),
      apiFetch<unknown>(EP.mlbStats, {
        stats: 'season',
        group: 'pitching',
        season: CURRENT_SEASON,
        playerPool: 'qualified',
        include: 'details'
      })
    ]);

    const teamAbbrByID = {
      ...normalizeMlbTeamsAbbrByIDFromDetails(hittingPayload),
      ...normalizeMlbTeamsAbbrByIDFromDetails(pitchingPayload)
    };

    board = buildLeaderBoardByCategory(hittingPayload, pitchingPayload, teamAbbrByID, 5);
    crosswalkByMLBID = {
      ...extractPlayerCrosswalkByMLBID(hittingPayload),
      ...extractPlayerCrosswalkByMLBID(pitchingPayload)
    };
    sourceMode = 'mlb';
  }

  function mapLeaderNames(
    source: Record<CategoryID, LeaderRow[]>,
    namesByLocalID: Record<string, string>
  ): Record<CategoryID, LeaderRow[]> {
    return {
      HR: mapCategoryRows(source.HR, namesByLocalID),
      AVG: mapCategoryRows(source.AVG, namesByLocalID),
      OPS: mapCategoryRows(source.OPS, namesByLocalID),
      RBI: mapCategoryRows(source.RBI, namesByLocalID),
      SB: mapCategoryRows(source.SB, namesByLocalID),
      ERA: mapCategoryRows(source.ERA, namesByLocalID),
      SO: mapCategoryRows(source.SO, namesByLocalID),
      W: mapCategoryRows(source.W, namesByLocalID),
      SV: mapCategoryRows(source.SV, namesByLocalID),
      WHIP: mapCategoryRows(source.WHIP, namesByLocalID)
    };
  }

  function mapCategoryRows(rowsToMap: LeaderRow[], namesByLocalID: Record<string, string>): LeaderRow[] {
    return rowsToMap.map((row) => {
      if (!row.localPlayerID) return row;
      const mapped = namesByLocalID[row.localPlayerID];
      if (!mapped) return row;
      return { ...row, playerName: mapped };
    });
  }

  async function resolveLocalPlayerNames(localIDs: string[]): Promise<Record<string, string>> {
    const namesByID: Record<string, string> = {};
    if (localIDs.length === 0) return namesByID;

    const lookups = await Promise.allSettled(
      localIDs.map(async (playerID) => {
        const payload = await apiFetch<unknown>(EP.player(playerID));
        const root = toObject(payload);
        const explicit = toString(root.name);
        if (explicit && explicit.trim().length > 0) {
          return { playerID, displayName: explicit.trim() };
        }

        const first = toString(root.first_name) ?? '';
        const last = toString(root.last_name) ?? '';
        const combined = `${first} ${last}`.trim();
        if (combined.length > 0) {
          return { playerID, displayName: combined };
        }

        return { playerID, displayName: playerID };
      })
    );

    for (const entry of lookups) {
      if (entry.status === 'fulfilled') {
        namesByID[entry.value.playerID] = entry.value.displayName;
      }
    }

    return namesByID;
  }

  function categoryButtonClass(id: CategoryID): string {
    if (id === activeCategory) return 'bg-primary/20 text-primary';
    return 'text-muted hover:text-foreground';
  }

  function playerHref(row: LeaderRow): PlayerHref {
    const localID = row.localPlayerID?.trim();
    if (localID) {
      const encoded = encodeURIComponent(localID);
      if (row.group === 'pitching') {
        return `/players/${encoded}/pitching`;
      }
      return `/players/${encoded}/batting`;
    }

    const crosswalkID = row.playerMlbID != null ? crosswalkByMLBID[row.playerMlbID] : undefined;
    if (crosswalkID) {
      if (row.group === 'pitching') {
        return `/players/${encodeURIComponent(crosswalkID)}/pitching`;
      }
      return `/players/${encodeURIComponent(crosswalkID)}/batting`;
    }

    return `/players?q=${encodeURIComponent(row.playerName)}`;
  }

  function extractPlayerCrosswalkByMLBID(payload: unknown): Record<number, string> {
    const root = toObject(payload);
    const meta = toObject(root.meta);
    const details = toObject(meta.details);
    const crosswalk = toObject(details.crosswalk);
    const mlbamPlayerToLocal = toObject(crosswalk.mlbam_player_to_local);

    const map: Record<number, string> = {};
    for (const [rawMLBID, rawLocalID] of Object.entries(mlbamPlayerToLocal)) {
      const mlbID = Number.parseInt(rawMLBID, 10);
      const localID = toString(rawLocalID);
      if (!Number.isFinite(mlbID) || !localID) continue;
      map[mlbID] = localID;
    }
    return map;
  }
</script>

<LiveHomeCard endpoint={activeStatsEndpoint}>
  <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
    <div>
      <h2 class="font-mono text-xs tracking-[0.08em] text-muted uppercase">Today’s leaders</h2>
      <p class="text-xs text-muted">Season {CURRENT_SEASON}</p>
    </div>
    <button
      type="button"
      class="rounded border border-outline px-2.5 py-1 font-sans text-xxs text-foreground transition-colors hover:bg-surface"
      onclick={() => void refreshLeaders('manual')}
      disabled={loading || refreshing}>
      {#if refreshing}
        <span class="inline-flex items-center gap-1">
          <i class="i-tabler-loader-2 animate-spin"></i> Refreshing…
        </span>
      {:else}
        <span class="inline-flex items-center gap-1">
          <i class="i-tabler-refresh"></i>
          Refresh
        </span>
      {/if}
    </button>
  </div>

  <div class="mb-3 flex flex-wrap gap-1">
    {#each LEADER_CATEGORIES as item (item.id)}
      <button
        type="button"
        class="rounded px-2 py-0.5 font-mono text-[0.63rem] transition-colors {categoryButtonClass(item.id)}"
        onclick={() => (activeCategory = item.id)}>
        {item.label}
      </button>
    {/each}
  </div>

  {#if loading}
    <div class="space-y-2">
      {#each Array.from({ length: 5 }) as _, index (index)}
        <div class="h-9 rounded-md border border-outline bg-surface/30"></div>
      {/each}
    </div>
  {:else if errorMessage}
    <div class="rounded-md border border-warning/35 bg-warning/10 px-3 py-2 font-mono text-xs text-warning">
      {errorMessage}
    </div>
  {:else}
    <ul class="space-y-1">
      {#each rows as row (`${activeCategory}:${row.rank}:${row.playerName}`)}
        <li class="rounded-md border border-outline bg-surface/25 px-3 py-2">
          <div class="flex items-center gap-3">
            <span class="w-4 text-center font-mono text-xs text-muted">{row.rank}</span>
            <a
              href={resolve(playerHref(row))}
              class="min-w-0 flex-1 truncate text-sm text-primary no-underline hover:underline">
              {row.playerName}
            </a>
            <span class="font-mono text-xs text-muted">{row.teamAbbr}</span>
            <span class="font-mono text-sm text-foreground">{row.displayValue}</span>
          </div>
        </li>
      {/each}
    </ul>

    {#if sourceMode === 'local'}
      <p class="mt-2 text-xs text-muted">Updated every 4h from local current-season leaders.</p>
    {:else if localFallbackUsed}
      <p class="mt-2 text-xs text-muted">Local leaders were empty. Showing live MLB proxy leaders.</p>
    {:else}
      <p class="mt-2 text-xs text-muted">Showing live leaders from MLB proxy.</p>
    {/if}
  {/if}
</LiveHomeCard>
