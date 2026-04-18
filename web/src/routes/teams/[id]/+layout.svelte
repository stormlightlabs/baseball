<script lang="ts">
  import { page } from '$app/state';
  import TabRow from '$lib/components/TabRow.svelte';
  import { intParam } from '$lib/players/routing';
  import { ALL_TEAM_TABS, DEFAULT_TEAM_TAB, type TeamTabId } from '$lib/teams/constants';
  import { onMount, type Snippet } from 'svelte';
  import { cubicOut } from 'svelte/easing';
  import { crossfade, fly } from 'svelte/transition';

  const OUTLET_CROSSFADE_KEY = 'teams-tab-outlet';

  let { children }: { children: Snippet } = $props();

  let teamId = $derived(page.params.id ?? '');
  let q = $derived(page.url.searchParams.get('q') ?? '');
  let franchiseId = $derived(page.url.searchParams.get('franchise_id') ?? '');
  let year = $derived(page.url.searchParams.get('year') ?? '');

  let activeTab = $derived.by((): TeamTabId => {
    const value = page.url.pathname.split('/')[3] ?? '';
    const match = ALL_TEAM_TABS.find((tab) => tab.id === value);
    return match?.id ?? DEFAULT_TEAM_TAB;
  });

  let prefersReducedMotion = $state(false);

  let tabs = $derived(ALL_TEAM_TABS.map((tab) => ({ ...tab, href: tabHref(tab.id) })));

  let transitionKey = $derived.by(() => {
    const params = page.url.searchParams;
    const tab = activeTab;
    const keys = [`id=${teamId}`, `tab=${tab}`, `year=${year}`];
    if (q) keys.push(`q=${q}`);
    if (franchiseId) keys.push(`franchise_id=${franchiseId}`);
    if (tab === 'schedule' || tab === 'daily-trends') {
      keys.push(`page=${intParam(params, 'page', 1)}`, `per_page=${intParam(params, 'per_page', 25)}`);
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

  function tabHref(tabId: TeamTabId): string {
    const encodedId = encodeURIComponent(teamId);
    const base = `/teams/${encodedId}/${tabId}`;
    const overrides: Record<string, string> = {};
    if (q) overrides.q = q;
    if (year) overrides.year = year;
    if (franchiseId) overrides.franchise_id = franchiseId;
    const qs = new URLSearchParams(overrides).toString();
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
