<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import type { Pathname } from '$app/types';
  import { EP } from '$lib/endpoints';
  import { DEFAULT_PLAYER_TAB } from '$lib/players/constants';
  import { fetchList } from '$lib/players/fetchers';
  import { AsyncListResource } from '$lib/players/resources.svelte';
  import type { Relative } from '$lib/players/types';
  import { onMount } from 'svelte';

  let playerId = $derived(page.params.id ?? '');
  let q = $derived(page.url.searchParams.get('q') ?? '');

  const relativesResource = new AsyncListResource<Relative>();
  let lastKey = $state();

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

  function relativeHref(id: string): string {
    const encodedId = encodeURIComponent(id);
    const base = `/players/${encodedId}/${DEFAULT_PLAYER_TAB}`;
    if (!q) return base;
    const qs = new URLSearchParams({ q }).toString();
    return `${base}?${qs}`;
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
              href={resolve(relativeHref(rel.player_id) as Pathname)}
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
