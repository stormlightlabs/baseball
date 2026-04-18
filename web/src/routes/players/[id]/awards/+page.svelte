<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear } from '$lib/eras';
  import { fetchList } from '$lib/players/fetchers';
  import { AsyncListResource } from '$lib/players/resources.svelte';
  import type { Award } from '$lib/players/types';
  import { onMount } from 'svelte';

  let playerId = $derived(page.params.id ?? '');

  const awardsResource = new AsyncListResource<Award>();
  let lastKey = $state();

  async function refresh(force = false): Promise<void> {
    const id = playerId;
    if (!force && id === lastKey) return;
    lastKey = id;

    await awardsResource.load(() => fetchList<Award>(EP.playerAwards(id)));
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-3">Awards timeline</div>
  {#if awardsResource.loading}
    <p class="font-mono text-xs text-muted">Loading…</p>
  {:else if awardsResource.items.length === 0}
    <p class="font-mono text-[0.78rem] text-muted">No awards on record.</p>
  {:else}
    <div class="flex flex-col gap-2">
      {#each awardsResource.items as award, idx (`${award.year}-${award.name ?? award.award_id}-${award.league ?? ''}-${award.notes ?? ''}-${idx}`)}
        {@const era = eraForYear(award.year)}
        <div class="flex items-center gap-3 rounded-md bg-surface px-3 py-2.5">
          <a
            href={resolve(`/seasons?year=${award.year}`)}
            class="min-w-11 font-mono text-xs text-muted hover:text-primary">
            {award.year}
          </a>
          {#if era}<EraBadge {era} size="xs" />{/if}
          <span class="font-display text-[0.82rem] text-foreground">{award.name ?? award.award_id}</span>
          {#if award.league}
            <span class="ml-auto font-mono text-[0.68rem] text-muted">{award.league}</span>
          {/if}
          {#if award.notes}
            <span class="font-mono text-[0.68rem] text-muted">{award.notes}</span>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>
