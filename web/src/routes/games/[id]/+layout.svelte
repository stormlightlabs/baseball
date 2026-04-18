<script lang="ts">
  import { page } from '$app/state';
  import { ALL_GAME_TABS, DEFAULT_GAME_TAB, type GameTabId } from '$lib/common/constants';
  import TabRow from '$lib/components/TabRow.svelte';
  import { intParam } from '$lib/url-state.svelte';
  import { onMount, type Snippet } from 'svelte';
  import { cubicOut } from 'svelte/easing';
  import { crossfade, fly } from 'svelte/transition';

  const OUTLET_CROSSFADE_KEY = 'games-tab-outlet';

  let { children }: { children: Snippet } = $props();

  let gameId = $derived(page.params.id ?? '');
  let activeTab = $derived.by((): GameTabId => {
    const value = page.url.pathname.split('/')[3] ?? '';
    const match = ALL_GAME_TABS.find((tab) => tab.id === value);
    return match?.id ?? DEFAULT_GAME_TAB;
  });

  let prefersReducedMotion = $state(false);

  let tabs = $derived(ALL_GAME_TABS.map((tab) => ({ ...tab, href: tabHref(tab.id) })));

  let transitionKey = $derived.by(() => {
    const params = page.url.searchParams;
    const keys = [`id=${gameId}`, `tab=${activeTab}`];

    if (activeTab === 'events') {
      keys.push(
        `events_page=${intParam(params, 'events_page', 1)}`,
        `events_per_page=${intParam(params, 'events_per_page', 30)}`,
        `event_seq=${params.get('event_seq') ?? ''}`
      );
    }

    if (activeTab === 'plays') {
      keys.push(
        `plays_page=${intParam(params, 'plays_page', 1)}`,
        `plays_per_page=${intParam(params, 'plays_per_page', 25)}`,
        `pitches_page=${intParam(params, 'pitches_page', 1)}`,
        `pitches_per_page=${intParam(params, 'pitches_per_page', 30)}`,
        `play_num=${params.get('play_num') ?? ''}`
      );
    }

    return keys.join('|');
  });

  const [send, receive] = crossfade({
    duration: (distance) => (prefersReducedMotion ? 0 : Math.min(260, 140 + distance / 6)),
    easing: cubicOut,
    fallback(node, _params, intro) {
      return fly(node, { y: intro ? 8 : -8, opacity: 0.2, duration: prefersReducedMotion ? 0 : 170, easing: cubicOut });
    }
  });

  onMount(() => {
    const media = globalThis.matchMedia('(prefers-reduced-motion: reduce)');
    const sync = () => {
      prefersReducedMotion = media.matches;
    };
    sync();
    media.addEventListener('change', sync);
    return () => {
      media.removeEventListener('change', sync);
    };
  });

  function tabHref(tabId: GameTabId): string {
    const encodedId = encodeURIComponent(gameId);
    const base = `/games/${encodedId}/${tabId}`;
    const qs = page.url.searchParams.toString();
    return qs ? `${base}?${qs}` : base;
  }
</script>

<div class="mb-1">
  <TabRow {tabs} active={activeTab} />
</div>

<div class="relative">
  {#key transitionKey}
    <div in:receive={{ key: OUTLET_CROSSFADE_KEY }} out:send={{ key: OUTLET_CROSSFADE_KEY }}>
      {@render children()}
    </div>
  {/key}
</div>
