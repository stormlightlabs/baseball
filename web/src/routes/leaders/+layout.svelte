<script lang="ts">
  import { page } from '$app/state';
  import { ALL_LEADER_TABS, DEFAULT_LEADER_TAB, type LeaderTabId } from '$lib/common/constants';
  import PromoBanner from '$lib/components/PromoBanner.svelte';
  import TabRow from '$lib/components/TabRow.svelte';
  import { onMount, type Snippet } from 'svelte';
  import { cubicOut } from 'svelte/easing';
  import { crossfade, fly } from 'svelte/transition';

  const OUTLET_CROSSFADE_KEY = 'leaders-tab-outlet';

  let { children }: { children: Snippet } = $props();

  let activeTab = $derived.by((): LeaderTabId => {
    const value = page.url.pathname.split('/')[2] ?? '';
    const match = ALL_LEADER_TABS.find((tab) => tab.id === value);
    return match?.id ?? DEFAULT_LEADER_TAB;
  });

  let prefersReducedMotion = $state(false);

  let tabs = $derived(ALL_LEADER_TABS.map((tab) => ({ ...tab, href: tabHref(tab.id) })));
  let transitionKey = $derived(`tab=${activeTab}|${page.url.searchParams.toString()}`);

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

  function tabHref(tabId: LeaderTabId): string {
    const base = `/leaders/${tabId}`;
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

<div class="px-4 pt-3 pb-4 sm:px-6">
  <PromoBanner variant="horizontal" />
</div>
