<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { EP } from '$lib/endpoints';
  import { fetchList } from '$lib/players/fetchers';
  import { AsyncListResource } from '$lib/players/resources.svelte';
  import type { Relative } from '$lib/players/types';
  import { onMount } from 'svelte';

  let playerId = $derived(page.params.id ?? '');
  let q = $derived(page.url.searchParams.get('q') ?? '');

  const relativesResource = new AsyncListResource<Relative>();
  let lastKey = '';

  async function refresh(force = false): Promise<void> {
    const id = playerId;
    if (!force && id === lastKey) return;
    lastKey = id;

    await relativesResource.load(() => fetchList<Relative>(EP.playerRelatives(id)));
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });

  function searchQuerySuffix(): string {
    if (!q) return '';
    return `?${new URLSearchParams({ q }).toString()}`;
  }
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-3">Relatives</div>
  {#if relativesResource.loading}
    <p class="font-mono text-xs text-muted">Loading…</p>
  {:else if relativesResource.items.length === 0}
    <p class="font-mono text-[0.78rem] text-muted">No relatives on record.</p>
  {:else}
    <div class="flex flex-col gap-2">
      {#each relativesResource.items as rel (rel.player_id ?? rel.name)}
        <div class="flex items-center gap-3 rounded-md bg-surface px-3 py-2.5">
          {#if rel.player_id}
            <a
              href={resolve(
                `/players/${encodeURIComponent(rel.player_id)}/batting${searchQuerySuffix()}` as `/players/${string}/batting`
              )}
              class="font-display text-[0.82rem] text-primary hover:underline">
              {rel.name ?? rel.player_id}
            </a>
          {:else}
            <span class="font-display text-[0.82rem] text-foreground">{rel.name ?? '?'}</span>
          {/if}
          {#if rel.relationship}
            <span class="ml-auto font-mono text-[0.68rem] text-muted">{rel.relationship}</span>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>
