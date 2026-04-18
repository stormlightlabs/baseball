<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, fetchPaginated } from '$lib/api';
  import Pagination from '$lib/components/Pagination.svelte';
  import { EP } from '$lib/endpoints';
  import { normalizePlayEvent, normalizePlayPage } from '$lib/games/normalizers';
  import type { PlayEvent } from '$lib/games/types';
  import { AsyncPaginatedListResource, AsyncValueResource } from '$lib/players/resources.svelte';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { intParam } from '$lib/url-state.svelte';
  import { onMount } from 'svelte';

  const DEFAULT_EVENTS_PAGE = 1;
  const DEFAULT_EVENTS_PER_PAGE = 30;

  let gameId = $derived(page.params.id ?? '');
  let eventsPage = $derived(intParam(page.url.searchParams, 'events_page', DEFAULT_EVENTS_PAGE));
  let eventsPerPage = $derived(intParam(page.url.searchParams, 'events_per_page', DEFAULT_EVENTS_PER_PAGE));
  let eventSeq = $derived(page.url.searchParams.get('event_seq') ?? '');

  const eventsResource = new AsyncPaginatedListResource<PlayEvent>();
  const singleEventResource = new AsyncValueResource<PlayEvent>();

  let eventSeqInput = $state('');
  let lastEventsKey = '';
  let lastEventKey = '';

  onMount(() => {
    eventSeqInput = eventSeq;
    void refreshEvents(true);
    void refreshSingleEvent(true);
  });

  afterNavigate(() => {
    eventSeqInput = eventSeq;
    void refreshEvents();
    void refreshSingleEvent();
  });

  async function refreshEvents(force = false): Promise<void> {
    const id = gameId;
    if (!id) {
      eventsResource.clear();
      lastEventsKey = '';
      return;
    }

    const key = `${id}|${eventsPage}|${eventsPerPage}`;
    if (!force && key === lastEventsKey) return;
    lastEventsKey = key;

    await eventsResource.load(async () => {
      const payload = await fetchPaginated<Record<string, unknown>>(EP.gameEvents(id), {
        page: eventsPage,
        per_page: eventsPerPage
      });
      return normalizePlayPage(payload);
    });
  }

  async function refreshSingleEvent(force = false): Promise<void> {
    const id = gameId;
    const seq = Number.parseInt(eventSeq, 10);
    if (!id || Number.isNaN(seq) || seq <= 0) {
      singleEventResource.clear();
      lastEventKey = '';
      return;
    }

    const key = `${id}|${seq}`;
    if (!force && key === lastEventKey) return;
    lastEventKey = key;

    await singleEventResource.load(async () => {
      const payload = await apiFetch<Record<string, unknown>>(EP.gameEvent(id, seq));
      return normalizePlayEvent(payload);
    });
  }

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const href = withMergedQuery(
      `/games/${encodeURIComponent(gameId)}/events`,
      page.url.searchParams,
      overrides,
      page.url.hash
    );
    void goto(resolve(href as `/games/${string}/events`), QUERY_NAV_OPTS);
  }

  function commitEventSeq(): void {
    const seq = Number.parseInt(eventSeqInput.trim(), 10);
    if (Number.isNaN(seq) || seq <= 0) {
      singleEventResource.error = 'Enter a positive event sequence number.';
      return;
    }
    singleEventResource.error = null;
    updateQuery({ event_seq: seq });
  }

  function jumpToSequence(playNum: number | undefined): void {
    if (playNum == null) return;
    updateQuery({ event_seq: playNum });
  }

  function fmtTopBottom(topBot: number | undefined): string {
    if (topBot == null) return '—';
    return topBot === 0 ? 'Top' : 'Bot';
  }
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-3">Event Stream</div>

  <div class="mb-3 grid gap-3 md:grid-cols-[1fr_auto]">
    <div>
      <p class="font-mono text-[0.69rem] text-muted">
        Browse paginated events from <code>/api/v1/games/{gameId}/events</code>.
      </p>
    </div>
    <div class="rounded border border-outline bg-surface px-2.5 py-2">
      <div class="mb-1 font-mono text-[0.62rem] text-muted uppercase">Single event fetch</div>
      <div class="flex gap-1.5">
        <input
          bind:value={eventSeqInput}
          placeholder="event_seq"
          class="w-24 rounded border border-outline bg-crust px-2 py-1 font-mono text-[0.72rem] text-foreground"
          onkeydown={(e) => {
            if (e.key === 'Enter') commitEventSeq();
          }} />
        <button
          onclick={commitEventSeq}
          class="rounded border border-primary/40 bg-primary/10 px-2.5 py-1 font-mono text-[0.68rem] text-foreground">
          Fetch
        </button>
      </div>
    </div>
  </div>

  {#if singleEventResource.error}
    <p class="mb-2 font-mono text-[0.72rem] text-warning">{singleEventResource.error}</p>
  {:else if singleEventResource.value}
    <div class="mb-3 rounded border border-outline bg-surface px-3 py-2">
      <div class="font-mono text-[0.67rem] text-muted">Single event #{singleEventResource.value.play_num ?? '—'}</div>
      <div class="font-mono text-[0.72rem] text-foreground">{singleEventResource.value.event ?? 'No event text'}</div>
    </div>
  {/if}

  {#if eventsResource.loading}
    <p class="font-mono text-[0.78rem] text-muted">Loading events…</p>
  {:else if eventsResource.error}
    <p class="font-mono text-[0.78rem] text-warning">{eventsResource.error}</p>
  {:else if eventsResource.items.length === 0}
    <p class="font-mono text-[0.78rem] text-muted">No event rows returned.</p>
  {:else}
    <div class="overflow-x-auto">
      <table class="w-full border-collapse text-[0.72rem]">
        <thead>
          <tr>
            {#each ['Seq', 'Inn', 'Half', 'Score', 'Batter', 'Pitcher', 'Event'] as col (col)}
              <th
                class="border-b border-outline px-2 py-1.5 text-left font-sans text-[0.69rem] font-medium whitespace-nowrap text-muted">
                {col}
              </th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each eventsResource.items as event (`${event.play_num ?? ''}-${event.event ?? ''}`)}
            <tr
              class="cursor-pointer border-b border-outline last:border-b-0 hover:[&>td]:bg-surface"
              onclick={() => jumpToSequence(event.play_num)}>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-primary">{event.play_num ?? '—'}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-muted">{event.inning ?? '—'}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-muted">{fmtTopBottom(event.top_bot)}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-muted"
                >{event.score_vis ?? 0}-{event.score_home ?? 0}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-foreground"
                >{event.batter_name ?? event.batter ?? '—'}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-foreground"
                >{event.pitcher_name ?? event.pitcher ?? '—'}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-muted">{event.event ?? '—'}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    {#if eventsResource.total > eventsPerPage}
      <div class="mt-3">
        <Pagination
          page={eventsPage}
          perPage={eventsPerPage}
          total={eventsResource.total}
          onPageChange={(nextPage) => updateQuery({ events_page: nextPage })}
          onPerPageChange={(nextPerPage) => updateQuery({ events_page: 1, events_per_page: nextPerPage })} />
      </div>
    {/if}
  {/if}
</div>
