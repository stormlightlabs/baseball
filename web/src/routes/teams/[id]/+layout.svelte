<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import { ALL_TEAM_TABS, DEFAULT_TEAM_TAB, type TeamTabId } from '$lib/common/constants';
  import TabRow from '$lib/components/TabRow.svelte';
  import { EP } from '$lib/endpoints';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { normalizeTeamSeasonProfile } from '$lib/teams/normalizers';
  import { intParam } from '$lib/url-state.svelte';
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
  let isCanonicalizing = false;
  let lastCanonicalKey = '';

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

  onMount(() => {
    void canonicalizeTeamRoute(true);
  });

  afterNavigate(() => {
    void canonicalizeTeamRoute();
  });

  function nonEmpty(value: unknown): string | null {
    if (typeof value !== 'string') return null;
    const trimmed = value.trim();
    return trimmed || null;
  }

  type TeamLookupPayload = { id?: unknown; team_id?: unknown; franchise_id?: unknown };

  async function resolveIdentifiers(
    routeTeamId: string,
    selectedYear: string,
    knownFranchiseId: string
  ): Promise<{ teamId: string; franchiseId: string | null }> {
    const cleanedYear = selectedYear.trim();

    try {
      const payload = await apiFetch<Record<string, unknown>>(
        EP.team(routeTeamId),
        cleanedYear ? { year: cleanedYear } : undefined
      );
      const profile = normalizeTeamSeasonProfile(payload);
      return {
        teamId: nonEmpty(profile.id) ?? routeTeamId,
        franchiseId: nonEmpty(profile.franchise_id) ?? nonEmpty(knownFranchiseId)
      };
    } catch {
      if (!cleanedYear) {
        return { teamId: routeTeamId, franchiseId: nonEmpty(knownFranchiseId) };
      }
    }

    try {
      const payload = await apiFetch<{ data?: TeamLookupPayload[] }>(EP.seasonTeams(cleanedYear), { per_page: 500 });
      const rows = Array.isArray(payload.data) ? payload.data : [];
      const franchiseLookup = nonEmpty(knownFranchiseId) ?? routeTeamId;

      const match = rows.find((row) => {
        const rowTeamId = nonEmpty(row.team_id);
        const rowId = nonEmpty(row.id);
        const rowFranchiseId = nonEmpty(row.franchise_id);
        if (rowTeamId === routeTeamId || rowId === routeTeamId) return true;
        if (!franchiseLookup) return false;
        return rowFranchiseId === franchiseLookup || rowId === franchiseLookup;
      });

      return {
        teamId: nonEmpty(match?.team_id) ?? nonEmpty(match?.id) ?? routeTeamId,
        franchiseId: nonEmpty(match?.franchise_id) ?? nonEmpty(knownFranchiseId)
      };
    } catch {
      return { teamId: routeTeamId, franchiseId: nonEmpty(knownFranchiseId) };
    }
  }

  async function canonicalizeTeamRoute(force = false): Promise<void> {
    if (isCanonicalizing) return;
    const currentTeamId = teamId;
    if (!currentTeamId) return;

    const currentFranchiseId = franchiseId;
    const currentYear = year;
    const key = `${currentTeamId}|${currentFranchiseId}|${currentYear}`;
    if (!force && key === lastCanonicalKey) return;
    lastCanonicalKey = key;

    isCanonicalizing = true;
    try {
      const { teamId: nextTeamId, franchiseId: nextFranchiseId } = await resolveIdentifiers(
        currentTeamId,
        currentYear,
        currentFranchiseId
      );

      const normalizedFranchiseId = nextFranchiseId ?? '';
      const needsTeamUpdate = nextTeamId !== currentTeamId;
      const needsFranchiseUpdate = normalizedFranchiseId !== currentFranchiseId;
      if (!needsTeamUpdate && !needsFranchiseUpdate) return;

      const href = withMergedQuery(
        `/teams/${encodeURIComponent(nextTeamId)}/${activeTab}`,
        page.url.searchParams,
        { franchise_id: normalizedFranchiseId || null },
        page.url.hash
      );

      void goto(resolve(href as `/teams/${string}/${string}`), QUERY_NAV_OPTS);
    } finally {
      isCanonicalizing = false;
    }
  }

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
