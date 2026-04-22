<script lang="ts">
  import { page } from '$app/state';
  import { ALL_SEASON_TABS, DEFAULT_SEASON_TAB, type SeasonTabId } from '$lib/common/constants';
  import TabRow from '$lib/components/TabRow.svelte';
  import { onMount, type Snippet } from 'svelte';
  import { cubicOut } from 'svelte/easing';
  import { crossfade, fly } from 'svelte/transition';

  const OUTLET_CROSSFADE_KEY = 'seasons-tab-outlet';

  let { children }: { children: Snippet } = $props();

  let selectedYear = $derived(page.params.year ?? '');
  let activeTab = $derived.by((): SeasonTabId => {
    const value = page.url.pathname.split('/')[3] ?? '';
    const match = ALL_SEASON_TABS.find((tab) => tab.id === value);
    return match?.id ?? DEFAULT_SEASON_TAB;
  });

  let prefersReducedMotion = $state(false);

  let tabs = $derived(ALL_SEASON_TABS.map((tab) => ({ ...tab, href: tabHref(tab.id) })));

  let transitionKey = $derived.by(() => {
    const params = page.url.searchParams;
    const keys = [`year=${selectedYear}`, `tab=${activeTab}`];

    if (activeTab === 'leaders') {
      keys.push(
        `league=${params.get('league') ?? ''}`,
        `bat=${params.get('bat') ?? ''}`,
        `pit=${params.get('pit') ?? ''}`
      );
    }

    if (activeTab === 'schedule') {
      keys.push(`league=${params.get('league') ?? ''}`, `date=${params.get('date') ?? ''}`);
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

  function tabHref(tabId: SeasonTabId): string {
    const encodedYear = encodeURIComponent(selectedYear);
    const base = tabId === 'overview' ? `/seasons/${encodedYear}` : `/seasons/${encodedYear}/${tabId}`;
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
