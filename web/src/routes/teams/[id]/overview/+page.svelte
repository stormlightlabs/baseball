<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import EraRangeChip from '$lib/components/EraRangeChip.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear, erasInRange } from '$lib/eras';
  import { AsyncValueResource } from '$lib/players/resources.svelte';
  import { normalizeFranchiseProfile, normalizeTeamSeasonProfile } from '$lib/teams/normalizers';
  import type { FranchiseProfile, TeamSeasonProfile } from '$lib/teams/types';
  import { onMount } from 'svelte';

  let teamId = $derived(page.params.id ?? '');
  let year = $derived(page.url.searchParams.get('year') ?? '');

  const franchiseResource = new AsyncValueResource<FranchiseProfile>();
  const seasonResource = new AsyncValueResource<TeamSeasonProfile>();
  let lastFranchiseKey = '';
  let lastSeasonKey = '';

  let franchiseEras = $derived.by(() => {
    const f = franchiseResource.value;
    if (!f?.active_from) return [];
    return erasInRange(f.active_from, f.active_to ?? new Date().getFullYear());
  });

  let seasonEra = $derived.by(() => {
    const y = seasonResource.value?.year ?? (year ? Number(year) : null);
    return y ? eraForYear(y) : null;
  });

  async function refreshFranchise(force = false): Promise<void> {
    const id = teamId;
    if (!id) {
      franchiseResource.clear();
      return;
    }
    if (!force && id === lastFranchiseKey) return;
    lastFranchiseKey = id;

    await franchiseResource.load(async () => {
      try {
        const payload = await apiFetch<Record<string, unknown>>(EP.franchise(id));
        return normalizeFranchiseProfile(payload);
      } catch {
        const payload = await apiFetch<Record<string, unknown>>(EP.team(id));
        return normalizeFranchiseProfile(payload);
      }
    });
  }

  async function refreshSeason(force = false): Promise<void> {
    const id = teamId;
    const y = year;
    if (!id || !y) {
      seasonResource.clear();
      return;
    }
    const key = `${id}|${y}`;
    if (!force && key === lastSeasonKey) return;
    lastSeasonKey = key;

    await seasonResource.load(async () => {
      const payload = await apiFetch<Record<string, unknown>>(EP.team(id), { year: y });
      return normalizeTeamSeasonProfile(payload);
    });
  }

  onMount(() => {
    void refreshFranchise(true);
    void refreshSeason(true);
  });

  afterNavigate(() => {
    void refreshFranchise();
    void refreshSeason();
  });

  function fmtRecord(wins?: number, losses?: number): string {
    if (wins == null && losses == null) return '—';
    return `${wins ?? '?'}–${losses ?? '?'}`;
  }
</script>

{#if franchiseResource.loading}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Loading…</p>
{:else if franchiseResource.error}
  <p class="mt-4 font-mono text-[0.78rem] text-warning">{franchiseResource.error}</p>
{:else if franchiseResource.value}
  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">Franchise</div>
    <div class="grid grid-cols-2 gap-x-6 gap-y-2 font-mono text-xs">
      <div>
        <div class="text-[0.65rem] tracking-wider text-muted uppercase">Name</div>
        <div class="text-foreground">{franchiseResource.value.name}</div>
      </div>
      {#if franchiseResource.value.league}
        <div>
          <div class="text-[0.65rem] tracking-wider text-muted uppercase">League</div>
          <div class="text-foreground">{franchiseResource.value.league}</div>
        </div>
      {/if}
      {#if franchiseResource.value.location}
        <div>
          <div class="text-[0.65rem] tracking-wider text-muted uppercase">Location</div>
          <div class="text-foreground">{franchiseResource.value.location}</div>
        </div>
      {/if}
      {#if franchiseResource.value.active_from}
        <div>
          <div class="text-[0.65rem] tracking-wider text-muted uppercase">Active</div>
          <div class="text-foreground">
            {franchiseResource.value.active_from}–{franchiseResource.value.active_to ?? 'pres.'}
          </div>
        </div>
      {/if}
    </div>

    {#if franchiseEras.length > 0}
      <div class="mt-4 border-t border-outline pt-3">
        <div class="mb-2 font-mono text-[0.65rem] tracking-wider text-muted uppercase">Franchise era span</div>
        <div class="flex flex-wrap gap-1">
          {#each franchiseEras as era (era.code)}
            <EraRangeChip {era} />
          {/each}
        </div>
      </div>
    {/if}
  </div>

  {#if year}
    <div class="mt-4 rounded-lg border border-outline bg-crust p-4">
      <div class="mb-3 flex items-center gap-2">
        <div class="panel-label">{year} Season</div>
        {#if seasonEra}<EraBadge era={seasonEra} size="xs" />{/if}
      </div>

      {#if seasonResource.loading}
        <p class="font-mono text-[0.78rem] text-muted">Loading season data…</p>
      {:else if seasonResource.error}
        <p class="font-mono text-[0.78rem] text-warning">{seasonResource.error}</p>
      {:else if seasonResource.value}
        <div class="grid grid-cols-2 gap-x-6 gap-y-2 font-mono text-xs">
          {#if seasonResource.value.name}
            <div>
              <div class="text-[0.65rem] tracking-wider text-muted uppercase">Team name</div>
              <div class="text-foreground">{seasonResource.value.name}</div>
            </div>
          {/if}
          {#if seasonResource.value.wins != null || seasonResource.value.losses != null}
            <div>
              <div class="text-[0.65rem] tracking-wider text-muted uppercase">Record</div>
              <div class="text-foreground">{fmtRecord(seasonResource.value.wins, seasonResource.value.losses)}</div>
            </div>
          {/if}
          {#if seasonResource.value.rank != null}
            <div>
              <div class="text-[0.65rem] tracking-wider text-muted uppercase">Rank</div>
              <div class="text-foreground">{seasonResource.value.rank}</div>
            </div>
          {/if}
          {#if seasonResource.value.division}
            <div>
              <div class="text-[0.65rem] tracking-wider text-muted uppercase">Division</div>
              <div class="text-foreground">{seasonResource.value.division}</div>
            </div>
          {/if}
          {#if seasonResource.value.park}
            <div class="col-span-2">
              <div class="text-[0.65rem] tracking-wider text-muted uppercase">Park</div>
              <div class="text-foreground">{seasonResource.value.park}</div>
            </div>
          {/if}
        </div>
      {:else}
        <p class="font-mono text-[0.78rem] text-muted">No season data found for {year}.</p>
      {/if}
    </div>
  {:else}
    <div class="mt-4 rounded-lg border border-outline bg-crust p-4">
      <p class="font-mono text-[0.78rem] text-muted">Enter a season year in the sidebar to view team-season details.</p>
    </div>
  {/if}
{/if}
