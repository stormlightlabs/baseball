<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { apiFetch } from '$lib/api';
  import LiveHomeCard from '$lib/components/LiveHomeCard.svelte';
  import { EP } from '$lib/endpoints';
  import type { ScoreboardGame, ScoreboardSnapshot } from '$lib/home/scoreboard';
  import { normalizeScoreboardResponse, todayLocalISODate } from '$lib/home/scoreboard';
  import { onMount } from 'svelte';

  const SCOREBOARD_ENDPOINT = EP.mlbSchedule;
  const SCOREBOARD_HYDRATE = 'linescore,team';
  const REFRESH_INTERVAL_MS = 30_000;

  const EMPTY_SNAPSHOT: ScoreboardSnapshot = { date: todayLocalISODate(), gamesInProgress: 0, games: [] };

  let snapshot = $state<ScoreboardSnapshot>(EMPTY_SNAPSHOT);
  let loading = $state(true);
  let refreshing = $state(false);
  let errorMessage = $state<string | null>(null);
  let autoRefreshPaused = $state(false);
  let documentVisible = $state(true);
  let windowFocused = $state(true);
  let openPopoverGameID = $state<string | null>(null);
  let lastUpdatedAt = $state<number | null>(null);

  let liveCount = $derived(
    snapshot.gamesInProgress > 0 ? snapshot.gamesInProgress : snapshot.games.filter((game) => game.isInProgress).length
  );
  let hasInProgressGames = $derived(liveCount > 0);
  let shouldAutoRefresh = $derived(
    hasInProgressGames && !autoRefreshPaused && documentVisible && windowFocused && !loading
  );

  $effect(() => {
    if (!shouldAutoRefresh) return;
    const intervalID = setInterval(() => {
      void refreshScoreboard('auto');
    }, REFRESH_INTERVAL_MS);
    return () => clearInterval(intervalID);
  });

  onMount(() => {
    documentVisible = globalThis.document.visibilityState === 'visible';
    windowFocused = globalThis.document.hasFocus();

    const handleVisibility = () => {
      documentVisible = globalThis.document.visibilityState === 'visible';
    };
    const handleWindowFocus = () => {
      windowFocused = true;
    };
    const handleWindowBlur = () => {
      windowFocused = false;
    };

    document.addEventListener('visibilitychange', handleVisibility);
    window.addEventListener('focus', handleWindowFocus);
    window.addEventListener('blur', handleWindowBlur);

    void refreshScoreboard('initial');

    return () => {
      document.removeEventListener('visibilitychange', handleVisibility);
      window.removeEventListener('focus', handleWindowFocus);
      window.removeEventListener('blur', handleWindowBlur);
    };
  });

  async function refreshScoreboard(source: 'initial' | 'manual' | 'auto'): Promise<void> {
    if (source === 'auto' && refreshing) return;

    if (source === 'initial') {
      loading = true;
    } else {
      refreshing = true;
    }

    try {
      const date = todayLocalISODate();
      const payload = await apiFetch<unknown>(SCOREBOARD_ENDPOINT, { date, hydrate: SCOREBOARD_HYDRATE });
      snapshot = normalizeScoreboardResponse(payload);
      errorMessage = null;
      lastUpdatedAt = Date.now();
    } catch (error) {
      const message = error instanceof Error && error.message ? error.message : 'Unable to load live scoreboard.';
      errorMessage = message;
    } finally {
      loading = false;
      refreshing = false;
    }
  }

  async function openGameCard(game: ScoreboardGame): Promise<void> {
    if (game.retrosheetGameId) {
      openPopoverGameID = null;
      const href = `/games/${game.retrosheetGameId}/overview`;
      await goto(resolve(href as `/games/${string}/overview`));
      return;
    }

    if (openPopoverGameID === game.id) {
      openPopoverGameID = null;
      return;
    }
    openPopoverGameID = game.id;
  }

  function displayDate(date: string | undefined): string {
    if (!date) return 'Today';
    const parsed = parseDateOnly(date);
    if (!parsed) return date;
    return parsed.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' });
  }

  function displayClock(value: string | undefined): string {
    if (!value) return 'Scheduled';
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return value;
    return parsed.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' });
  }

  function parseDateOnly(raw: string): Date | null {
    const matched = raw.match(/^(\d{4})-(\d{2})-(\d{2})$/);
    if (!matched) return null;
    const year = Number.parseInt(matched[1] ?? '', 10);
    const month = Number.parseInt(matched[2] ?? '', 10);
    const day = Number.parseInt(matched[3] ?? '', 10);
    if (!Number.isFinite(year) || !Number.isFinite(month) || !Number.isFinite(day)) return null;
    return new Date(year, month - 1, day);
  }

  function finderHrefFor(game: ScoreboardGame): `/games?${string}` {
    const search = [game.away.abbreviation, game.home.abbreviation, game.date].filter(Boolean).join(' ');
    return `/games?mode=nl&q=${encodeURIComponent(search)}`;
  }

  function teamColor(game: ScoreboardGame, side: 'away' | 'home'): string {
    const color = side === 'away' ? game.away.color : game.home.color;
    return color ?? '#374151';
  }

  function gameCardBorderClass(game: ScoreboardGame): string {
    if (game.isInProgress) return 'border-rose-400/35';
    return 'border-outline';
  }

  function gameStatusToneClass(game: ScoreboardGame): string {
    if (game.isInProgress) return 'text-rose-300';
    if (game.isFinal) return 'text-secondary';
    return 'text-muted';
  }

  function gameStatusLabel(game: ScoreboardGame): string {
    if (game.isInProgress) return game.statusText;
    if (game.isFinal) return game.statusText;
    return displayClock(game.scheduledLabel);
  }

  let lastUpdatedLabel = $derived.by(() => {
    if (lastUpdatedAt == null) return null;
    return new Date(lastUpdatedAt).toLocaleTimeString(undefined, {
      hour: 'numeric',
      minute: '2-digit',
      second: '2-digit'
    });
  });

  let activeScoreboardEndpoint = $derived(
    `/v1${SCOREBOARD_ENDPOINT}?date=${todayLocalISODate()}&hydrate=${SCOREBOARD_HYDRATE}`
  );
