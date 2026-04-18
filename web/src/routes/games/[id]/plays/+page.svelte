<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, fetchPaginated } from '$lib/api';
  import Pagination from '$lib/components/Pagination.svelte';
  import { EP } from '$lib/endpoints';
  import { normalizePitchPage, normalizePlayPage, normalizePlayPitches } from '$lib/games/normalizers';
  import type { PitchEvent, PlayEvent } from '$lib/games/types';
  import { AsyncListResource, AsyncPaginatedListResource } from '$lib/players/resources.svelte';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { intParam } from '$lib/url-state.svelte';
  import { onMount } from 'svelte';

  const DEFAULT_PLAYS_PAGE = 1;
  const DEFAULT_PLAYS_PER_PAGE = 25;
  const DEFAULT_PITCHES_PAGE = 1;
  const DEFAULT_PITCHES_PER_PAGE = 30;

  let gameId = $derived(page.params.id ?? '');

  let playsPage = $derived(intParam(page.url.searchParams, 'plays_page', DEFAULT_PLAYS_PAGE));
  let playsPerPage = $derived(intParam(page.url.searchParams, 'plays_per_page', DEFAULT_PLAYS_PER_PAGE));

  let pitchesPage = $derived(intParam(page.url.searchParams, 'pitches_page', DEFAULT_PITCHES_PAGE));
  let pitchesPerPage = $derived(intParam(page.url.searchParams, 'pitches_per_page', DEFAULT_PITCHES_PER_PAGE));

  let selectedPlayNum = $derived.by(() => {
    const raw = page.url.searchParams.get('play_num');
    if (!raw) return null;
    const parsed = Number.parseInt(raw, 10);
    return Number.isNaN(parsed) || parsed <= 0 ? null : parsed;
  });

  const playsResource = new AsyncPaginatedListResource<PlayEvent>();
  const pitchesResource = new AsyncPaginatedListResource<PitchEvent>();
  const playPitchesResource = new AsyncListResource<PitchEvent>();

  let lastPlaysKey = '';
  let lastPitchesKey = '';
  let lastPlayPitchesKey = '';

  onMount(() => {
    void refreshAll(true);
  });

  afterNavigate(() => {
    void refreshAll();
  });

  async function refreshAll(force = false): Promise<void> {
    await Promise.all([refreshPlays(force), refreshPitches(force), refreshPlayPitches(force)]);
  }

  async function refreshPlays(force = false): Promise<void> {
    const id = gameId;
    if (!id) {
      playsResource.clear();
      lastPlaysKey = '';
      return;
    }

    const key = `${id}|${playsPage}|${playsPerPage}`;
    if (!force && key === lastPlaysKey) return;
    lastPlaysKey = key;

    await playsResource.load(async () => {
      const payload = await fetchPaginated<Record<string, unknown>>(EP.gamePlays(id), {
        page: playsPage,
        per_page: playsPerPage
      });
      return normalizePlayPage(payload);
    });
  }

  async function refreshPitches(force = false): Promise<void> {
    const id = gameId;
    if (!id) {
      pitchesResource.clear();
      lastPitchesKey = '';
      return;
    }

    const key = `${id}|${pitchesPage}|${pitchesPerPage}`;
    if (!force && key === lastPitchesKey) return;
    lastPitchesKey = key;

    await pitchesResource.load(async () => {
      const payload = await fetchPaginated<Record<string, unknown>>(EP.gamePitches(id), {
        page: pitchesPage,
        per_page: pitchesPerPage
      });
      return normalizePitchPage(payload);
    });
  }

  async function refreshPlayPitches(force = false): Promise<void> {
    const id = gameId;
    const playNum = selectedPlayNum;
    if (!id || playNum == null) {
      playPitchesResource.clear();
      lastPlayPitchesKey = '';
      return;
    }

    const key = `${id}|${playNum}`;
    if (!force && key === lastPlayPitchesKey) return;
    lastPlayPitchesKey = key;

    await playPitchesResource.load(async () => {
      const payload = await apiFetch<unknown>(EP.gamePlayPitches(id, playNum));
      return normalizePlayPitches(payload);
    });
  }

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const href = withMergedQuery(
      `/games/${encodeURIComponent(gameId)}/plays`,
      page.url.searchParams,
      overrides,
      page.url.hash
    );
    void goto(resolve(href as `/games/${string}/plays`), QUERY_NAV_OPTS);
  }

  function selectPlay(playNum: number | undefined): void {
    if (playNum == null) return;
    updateQuery({ play_num: playNum });
  }

  function fmtTopBottom(topBot: number | undefined): string {
    if (topBot == null) return '—';
    return topBot === 0 ? 'Top' : 'Bot';
  }
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-3">Plays + Pitches Drilldown</div>

  <div class="grid gap-4 lg:grid-cols-2">
    <div>
      <div class="mb-2 font-mono text-[0.63rem] tracking-wider text-muted uppercase">Plays</div>
      {#if playsResource.loading}
        <p class="font-mono text-[0.74rem] text-muted">Loading plays…</p>
      {:else if playsResource.error}
        <p class="font-mono text-[0.74rem] text-warning">{playsResource.error}</p>
      {:else if playsResource.items.length === 0}
        <p class="font-mono text-[0.74rem] text-muted">No plays returned.</p>
      {:else}
        <div class="max-h-72 overflow-y-auto rounded border border-outline">
          <table class="w-full border-collapse text-[0.7rem]">
            <thead>
              <tr>
                {#each ['Play', 'Inn', 'Half', 'Event'] as col (col)}
                  <th
                    class="border-b border-outline bg-surface px-2 py-1 text-left font-sans text-[0.66rem] text-muted">
                    {col}
                  </th>
                {/each}
              </tr>
            </thead>
            <tbody>
              {#each playsResource.items as play (`${play.play_num ?? ''}-${play.event ?? ''}`)}
                <tr
                  class="cursor-pointer border-b border-outline last:border-b-0 hover:[&>td]:bg-surface {selectedPlayNum ===
                  play.play_num
                    ? '[&>td]:bg-surface'
                    : ''}"
                  onclick={() => selectPlay(play.play_num)}>
                  <td class="px-2 py-1.5 font-mono text-[0.68rem] text-primary">{play.play_num ?? '—'}</td>
                  <td class="px-2 py-1.5 font-mono text-[0.68rem] text-muted">{play.inning ?? '—'}</td>
                  <td class="px-2 py-1.5 font-mono text-[0.68rem] text-muted">{fmtTopBottom(play.top_bot)}</td>
                  <td class="px-2 py-1.5 font-mono text-[0.68rem] text-foreground">{play.event ?? '—'}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        {#if playsResource.total > playsPerPage}
          <div class="mt-2">
            <Pagination
              page={playsPage}
              perPage={playsPerPage}
              total={playsResource.total}
              onPageChange={(nextPage) => updateQuery({ plays_page: nextPage })}
              onPerPageChange={(nextPerPage) => updateQuery({ plays_page: 1, plays_per_page: nextPerPage })} />
          </div>
        {/if}
      {/if}
    </div>

    <div>
      <div class="mb-2 flex flex-wrap items-center gap-2">
        <div class="font-mono text-[0.63rem] tracking-wider text-muted uppercase">Pitches</div>
        {#if selectedPlayNum != null}
          <span class="rounded border border-outline px-2 py-0.5 font-mono text-[0.64rem] text-muted"
            >play {selectedPlayNum}</span>
        {/if}
      </div>

      {#if pitchesResource.loading}
        <p class="font-mono text-[0.74rem] text-muted">Loading pitches…</p>
      {:else if pitchesResource.error}
        <p class="font-mono text-[0.74rem] text-warning">{pitchesResource.error}</p>
      {:else if pitchesResource.items.length === 0}
        <p class="font-mono text-[0.74rem] text-muted">No pitches returned.</p>
      {:else}
        <div class="max-h-72 overflow-y-auto rounded border border-outline">
          <table class="w-full border-collapse text-[0.7rem]">
            <thead>
              <tr>
                {#each ['Play', 'Seq', 'Type', 'Count', 'Pitcher', 'Desc'] as col (col)}
                  <th
                    class="border-b border-outline bg-surface px-2 py-1 text-left font-sans text-[0.66rem] text-muted">
                    {col}
                  </th>
                {/each}
              </tr>
            </thead>
            <tbody>
              {#each pitchesResource.items as pitch (`${pitch.play_num ?? ''}-${pitch.seq_num ?? ''}`)}
                <tr
                  class="border-b border-outline last:border-b-0 hover:[&>td]:bg-surface {selectedPlayNum != null &&
                  selectedPlayNum === pitch.play_num
                    ? '[&>td]:bg-surface'
                    : ''}">
                  <td class="px-2 py-1.5 font-mono text-[0.68rem] text-primary">{pitch.play_num ?? '—'}</td>
                  <td class="px-2 py-1.5 font-mono text-[0.68rem] text-muted">{pitch.seq_num ?? '—'}</td>
                  <td class="px-2 py-1.5 font-mono text-[0.68rem] text-muted">{pitch.pitch_type ?? '—'}</td>
                  <td class="px-2 py-1.5 font-mono text-[0.68rem] text-muted"
                    >{pitch.ball_count ?? 0}-{pitch.strike_count ?? 0}</td>
                  <td class="px-2 py-1.5 font-mono text-[0.68rem] text-foreground"
                    >{pitch.pitcher_name ?? pitch.pitcher ?? '—'}</td>
                  <td class="px-2 py-1.5 font-mono text-[0.68rem] text-muted">{pitch.description ?? '—'}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        {#if pitchesResource.total > pitchesPerPage}
          <div class="mt-2">
            <Pagination
              page={pitchesPage}
              perPage={pitchesPerPage}
              total={pitchesResource.total}
              onPageChange={(nextPage) => updateQuery({ pitches_page: nextPage })}
              onPerPageChange={(nextPerPage) => updateQuery({ pitches_page: 1, pitches_per_page: nextPerPage })} />
          </div>
        {/if}
      {/if}

      {#if selectedPlayNum != null}
        <div class="mt-3 rounded border border-outline bg-surface p-2.5">
          <div class="mb-1 font-mono text-[0.62rem] tracking-wider text-muted uppercase">
            Selected play pitch packet
          </div>
          {#if playPitchesResource.loading}
            <p class="font-mono text-[0.7rem] text-muted">Loading play-specific pitches…</p>
          {:else if playPitchesResource.error}
            <p class="font-mono text-[0.7rem] text-warning">{playPitchesResource.error}</p>
          {:else if playPitchesResource.items.length === 0}
            <p class="font-mono text-[0.7rem] text-muted">No dedicated play-pitch rows returned.</p>
          {:else}
            <p class="font-mono text-[0.7rem] text-foreground">
              /plays/{selectedPlayNum}/pitches returned {playPitchesResource.items.length.toLocaleString()} pitch rows.
            </p>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>
