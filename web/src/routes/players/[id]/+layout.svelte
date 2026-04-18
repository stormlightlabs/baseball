<script lang="ts">
  import { page } from '$app/state';
  import TabRow from '$lib/components/TabRow.svelte';
  import { ADV_TABS, ALL_PLAYER_TABS, DEFAULT_PLAYER_TAB, MAIN_TABS, type PlayerTabId } from '$lib/players/constants';
  import { intParam } from '$lib/players/routing';
  import { parseGameLogType } from '$lib/players/tab-endpoints';
  import { onMount, type Snippet } from 'svelte';
  import { cubicOut } from 'svelte/easing';
  import { crossfade, fly } from 'svelte/transition';

  const ADVANCED_TAB_STORAGE_KEY = 'players.showAdvancedTabs';
  const OUTLET_CROSSFADE_KEY = 'players-tab-outlet';

  let { children }: { children: Snippet } = $props();

  let playerId = $derived(page.params.id ?? '');
  let q = $derived(page.url.searchParams.get('q') ?? '');

  let activeTab = $derived.by((): PlayerTabId => {
    const value = page.url.pathname.split('/')[3] ?? '';
    const match = ALL_PLAYER_TABS.find((tab) => tab.id === value);
    return match?.id ?? DEFAULT_PLAYER_TAB;
  });

  let showAdvanced = $state(false);
  let prefersReducedMotion = $state(false);

  let navTabs = $derived.by(() => {
    if (showAdvanced) return ALL_PLAYER_TABS;
    if (ADV_TABS.some((tab) => tab.id === activeTab)) {
      return [...MAIN_TABS, ...ADV_TABS.filter((tab) => tab.id === activeTab)];
    }
    return MAIN_TABS;
  });

  let tabs = $derived(navTabs.map((tab) => ({ ...tab, href: tabHref(tab.id) })));

  let transitionKey = $derived.by(() => {
    const params = page.url.searchParams;
    const tab = activeTab;

    const keys = [`id=${playerId}`, `tab=${tab}`];
    if (q) keys.push(`q=${q}`);

    if (tab === 'batting' || tab === 'pitching' || tab === 'game-logs') {
      keys.push(`page=${intParam(params, 'page', 1)}`, `per_page=${intParam(params, 'per_page', 20)}`);
    }

    if (tab === 'batting') {
      keys.push(`stat=${params.get('stat') ?? 'hr'}`);
    }

    if (tab === 'game-logs') {
      keys.push(`log=${parseGameLogType(params.get('log'))}`);
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
    const saved = localStorage.getItem(ADVANCED_TAB_STORAGE_KEY);
    if (saved === '1') showAdvanced = true;

    const media = globalThis.matchMedia('(prefers-reduced-motion: reduce)');
    const syncReducedMotion = () => {
      prefersReducedMotion = media.matches;
    };

    syncReducedMotion();
    media.addEventListener('change', syncReducedMotion);

    return () => {
      media.removeEventListener('change', syncReducedMotion);
    };
  });

  function tabHref(tabId: PlayerTabId): string {
    if (!playerId) return `/players?tab=${tabId}`;
    const encodedId = encodeURIComponent(playerId);
    const base = `/players/${encodedId}/${tabId}`;
    if (!q) return base;
    const qs = new URLSearchParams({ q }).toString();
    return `${base}?${qs}`;
  }

  function toggleAdvanced(): void {
    showAdvanced = !showAdvanced;
    localStorage.setItem(ADVANCED_TAB_STORAGE_KEY, showAdvanced ? '1' : '0');
  }
</script>

<div class="mb-1 flex flex-wrap items-center gap-2">
  <TabRow {tabs} active={activeTab} />
  <button
    onclick={toggleAdvanced}
    class="ml-auto shrink-0 rounded border px-2.5 py-1 font-mono text-[0.68rem] transition-colors {showAdvanced
      ? 'border-primary/40 text-primary hover:border-primary'
      : 'border-outline text-muted hover:border-primary hover:text-foreground'}">
    {showAdvanced ? 'Hide advanced' : 'Show advanced'}
  </button>
</div>

<div class="relative">
  {#key transitionKey}
    <div in:receive={{ key: OUTLET_CROSSFADE_KEY }} out:send={{ key: OUTLET_CROSSFADE_KEY }}>
      {@render children()}
    </div>
  {/key}
</div>