</script>

<LiveHomeCard endpoint={activeScoreboardEndpoint}>
  <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
    <div class="min-w-0">
      <div class="mb-0.5 flex items-center gap-2">
        <h2 class="font-mono text-[0.74rem] tracking-[0.08em] text-muted uppercase">Live scoreboard</h2>
        {#if liveCount > 0}
          <span
            class="inline-flex items-center gap-1 rounded-full border border-rose-400/30 bg-rose-500/10 px-2 py-0.5 font-mono text-[0.62rem] text-rose-400">
            <span class="live-dot"></span>LIVE {liveCount}
          </span>
        {/if}
      </div>
      <p class="text-xs text-muted">Today: {displayDate(snapshot.date)}</p>
    </div>

    <div class="flex flex-wrap items-center justify-end gap-2">
      <button
        type="button"
        class="rounded border border-outline px-2.5 py-1 font-sans text-xxs text-foreground transition-colors hover:bg-surface"
        onclick={() => void refreshScoreboard('manual')}
        disabled={refreshing || loading}>
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
      <button
        type="button"
        class="rounded border border-outline px-2.5 py-1 font-sans text-xxs transition-colors hover:bg-surface {autoRefreshPaused
          ? 'text-warning'
          : 'text-secondary'}"
        onclick={() => (autoRefreshPaused = !autoRefreshPaused)}>
        {#if autoRefreshPaused}
          <span class="ml-1 inline-flex items-center gap-1 text-warning">
            <i class="i-tabler-play"></i>
            <span> Resume Auto-Refresh </span>
          </span>
        {:else}
          <span class="ml-1 inline-flex items-center gap-1 text-secondary">
            <i class="i-tabler-pause"></i>
            <span> Pause Auto-Refresh </span>
          </span>
        {/if}
      </button>
    </div>
  </div>

  {#if lastUpdatedLabel}
    <p class="mb-2 font-mono text-[0.62rem] text-muted">Last updated: {lastUpdatedLabel}</p>
  {/if}

  {#if loading}
    <div class="flex gap-3 overflow-x-auto pb-1">
      {#each Array.from({ length: 4 }) as _, index (index)}
        <div class="h-28 w-44 shrink-0 rounded-md border border-outline bg-surface/40"></div>
      {/each}
    </div>
  {:else if errorMessage}
    <div class="rounded-md border border-warning/35 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
      <p>{errorMessage}</p>
      <button
        type="button"
        class="mt-1 underline decoration-dotted underline-offset-2"
        onclick={() => void refreshScoreboard('manual')}>
        Retry
      </button>
    </div>
  {:else if snapshot.games.length === 0}
    <div class="rounded-md border border-outline bg-surface/30 px-3 py-3">
      <p class="text-sm text-foreground">No games today.</p>
      {#if snapshot.nextGameDate}
        <p class="mt-1 text-[0.78rem] text-muted">Next game date: {displayDate(snapshot.nextGameDate)}</p>
      {/if}
    </div>
  {:else}
    <div class="overflow-x-auto pb-1">
      <div class="flex min-w-max gap-3">
        {#each snapshot.games as game (game.id)}
          <div class="w-48 shrink-0">
            <button
              type="button"
              class="w-full rounded-md border bg-surface/35 p-2 text-left transition-colors hover:bg-surface/55 focus-visible:ring-2 focus-visible:ring-primary/60 focus-visible:outline-hidden {gameCardBorderClass(
                game
              )}"
              onclick={() => void openGameCard(game)}>
              <div class="mb-2 flex items-center justify-between gap-2">
                <div class="truncate font-mono text-xxs uppercase {gameStatusToneClass(game)}">
                  {gameStatusLabel(game)}
                </div>
                {#if game.isInProgress}
                  <span
                    class="inline-flex items-center gap-1 rounded-full border border-rose-400/30 bg-rose-500/10 px-1.5 py-0.5 font-mono text-[0.57rem] text-rose-400">
                    <span class="live-dot"></span>LIVE
                  </span>
                {/if}
              </div>

              <div class="space-y-1.5">
                <div class="flex items-center justify-between gap-2">
                  <div class="flex min-w-0 items-center gap-2">
                    <span class="h-4 w-1 rounded-xs" style={`background-color:${teamColor(game, 'away')}`}></span>
                    <span class="font-mono text-[0.76rem] text-foreground">{game.away.abbreviation}</span>
                  </div>
                  <span class="font-mono text-sm text-foreground">{game.away.score ?? '-'}</span>
                </div>
                <div class="flex items-center justify-between gap-2">
                  <div class="flex min-w-0 items-center gap-2">
                    <span class="h-4 w-1 rounded-xs" style={`background-color:${teamColor(game, 'home')}`}></span>
                    <span class="font-mono text-[0.76rem] text-foreground">{game.home.abbreviation}</span>
                  </div>
                  <span class="font-mono text-sm text-foreground">{game.home.score ?? '-'}</span>
                </div>
              </div>

              <div class="mt-2 truncate text-[0.68rem] text-muted">{game.venue ?? 'Venue TBD'}</div>
            </button>

            {#if openPopoverGameID === game.id && !game.retrosheetGameId}
              <div class="mt-2 rounded-md border border-outline bg-mantle px-2 py-2">
                <p class="font-mono text-[0.62rem] text-warning uppercase">MLB-only live detail</p>
                <p class="mt-1 text-[0.7rem] text-muted">No Retrosheet crosswalk is available for this game yet.</p>
                <div class="mt-1 space-y-0.5 font-mono text-xxs text-foreground">
                  <p>Matchup: {game.away.abbreviation} @ {game.home.abbreviation}</p>
                  <p>gamePk: {game.gamePk ?? '-'}</p>
                  <p>Status: {game.statusText}</p>
                </div>
                <a
                  class="mt-1 inline-block text-[0.68rem] text-primary no-underline hover:underline"
                  href={resolve(finderHrefFor(game))}>
                  Open in game finder
                </a>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</LiveHomeCard>

<style>
  .live-dot {
    height: 0.35rem;
    width: 0.35rem;
    border-radius: 9999px;
    background: currentColor;
    animation: scoreboard-pulse 1.25s ease-in-out infinite;
  }

  @keyframes scoreboard-pulse {
    0%,
    100% {
      opacity: 0.35;
      transform: scale(0.8);
    }
    50% {
      opacity: 1;
      transform: scale(1);
    }
  }
</style>
