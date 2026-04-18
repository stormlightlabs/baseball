<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, fetchPaginated } from '$lib/api';
  import CoverageBar from '$lib/components/CoverageBar.svelte';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import EraDisclaimer from '$lib/components/EraDisclaimer.svelte';
  import { EP } from '$lib/endpoints';
  import {
    detectSeason,
    normalizeGameBoxscore,
    normalizeGameRecord,
    normalizeGameSummary,
    parseDateOnly,
    summaryPitcherLine,
    toBoxscoreStatRows,
    toCompactLineup
  } from '$lib/games/normalizers';
  import type { GameBoxscore, GameRecord, GameSummary } from '$lib/games/types';
  import { eraForYear } from '$lib/eras';
  import { AsyncValueResource } from '$lib/players/resources.svelte';
  import { onMount } from 'svelte';

  type DensityIndicator = {
    level: 'dense' | 'partial' | 'sparse';
    percent: number;
    variant: 'primary' | 'secondary' | 'warning';
    detail: string;
  };

  let gameId = $derived(page.params.id ?? '');

  const gameResource = new AsyncValueResource<GameRecord>();
  const summaryResource = new AsyncValueResource<GameSummary>();
  const boxscoreResource = new AsyncValueResource<GameBoxscore>();

  let eventTotal = $state(0);
  let pitchTotal = $state(0);

  let lastGameKey = '';
  let lastSummaryKey = '';
  let lastBoxscoreKey = '';
  let lastDensityKey = '';

  let selectedGame = $derived(gameResource.value);

  let selectedGameSeason = $derived.by(() => detectSeason(selectedGame));
  let selectedEra = $derived.by(() => {
    if (selectedGameSeason == null) return;
    return eraForYear(selectedGameSeason);
  });

  let boxscoreRows = $derived(
    toBoxscoreStatRows(boxscoreResource.value?.home_stats, boxscoreResource.value?.away_stats)
  );
  let homeLineup = $derived(toCompactLineup(boxscoreResource.value?.home_lineup ?? []));
  let awayLineup = $derived(toCompactLineup(boxscoreResource.value?.away_lineup ?? []));
  let summaryPitchers = $derived(summaryPitcherLine(selectedGame));

  let eventDensity = $derived.by((): DensityIndicator | null => {
    const game = selectedGame;
    if (!game) return null;

    const innings = Math.max(1, game.innings ?? 9);
    const eventCount = eventTotal;

    if (eventCount === 0) {
      return { level: 'sparse', percent: 8, variant: 'warning', detail: 'No play events available for this game.' };
    }

    const eventsPerInning = eventCount / innings;
    const pitchesPerEvent = pitchTotal > 0 ? pitchTotal / eventCount : 0;

    let score = Math.round(Math.min(100, eventsPerInning * 9 + pitchesPerEvent * 16));

    if (selectedEra?.code === 'fed' || selectedEra?.code === 'nlg') {
      score = Math.min(score, 62);
    }

    if (score >= 65) {
      return {
        level: 'dense',
        percent: score,
        variant: 'primary',
        detail: `${eventCount.toLocaleString()} events across ${innings} innings (~${eventsPerInning.toFixed(1)}/inning)`
      };
    }

    if (score >= 35) {
      return {
        level: 'partial',
        percent: score,
        variant: 'secondary',
        detail: `${eventCount.toLocaleString()} events and ${pitchTotal.toLocaleString()} pitches captured.`
      };
    }

    return {
      level: 'sparse',
      percent: score,
      variant: 'warning',
      detail: `Sparse event capture (${eventCount.toLocaleString()} events, ${pitchTotal.toLocaleString()} pitches).`
    };
  });

  onMount(() => {
    void refreshAll(true);
  });

  afterNavigate(() => {
    void refreshAll();
  });

  async function refreshAll(force = false): Promise<void> {
    await Promise.all([refreshGame(force), refreshSummary(force), refreshBoxscore(force), refreshDensity(force)]);
  }

  async function refreshGame(force = false): Promise<void> {
    const id = gameId;
    if (!id) {
      gameResource.clear();
      lastGameKey = '';
      return;
    }

    if (!force && id === lastGameKey) return;
    lastGameKey = id;

    await gameResource.load(async () => {
      const payload = await apiFetch<Record<string, unknown>>(EP.game(id));
      return normalizeGameRecord(payload);
    });
  }

  async function refreshSummary(force = false): Promise<void> {
    const id = gameId;
    if (!id) {
      summaryResource.clear();
      lastSummaryKey = '';
      return;
    }

    if (!force && id === lastSummaryKey) return;
    lastSummaryKey = id;

    await summaryResource.load(async () => {
      const payload = await apiFetch<Record<string, unknown>>(EP.gameSummary(id));
      return normalizeGameSummary(payload);
    });
  }

  async function refreshBoxscore(force = false): Promise<void> {
    const id = gameId;
    if (!id) {
      boxscoreResource.clear();
      lastBoxscoreKey = '';
      return;
    }

    if (!force && id === lastBoxscoreKey) return;
    lastBoxscoreKey = id;

    await boxscoreResource.load(async () => {
      const payload = await apiFetch<Record<string, unknown>>(EP.gameBoxscore(id));
      return normalizeGameBoxscore(payload);
    });
  }

  async function refreshDensity(force = false): Promise<void> {
    const id = gameId;
    if (!id) {
      eventTotal = 0;
      pitchTotal = 0;
      lastDensityKey = '';
      return;
    }

    if (!force && id === lastDensityKey) return;
    lastDensityKey = id;

    const [eventsPage, pitchesPage] = await Promise.all([
      fetchPaginated<Record<string, unknown>>(EP.gameEvents(id), { page: 1, per_page: 1 }),
      fetchPaginated<Record<string, unknown>>(EP.gamePitches(id), { page: 1, per_page: 1 })
    ]);

    eventTotal = eventsPage.total;
    pitchTotal = pitchesPage.total;
  }

  function fmtNum(value: number | undefined): string {
    return value == null ? '—' : value.toLocaleString();
  }

  function fmtGameScore(game: GameRecord): string {
    if (game.away_score == null && game.home_score == null) return '—';
    return `${game.away_score ?? '?'} – ${game.home_score ?? '?'}`;
  }
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="mb-3 flex flex-wrap items-center gap-2">
    <div class="panel-label mb-0 grow border-b-0 pb-0">Game Metadata</div>
    <span class="rounded border border-outline px-2 py-0.5 font-mono text-[0.64rem] text-muted">{gameId}</span>
    {#if selectedEra}
      <EraBadge era={selectedEra} />
    {/if}
  </div>

  {#if gameResource.loading}
    <p class="font-mono text-[0.78rem] text-muted">Loading game detail…</p>
  {:else if gameResource.error}
    <p class="font-mono text-[0.78rem] text-warning">{gameResource.error}</p>
  {:else if selectedGame}
    <div class="grid grid-cols-2 gap-2 md:grid-cols-4">
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Date</div>
        <div class="font-display text-[0.86rem] text-foreground">{parseDateOnly(selectedGame.date)}</div>
      </div>
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Matchup</div>
        <div class="font-display text-[0.86rem] text-foreground">
          {selectedGame.away_team ?? '—'} @ {selectedGame.home_team ?? '—'}
        </div>
      </div>
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Score</div>
        <div class="font-display text-[0.86rem] text-foreground">{fmtGameScore(selectedGame)}</div>
      </div>
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Innings</div>
        <div class="font-display text-[0.86rem] text-foreground">{selectedGame.innings ?? '—'}</div>
      </div>
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Park</div>
        <div class="font-display text-[0.82rem] text-foreground">
          {selectedGame.park_name ?? selectedGame.park_id ?? '—'}
        </div>
      </div>
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Attendance</div>
        <div class="font-display text-[0.86rem] text-foreground">{fmtNum(selectedGame.attendance)}</div>
      </div>
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Duration</div>
        <div class="font-display text-[0.86rem] text-foreground">
          {selectedGame.duration_min == null ? '—' : `${selectedGame.duration_min} min`}
        </div>
      </div>
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Postseason</div>
        <div class="font-display text-[0.86rem] text-foreground">{selectedGame.is_postseason ? 'Yes' : 'No'}</div>
      </div>
    </div>

    {#if summaryPitchers.length > 0}
      <div class="mt-3 flex flex-wrap gap-1.5">
        {#each summaryPitchers as item (item)}
          <span class="rounded border border-outline px-2 py-0.5 font-mono text-[0.66rem] text-muted">{item}</span>
        {/each}
      </div>
    {/if}

    {#if selectedGame.home_team}
      <div class="mt-3 flex flex-wrap items-center gap-2">
        <a
          href={resolve(`/teams?q=${encodeURIComponent(selectedGame.home_team)}` as '/teams')}
          class="rounded border border-outline px-2 py-0.5 font-mono text-[0.66rem] text-primary hover:underline">
          Explore team: {selectedGame.home_team}
        </a>
        {#if selectedGame.away_team}
          <a
            href={resolve(`/teams?q=${encodeURIComponent(selectedGame.away_team)}` as '/teams')}
            class="rounded border border-outline px-2 py-0.5 font-mono text-[0.66rem] text-primary hover:underline">
            Explore team: {selectedGame.away_team}
          </a>
        {/if}
      </div>
    {/if}
  {/if}
</div>

{#if selectedEra}
  <div class="mt-3 rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-2">Event Density Confidence</div>

    {#if eventDensity}
      <CoverageBar
        label={`Coverage level: ${eventDensity.level}`}
        range={eventDensity.detail}
        percent={eventDensity.percent}
        variant={eventDensity.variant}
        tooltip="Confidence is estimated from event and pitch density for this game." />
    {/if}

    {#if selectedEra.caveat}
      <div class="mt-3">
        <EraDisclaimer eras={[selectedEra]} message={`${selectedEra.code}: ${selectedEra.caveat}`} />
      </div>
    {/if}
  </div>
{/if}

<div class="mt-4 rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-3">Summary + Boxscore</div>

  <div class="mb-4 grid gap-3 md:grid-cols-2">
    <div class="rounded border border-outline bg-surface p-3">
      <div class="mb-2 font-mono text-[0.63rem] tracking-wider text-muted uppercase">Game summary endpoint</div>

      {#if summaryResource.loading}
        <p class="font-mono text-[0.74rem] text-muted">Loading summary…</p>
      {:else if summaryResource.error}
        <p class="font-mono text-[0.74rem] text-warning">{summaryResource.error}</p>
      {:else if summaryResource.value}
        <div class="font-mono text-[0.72rem] text-muted">
          Winner: <span class="text-foreground">{summaryResource.value.winner ?? '—'}</span>
        </div>
        <div class="mt-2 font-mono text-[0.68rem] text-muted">
          Home lineup: {summaryResource.value.home_lineup.length} starters · Away lineup: {summaryResource.value
            .away_lineup.length} starters
        </div>
      {/if}
    </div>

    <div class="rounded border border-outline bg-surface p-3">
      <div class="mb-2 font-mono text-[0.63rem] tracking-wider text-muted uppercase">Boxscore endpoint</div>

      {#if boxscoreResource.loading}
        <p class="font-mono text-[0.74rem] text-muted">Loading boxscore…</p>
      {:else if boxscoreResource.error}
        <p class="font-mono text-[0.74rem] text-warning">{boxscoreResource.error}</p>
      {:else if boxscoreResource.value}
        <div class="overflow-x-auto">
          <table class="w-full border-collapse text-[0.7rem]">
            <thead>
              <tr>
                {#each ['Stat', 'Away', 'Home'] as col (col)}
                  <th
                    class="border-b border-outline px-2 py-1 text-left font-sans text-[0.67rem] font-medium text-muted"
                    >{col}</th>
                {/each}
              </tr>
            </thead>
            <tbody>
              {#each boxscoreRows as row (row.label)}
                <tr class="border-b border-outline last:border-b-0">
                  <td class="px-2 py-1 font-mono text-[0.68rem] text-foreground">{row.label}</td>
                  <td class="px-2 py-1 font-mono text-[0.68rem] text-muted">{row.away}</td>
                  <td class="px-2 py-1 font-mono text-[0.68rem] text-muted">{row.home}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>
  </div>

  {#if boxscoreResource.value && (awayLineup.length > 0 || homeLineup.length > 0)}
    <div class="grid gap-3 md:grid-cols-2">
      <div class="rounded border border-outline bg-surface p-3">
        <div class="mb-1.5 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Away lineup</div>
        {#if awayLineup.length === 0}
          <p class="font-mono text-[0.72rem] text-muted">No lineup rows.</p>
        {:else}
          <ul class="space-y-0.5">
            {#each awayLineup as row (row)}
              <li class="font-mono text-[0.68rem] text-foreground">{row}</li>
            {/each}
          </ul>
        {/if}
      </div>
      <div class="rounded border border-outline bg-surface p-3">
        <div class="mb-1.5 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Home lineup</div>
        {#if homeLineup.length === 0}
          <p class="font-mono text-[0.72rem] text-muted">No lineup rows.</p>
        {:else}
          <ul class="space-y-0.5">
            {#each homeLineup as row (row)}
              <li class="font-mono text-[0.68rem] text-foreground">{row}</li>
            {/each}
          </ul>
        {/if}
      </div>
    </div>
  {/if}
</div>
