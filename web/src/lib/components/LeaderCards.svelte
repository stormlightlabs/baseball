<script lang="ts">
  import { resolve } from '$app/paths';
  import { apiFetch } from '$lib/api';
  import { EP } from '$lib/endpoints';
  import {
    buildLeaderBoardByCategory,
    LEADER_CATEGORIES,
    normalizeMlbTeamsAbbrByIDFromDetails,
    type LeaderCategory,
    type LeaderRow
  } from '$lib/home/leaders';
  import { onMount } from 'svelte';

  type CategoryID = LeaderCategory['id'];

  const CURRENT_SEASON = new Date().getFullYear();
  const STATS_HINT = `/v1${EP.mlbStats}?stats=season&group={hitting|pitching}&season=${CURRENT_SEASON}&playerPool=qualified&include=details`;

  let loading = $state(true);
  let refreshing = $state(false);
  let errorMessage = $state<string | null>(null);
  let activeCategory = $state<CategoryID>('HR');

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

  const category = $derived.by(
    () => LEADER_CATEGORIES.find((entry) => entry.id === activeCategory) ?? LEADER_CATEGORIES[0]
  );
  const rows = $derived(board[activeCategory] ?? []);

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
      errorMessage = null;
    } catch (error) {
      const message = error instanceof Error && error.message ? error.message : 'Failed to load today’s leaders.';
      errorMessage = message;
    } finally {
      loading = false;
      refreshing = false;
    }
  }

  function categoryButtonClass(id: CategoryID): string {
    if (id === activeCategory) return 'bg-primary/20 text-primary';
    return 'text-muted hover:text-foreground';
  }

  type PlayerHref = `/players/${string}/batting` | `/players/${string}/pitching` | `/players?${string}`;

  function playerHref(row: LeaderRow): PlayerHref {
    const localID = row.playerMlbID != null ? crosswalkByMLBID[row.playerMlbID] : undefined;
    if (localID) {
      if (row.group === 'pitching') {
        return `/players/${encodeURIComponent(localID)}/pitching`;
      }
      return `/players/${encodeURIComponent(localID)}/batting`;
    }
    return `/players?q=${encodeURIComponent(row.playerName)}`;
  }

  function toObject(value: unknown): Record<string, unknown> {
    if (value != null && typeof value === 'object' && !Array.isArray(value)) return value as Record<string, unknown>;
    return {};
  }

  function toString(value: unknown): string | undefined {
    if (typeof value === 'string') {
      const trimmed = value.trim();
      return trimmed.length > 0 ? trimmed : undefined;
    }
    if (typeof value === 'number') return String(value);
    return undefined;
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

<section class="rounded-lg border border-outline bg-crust p-4">
  <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
    <div>
      <h2 class="font-mono text-xs tracking-[0.08em] text-muted uppercase">Today’s leaders</h2>
      <p class="text-xs text-muted">Season {CURRENT_SEASON}</p>
      <p class="font-mono text-[0.63rem] text-muted">{STATS_HINT}</p>
    </div>
    <button
      type="button"
      class="rounded border border-outline px-2.5 py-1 font-mono text-[0.64rem] text-foreground transition-colors hover:bg-surface"
      onclick={() => void refreshLeaders('manual')}
      disabled={loading || refreshing}>
      {refreshing ? 'Refreshing…' : 'Refresh'}
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
    <p class="mt-2 text-xs text-muted">
      Showing top 5 by {category?.label}. Player links resolve from MLBAM IDs via `/v1/meta/crosswalk/players`.
    </p>
  {/if}
</section>
