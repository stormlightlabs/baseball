<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import EraDisclaimer from '$lib/components/EraDisclaimer.svelte';
  import EraRangeChip from '$lib/components/EraRangeChip.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear, erasInRange, type Era } from '$lib/eras';
  import { AsyncValueResource } from '$lib/players/resources.svelte';
  import { normalizeFranchiseProfile, normalizeTeamSeasonProfile } from '$lib/teams/normalizers';
  import type { FranchiseProfile, TeamSeasonProfile } from '$lib/teams/types';
  import { onMount } from 'svelte';

  let teamId = $derived(page.params.id ?? '');
  let franchiseId = $derived(page.url.searchParams.get('franchise_id') ?? '');
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

  let comparisonGapMessage = $derived.by(() => {
    if (franchiseEras.length < 2 || !franchiseEras.some((era) => era.caveat)) return null;
    const hasNegroLeagues = franchiseEras.some((era) => era.code === 'nlg');
    const hasModern = franchiseEras.some((era) => era.code === 'modern');
    if (hasNegroLeagues && hasModern) {
      return 'This franchise spans Negro Leagues and modern eras. Cross-era comparisons should be treated as directional.';
    }
    return 'This franchise spans eras with uneven source coverage. Use cross-era comparisons with caution.';
  });

  type EraTimelineSegment = { era: Era; from: number; to: number; leftPct: number; widthPct: number };

  let franchiseTimeline = $derived.by((): EraTimelineSegment[] => {
    const profile = franchiseResource.value;
    if (!profile?.active_from) return [];
    const rangeFrom = profile.active_from;
    const rangeTo = profile.active_to ?? new Date().getFullYear();
    const span = Math.max(1, rangeTo - rangeFrom + 1);

    return franchiseEras
      .map((era) => {
        const segmentFrom = Math.max(rangeFrom, era.from);
        const segmentTo = Math.min(rangeTo, era.to);
        if (segmentFrom > segmentTo) return null;

        return {
          era,
          from: segmentFrom,
          to: segmentTo,
          leftPct: ((segmentFrom - rangeFrom) / span) * 100,
          widthPct: ((segmentTo - segmentFrom + 1) / span) * 100
        };
      })
      .filter((segment): segment is EraTimelineSegment => segment != null);
  });

  async function refreshFranchise(force = false): Promise<void> {
    const selectedTeamId = teamId;
    if (!selectedTeamId) {
      franchiseResource.clear();
      return;
    }
    const selectedFranchiseId = franchiseId || selectedTeamId;
    const key = `${selectedTeamId}|${selectedFranchiseId}`;
    if (!force && key === lastFranchiseKey) return;
    lastFranchiseKey = key;

    await franchiseResource.load(async () => {
      try {
        const payload = await apiFetch<Record<string, unknown>>(EP.franchise(selectedFranchiseId));
        return normalizeFranchiseProfile(payload);
      } catch {
        const payload = await apiFetch<Record<string, unknown>>(EP.team(selectedTeamId));
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

  function timelineSegmentClass(color: Era['color']): string {
    switch (color) {
      case 'warning': {
        return 'bg-warning/25 border-warning/35';
      }
      case 'secondary': {
        return 'bg-secondary/25 border-secondary/35';
      }
      case 'primary': {
        return 'bg-primary/25 border-primary/35';
      }
      default: {
        return 'bg-surface border-outline';
      }
    }
  }
</script>

{#if franchiseResource.loading}
  <p class="mt-4 font-mono text-xs text-muted">Loading…</p>
{:else if franchiseResource.error}
  <p class="mt-4 font-mono text-xs text-warning">{franchiseResource.error}</p>
{:else if franchiseResource.value}
  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">Franchise</div>
    <div class="grid grid-cols-2 gap-x-6 gap-y-2 font-mono text-xs">
      <div>
        <div class="kv-list--key">Name</div>
        <div class="kv-list--value">{franchiseResource.value.name}</div>
      </div>
      <div>
        <div class="kv-list--key">Franchise ID</div>
        <div class="kv-list--value">{franchiseResource.value.id}</div>
      </div>
      {#if franchiseResource.value.league}
        <div>
          <div class="kv-list--key">League</div>
          <div class="kv-list--value">{franchiseResource.value.league}</div>
        </div>
      {/if}
      {#if franchiseResource.value.location}
        <div>
          <div class="kv-list--key">Location</div>
          <div class="kv-list--value">{franchiseResource.value.location}</div>
        </div>
      {/if}
      {#if franchiseResource.value.active_from}
        <div>
          <div class="kv-list--key">Active</div>
          <div class="kv-list--value">
            {franchiseResource.value.active_from}–{franchiseResource.value.active_to ?? 'pres.'}
          </div>
        </div>
      {/if}
    </div>

    {#if franchiseEras.length > 0}
      <div class="mt-4 border-t border-outline pt-3">
        <div class="kv-list--key mb-2 font-mono">Franchise era span</div>
        <div class="flex flex-wrap gap-1">
          {#each franchiseEras as era (era.code)}
            <EraRangeChip {era} />
          {/each}
        </div>

        {#if franchiseTimeline.length > 0}
          <div class="mt-3 rounded border border-outline/70 bg-surface px-2.5 py-2">
            <div class="mb-2 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Continuity timeline</div>
            <div class="relative h-5 rounded border border-outline bg-mantle/70">
              {#each franchiseTimeline as segment (segment.era.code)}
                <div
                  class="absolute top-0.5 bottom-0.5 rounded border {timelineSegmentClass(segment.era.color)}"
                  style="left: {segment.leftPct}%; width: {Math.max(segment.widthPct, 1.6)}%"
                  title={`${segment.era.label} (${segment.from}–${segment.to})`}>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}

    {#if comparisonGapMessage}
      <div class="mt-4">
        <EraDisclaimer eras={franchiseEras} message={comparisonGapMessage} />
      </div>
    {/if}
  </div>

  {#if year}
    <div class="mt-4 rounded-lg border border-outline bg-crust p-4">
      <div class="panel-label">
        <div class="flex items-center justify-between">
          <span>
            {year} season
          </span>
          <span>
            Era: {#if seasonEra}<EraBadge era={seasonEra} size="xs" />{/if}
          </span>
        </div>
      </div>

      {#if seasonResource.loading}
        <p class="font-mono text-xs text-muted">Loading season data…</p>
      {:else if seasonResource.error}
        <p class="font-mono text-xs text-warning">{seasonResource.error}</p>
      {:else if seasonResource.value}
        <div class="grid grid-cols-2 gap-x-6 gap-y-2 text-xs">
          <div>
            <div class="kv-list--key">Team ID</div>
            <div class="kv-list--value">{seasonResource.value.id}</div>
          </div>
          {#if seasonResource.value.franchise_id}
            <div>
              <div class="kv-list--key">Franchise ID</div>
              <div class="kv-list--value">{seasonResource.value.franchise_id}</div>
            </div>
          {/if}
          {#if seasonResource.value.name}
            <div>
              <div class="kv-list--key">Team name</div>
              <div class="kv-list--value">{seasonResource.value.name}</div>
            </div>
          {/if}
          {#if seasonResource.value.wins != null || seasonResource.value.losses != null}
            <div>
              <div class="kv-list--key">Record</div>
              <div class="kv-list--value">{fmtRecord(seasonResource.value.wins, seasonResource.value.losses)}</div>
            </div>
          {/if}
          {#if seasonResource.value.rank != null}
            <div>
              <div class="kv-list--key">Rank</div>
              <div class="kv-list--value">{seasonResource.value.rank}</div>
            </div>
          {/if}
          {#if seasonResource.value.division}
            <div>
              <div class="kv-list--key">Division</div>
              <div class="kv-list--value">{seasonResource.value.division}</div>
            </div>
          {/if}
          {#if seasonResource.value.park}
            <div class="col-span-2">
              <div class="kv-list--key">Park</div>
              <div class="kv-list--value">{seasonResource.value.park}</div>
            </div>
          {/if}
        </div>
      {:else}
        <p class="font-mono text-xs text-muted">No season data found for {year}.</p>
      {/if}
    </div>
  {:else}
    <div class="mt-4 rounded-lg border border-outline bg-crust p-4">
      <p class="font-mono text-xs text-muted">Enter a year in the sidebar to view team-season details.</p>
    </div>
  {/if}
{/if}

<style lang="postcss">
  @reference '$tailwind';

  .kv-list--key {
    @apply font-mono text-xxs tracking-wider text-muted uppercase;
  }

  .kv-list--value {
    @apply font-display text-sm text-foreground;
  }
</style>
