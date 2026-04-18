<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, apiUrl, fetchPaginated } from '$lib/api';
  import ApiPanel from '$lib/components/ApiPanel.svelte';
  import EraRangeChip from '$lib/components/EraRangeChip.svelte';
  import SearchInput from '$lib/components/SearchInput.svelte';
  import { EP } from '$lib/endpoints';
  import { erasInRange } from '$lib/eras';
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
  import type { ApiPlayerPayload } from '$lib/players/api-payloads';
  import { isPlayerTabId } from '$lib/players/constants';
  import { normalizePlayerProfile, normalizeSearchPlayersPage } from '$lib/players/normalizers';
  import { AsyncPaginatedListResource, AsyncValueResource } from '$lib/players/resources.svelte';
  import { endpointForPlayerTab, parseGameLogType } from '$lib/players/tab-endpoints';
  import type { PlayerProfile, PlayerResult } from '$lib/players/types';
  import { onMount, type Snippet } from 'svelte';

  let { children }: { children: Snippet } = $props();

  let q = $derived(page.url.searchParams.get('q') ?? '');
  let playerId = $derived(page.params.id ?? '');
  let rawTab = $derived(page.url.pathname.split('/')[3] ?? '');
  let gameLogType = $derived(parseGameLogType(page.url.searchParams.get('log')));

  let activeEndpoint = $derived.by(() => {
    const id = playerId;
    if (!id) return EP.searchPlayers;

    if (!rawTab || !isPlayerTabId(rawTab)) return EP.player(id);
    return endpointForPlayerTab(id, rawTab, gameLogType);
  });

  let activeUrl = $derived(apiUrl(activeEndpoint));

  const searchResource = new AsyncPaginatedListResource<PlayerResult>();
  const profileResource = new AsyncValueResource<PlayerProfile>();

  let searchInput = $state('');
  let lastSearchKey = '';
  let lastProfileKey = '';

  let careerEras = $derived.by(() => {
    if (!profileResource.value?.debut_year || !profileResource.value?.final_year) return [];
    return erasInRange(profileResource.value.debut_year, profileResource.value.final_year);
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
      const payload = await fetchPaginated<ApiPlayerPayload>(EP.searchPlayers, { q: term, per_page: 20 });
      return normalizeSearchPlayersPage(payload);
    });
  }

  async function refreshProfile(force = false): Promise<void> {
    const id = playerId;
    if (!id) {
      profileResource.clear();
      lastProfileKey = '';
      return;
    }

    if (!force && id === lastProfileKey) return;
    lastProfileKey = id;

    await profileResource.load(async () => {
      const payload = await apiFetch<ApiPlayerPayload>(EP.player(id));
      return normalizePlayerProfile(payload);
    });
  }

  onMount(() => {
    searchInput = q;
    void refreshSearch(true);
    void refreshProfile(true);
  });

  afterNavigate(() => {
    searchInput = q;
    void refreshSearch();
    void refreshProfile();
  });

  function searchQuerySuffix(): string {
    const trimmed = q.trim();
    if (!trimmed) return '';
    return `?${new URLSearchParams({ q: trimmed }).toString()}`;
  }

  function handleSearch(value = searchInput): void {
    const trimmed = value.trim();
    if (!trimmed) return;
    const href = `/players?${new URLSearchParams({ q: trimmed }).toString()}`;
    void goto(resolve(href as '/players'));
  }

  const fmtAvg = (v: number | undefined) => (v != null ? Number(v).toFixed(3) : '—');
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Player search</div>
    <SearchInput mini bind:value={searchInput} placeholder="Name, ID…" onsubmit={handleSearch} />

    {#if playerId}
      <div class="mt-4 rounded-lg border border-outline bg-surface p-4">
        {#if profileResource.loading}
          <p class="font-mono text-xs text-muted">Loading…</p>
        {:else if profileResource.error}
          <p class="font-mono text-xs text-warning">{profileResource.error}</p>
        {:else if profileResource.value}
          <div class="mb-0.5 text-center font-mono text-[0.6rem] tracking-wider text-muted uppercase">
            {profileResource.value.primary_position ?? profileResource.value.positions?.[0] ?? ''}
          </div>
          <div class="mb-1 text-center font-display text-[0.95rem] font-medium text-foreground">
            {profileResource.value.name}
          </div>
          {#if profileResource.value.birth_date || profileResource.value.birth_city}
            <div class="mb-1 text-center font-mono text-[0.68rem] text-muted">
              {profileResource.value.birth_date ?? ''}
              {#if profileResource.value.birth_city}
                · {profileResource.value.birth_city}{#if profileResource.value.birth_state}, {profileResource.value
                    .birth_state}{/if}{/if}
            </div>
          {/if}
          {#if profileResource.value.bats || profileResource.value.throws || profileResource.value.debut_year}
            <div class="mb-3 text-center font-mono text-[0.68rem] text-muted">
              {#if profileResource.value.bats}Bats: {profileResource.value.bats}{/if}
              {#if profileResource.value.bats && profileResource.value.throws}
                ·
              {/if}
              {#if profileResource.value.throws}Throws: {profileResource.value.throws}{/if}
              {#if profileResource.value.debut_year}
                · {profileResource.value.debut_year}–{profileResource.value.final_year ?? 'pres.'}{/if}
            </div>
          {/if}
          {#if profileResource.value.career_hr != null || profileResource.value.career_avg != null || profileResource.value.career_rbi != null}
            <div class="mb-3 grid grid-cols-3 gap-1 text-center">
              {#if profileResource.value.career_hr != null}
                <div>
                  <div class="font-display text-[0.95rem] font-semibold text-foreground">
                    {profileResource.value.career_hr}
                  </div>
                  <div class="font-mono text-[0.58rem] text-muted uppercase">HR</div>
                </div>
              {/if}
              {#if profileResource.value.career_avg != null}
                <div>
                  <div class="font-display text-[0.95rem] font-semibold text-foreground">
                    {fmtAvg(profileResource.value.career_avg)}
                  </div>
                  <div class="font-mono text-[0.58rem] text-muted uppercase">AVG</div>
                </div>
              {/if}
              {#if profileResource.value.career_rbi != null}
                <div>
                  <div class="font-display text-[0.95rem] font-semibold text-foreground">
                    {profileResource.value.career_rbi}
                  </div>
                  <div class="font-mono text-[0.58rem] text-muted uppercase">RBI</div>
                </div>
              {/if}
            </div>
          {/if}

          {#if careerEras.length > 0}
            <div class="border-t border-outline pt-3">
              <div class="mb-1.5 font-mono text-[0.6rem] tracking-wider text-muted uppercase">Career eras</div>
              <div class="flex flex-wrap gap-1">
                {#each careerEras as era (era.code)}
                  <EraRangeChip {era} />
                {/each}
              </div>
            </div>
          {/if}
        {/if}
      </div>
    {/if}

    {#if searchResource.items.length > 0}
      <div class="mt-4 panel-label">Results</div>
      <div class="flex flex-col gap-0.5">
        {#each searchResource.items as result (result.id)}
          <a
            href={resolve(
              `/players/${encodeURIComponent(result.id)}/batting${searchQuerySuffix()}` as `/players/${string}/batting`
            )}
            class="rounded-md px-3 py-2 text-left no-underline transition-colors hover:bg-surface {playerId ===
            result.id
              ? 'bg-surface'
              : ''}">
            <div class="font-display text-[0.8rem] text-foreground">{result.name}</div>
            <div class="font-mono text-[0.68rem] text-muted">
              {result.id}
              {#if result.primary_position ?? result.position}
                · {result.primary_position ?? result.position}{/if}
              {#if result.debut_year}
                · {result.debut_year}–{result.final_year ?? 'pres.'}{/if}
            </div>
          </a>
        {/each}
      </div>
    {:else if searchResource.loading}
      <p class="mt-3 font-mono text-xs text-muted">Searching…</p>
    {:else if !q && !playerId}
      <p class="mt-4 font-mono text-xs text-muted">Search for a player to begin.</p>
    {/if}
  {/snippet}

  {#snippet center()}
    {@render children()}
  {/snippet}

  {#snippet panel()}
    <ApiPanel endpoint={`/v1${activeEndpoint}`} url={activeUrl} />
  {/snippet}
</ThreeColLayout>
