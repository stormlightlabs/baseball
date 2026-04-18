<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, apiUrl, fetchPaginated } from '$lib/api';
  import { DEFAULT_TEAM_TAB, isTeamTabId } from '$lib/common/constants';
  import { endpointForTeamTab } from '$lib/common/tabs';
  import ApiPanel from '$lib/components/ApiPanel.svelte';
  import EraRangeChip from '$lib/components/EraRangeChip.svelte';
  import SearchInput from '$lib/components/SearchInput.svelte';
  import { EP } from '$lib/endpoints';
  import { erasInRange } from '$lib/eras';
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
  import { AsyncListResource, AsyncPaginatedListResource, AsyncValueResource } from '$lib/players/resources.svelte';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { normalizeFranchiseProfile, normalizeSearchTeamsPage, normalizeTeamResult } from '$lib/teams/normalizers';
  import type { FranchiseProfile, TeamResult } from '$lib/teams/types';
  import { onMount, type Snippet } from 'svelte';

  let { children }: { children: Snippet } = $props();

  let q = $derived(page.url.searchParams.get('q') ?? '');
  let teamId = $derived(page.params.id ?? '');
  let franchiseId = $derived(page.url.searchParams.get('franchise_id') ?? '');
  let year = $derived(page.url.searchParams.get('year') ?? '');
  let rawTab = $derived(page.url.pathname.split('/')[3] ?? '');

  let activeEndpoint = $derived.by(() => {
    const id = teamId;
    if (!id) return q.trim() ? EP.searchTeams : EP.franchises;
    const tab = isTeamTabId(rawTab) ? rawTab : DEFAULT_TEAM_TAB;
    const endpointId = tab === 'overview' && !year ? franchiseId || id : id;
    return endpointForTeamTab(endpointId, tab, year || null);
  });

  let activeUrl = $derived.by(() => {
    if (!teamId && q.trim()) {
      return apiUrl(activeEndpoint, { q: q.trim(), per_page: 20 });
    }
    if (rawTab === 'daily-trends' && year) {
      return apiUrl(activeEndpoint, {
        season: year,
        page: page.url.searchParams.get('page') ?? undefined,
        per_page: page.url.searchParams.get('per_page') ?? undefined
      });
    }
    if (rawTab === 'run-diff' && year) {
      return apiUrl(activeEndpoint, { season: year });
    }
    return apiUrl(activeEndpoint);
  });

  const searchResource = new AsyncPaginatedListResource<TeamResult>();
  const franchisesResource = new AsyncListResource<TeamResult>();
  const profileResource = new AsyncValueResource<FranchiseProfile>();

  let searchInput = $state('');
  let yearInput = $state('');
  let lastSearchKey = '';
  let lastProfileKey = '';

  let franchiseEras = $derived.by(() => {
    const p = profileResource.value;
    if (!p?.active_from) return [];
    return erasInRange(p.active_from, p.active_to ?? new Date().getFullYear());
  });

  async function refreshSearch(force = false): Promise<void> {
    const term = q.trim();
    if (!term) {
      searchResource.clear();
      lastSearchKey = '';
      return;
    }
    if (!force && term === lastSearchKey) return;
    lastSearchKey = term;

    await searchResource.load(async () => {
      const payload = await fetchPaginated<Record<string, unknown>>(EP.searchTeams, { q: term, per_page: 20 });
      return normalizeSearchTeamsPage(payload);
    });
  }

  async function refreshFranchises(): Promise<void> {
    if (franchisesResource.items.length > 0) return;
    await franchisesResource.load(async () => {
      const payload = await apiFetch<{ franchises?: Record<string, unknown>[] }>(EP.franchises);
      return (payload.franchises ?? []).map((r) => normalizeTeamResult(r));
    });
  }

  async function refreshProfile(force = false): Promise<void> {
    const selectedTeamId = teamId;
    if (!selectedTeamId) {
      profileResource.clear();
      lastProfileKey = '';
      return;
    }
    const selectedFranchiseId = franchiseId || selectedTeamId;
    const key = `${selectedTeamId}|${selectedFranchiseId}`;
    if (!force && key === lastProfileKey) return;
    lastProfileKey = key;

    await profileResource.load(async () => {
      try {
        const payload = await apiFetch<Record<string, unknown>>(EP.franchise(selectedFranchiseId));
        return normalizeFranchiseProfile(payload);
      } catch {
        const payload = await apiFetch<Record<string, unknown>>(EP.team(selectedTeamId));
        return normalizeFranchiseProfile(payload);
      }
    });
  }

  onMount(() => {
    searchInput = q;
    yearInput = year;
    void refreshSearch(true);
    void refreshFranchises();
    void refreshProfile(true);
  });

  afterNavigate(() => {
    searchInput = q;
    yearInput = year;
    void refreshSearch();
    void refreshProfile();
  });

  function teamQuerySuffix(nextFranchiseId?: string): string {
    const overrides: Record<string, string> = {};
    if (q.trim()) overrides.q = q.trim();
    if (year) overrides.year = year;
    const franchiseLookupId = nextFranchiseId?.trim() || franchiseId.trim();
    if (franchiseLookupId) overrides.franchise_id = franchiseLookupId;
    const qs = new URLSearchParams(overrides).toString();
    return qs ? `?${qs}` : '';
  }

  function teamResultIdSummary(result: TeamResult): string {
    const teamCode = result.team_id?.trim();
    const franchiseCode = result.franchise_id?.trim();
    if (teamCode && franchiseCode && teamCode !== franchiseCode) {
      return `team ${teamCode} · franchise ${franchiseCode}`;
    }
    return teamCode ?? franchiseCode ?? result.id;
  }

  function handleSearch(value = searchInput): void {
    const trimmed = value.trim();
    if (!trimmed) return;
    const href = `/teams?${new URLSearchParams({ q: trimmed }).toString()}`;
    void goto(resolve(href as '/teams'));
  }

  function handleYearChange(value: string): void {
    const trimmed = value.trim();
    if (!teamId) return;
    const encodedId = encodeURIComponent(teamId);
    const tab = isTeamTabId(rawTab) ? rawTab : DEFAULT_TEAM_TAB;
    const gotoWithYear = (basePath: string): void => {
      const href = withMergedQuery(basePath, page.url.searchParams, { year: trimmed || null });
      void goto(resolve(href as `/teams/${string}`), QUERY_NAV_OPTS);
    };

    switch (tab) {
      case 'overview': {
        gotoWithYear(`/teams/${encodedId}/overview`);
        return;
      }
      case 'roster': {
        gotoWithYear(`/teams/${encodedId}/roster`);
        return;
      }
      case 'batting': {
        gotoWithYear(`/teams/${encodedId}/batting`);
        return;
      }
      case 'pitching': {
        gotoWithYear(`/teams/${encodedId}/pitching`);
        return;
      }
      case 'fielding': {
        gotoWithYear(`/teams/${encodedId}/fielding`);
        return;
      }
      case 'schedule': {
        gotoWithYear(`/teams/${encodedId}/schedule`);
        return;
      }
      case 'daily-trends': {
        gotoWithYear(`/teams/${encodedId}/daily-trends`);
        return;
      }
      case 'run-diff': {
        gotoWithYear(`/teams/${encodedId}/run-diff`);
        return;
      }
    }
  }
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Teams & Franchises</div>
    <SearchInput mini bind:value={searchInput} placeholder="Team or franchise…" onsubmit={handleSearch} />

    {#if teamId}
      <div class="mt-4 rounded-lg border border-outline bg-surface p-4">
        {#if profileResource.loading}
          <p class="font-mono text-xs text-muted">Loading…</p>
        {:else if profileResource.error}
          <p class="font-mono text-xs text-warning">{profileResource.error}</p>
        {:else if profileResource.value}
          {#if profileResource.value.league}
            <div class="mb-0.5 text-center font-mono text-[0.6rem] tracking-wider text-muted uppercase">
              {profileResource.value.league}
            </div>
          {/if}
          <div class="mb-1 text-center font-display text-[0.95rem] font-medium text-foreground">
            {profileResource.value.name}
          </div>
          <div class="mb-2 text-center font-mono text-[0.65rem] text-muted">
            team: {teamId} · franchise: {franchiseId || profileResource.value.id}
          </div>
          {#if profileResource.value.active_from}
            <div class="mb-3 text-center font-mono text-[0.68rem] text-muted">
              {profileResource.value.active_from}–{profileResource.value.active_to ?? 'pres.'}
            </div>
          {/if}

          {#if franchiseEras.length > 0}
            <div class="mb-3 border-t border-outline pt-3">
              <div class="mb-1.5 font-mono text-[0.6rem] tracking-wider text-muted uppercase">Eras active</div>
              <div class="flex flex-wrap gap-1">
                {#each franchiseEras as era (era.code)}
                  <EraRangeChip {era} />
                {/each}
              </div>
            </div>
          {/if}

          <div class="border-t border-outline pt-3">
            <div class="mb-1.5 font-mono text-[0.6rem] tracking-wider text-muted uppercase">Season year</div>
            <input
              type="number"
              min="1871"
              max="2025"
              value={yearInput}
              oninput={(e) => {
                yearInput = (e.target as HTMLInputElement).value;
              }}
              onchange={(e) => handleYearChange((e.target as HTMLInputElement).value)}
              onkeydown={(e) => {
                if (e.key === 'Enter') handleYearChange((e.target as HTMLInputElement).value);
              }}
              class="w-full rounded border border-outline bg-mantle px-2 py-1.5 font-mono text-xs text-foreground placeholder-muted/50 focus:ring-1 focus:ring-primary/50 focus:outline-none"
              placeholder="e.g. 2021" />
          </div>
        {/if}
      </div>
    {/if}

    {#if searchResource.items.length > 0}
      <div class="mt-4 panel-label">Results</div>
      <div class="flex flex-col gap-0.5">
        {#each searchResource.items as result (result.id)}
          <a
            href={resolve(
              `/teams/${encodeURIComponent(result.id)}/overview${teamQuerySuffix(result.franchise_id)}` as `/teams/${string}/overview`
            )}
            class="rounded-md px-3 py-2 text-left no-underline transition-colors hover:bg-surface {teamId === result.id
              ? 'bg-surface'
              : ''}">
            <div class="font-display text-[0.8rem] text-foreground">{result.name}</div>
            <div class="font-mono text-[0.68rem] text-muted">
              {teamResultIdSummary(result)}{#if result.league}
                · {result.league}{/if}{#if result.year}
                · {result.year}{/if}
            </div>
          </a>
        {/each}
      </div>
    {:else if searchResource.loading}
      <p class="mt-3 font-mono text-xs text-muted">Searching…</p>
    {:else if !q && !teamId}
      {#if franchisesResource.loading}
        <p class="mt-3 font-mono text-xs text-muted">Loading franchises…</p>
      {:else if franchisesResource.items.length > 0}
        <div class="mt-4 panel-label">All Franchises</div>
        <div class="flex flex-col gap-0.5">
          {#each franchisesResource.items as franchise (franchise.id)}
            <a
              href={resolve(
                `/teams/${encodeURIComponent(franchise.id)}/overview${teamQuerySuffix(franchise.id)}` as `/teams/${string}/overview`
              )}
              class="rounded-md px-3 py-2 text-left no-underline transition-colors hover:bg-surface">
              <div class="font-display text-[0.8rem] text-foreground">{franchise.name}</div>
              <div class="font-mono text-[0.68rem] text-muted">
                {franchise.id}{#if franchise.league}
                  · {franchise.league}{/if}{#if franchise.active_from}
                  · {franchise.active_from}–{franchise.active_to ?? 'pres.'}{/if}
              </div>
            </a>
          {/each}
        </div>
      {:else}
        <p class="mt-4 font-mono text-xs text-muted">Search for a team or franchise to begin.</p>
      {/if}
    {/if}
  {/snippet}

  {#snippet center()}
    {@render children()}
  {/snippet}

  {#snippet panel()}
    <ApiPanel endpoint={`/v1${activeEndpoint}`} url={activeUrl} />
  {/snippet}
</ThreeColLayout>
