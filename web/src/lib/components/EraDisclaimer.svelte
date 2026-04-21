<script lang="ts">
  import type { Era } from '$lib/eras';
  import { meta } from '$lib/meta.svelte';

  type Props = {
    /** Specific eras that triggered this disclaimer. When empty, shows a generic historical-data warning. */
    eras?: Era[];
    message?: string;
  };

  let { eras = [], message }: Props = $props();

  const hasCaveats = $derived(eras.some((e) => e.caveat));

  function eraLabel(era: Era): string {
    return meta.data?.era_labels?.[era.code] ?? era.label;
  }

  const eraLabels = $derived(eras.map((e) => eraLabel(e)).join(', '));
</script>

<div class="flex items-start gap-2.5 rounded-md border border-warning/30 bg-warning/8 px-3 py-2.5">
  <span class="mt-px shrink-0 text-[0.85rem] text-warning">⚠</span>
  <div class="min-w-0">
    {#if message}
      <p class="text-[0.78rem] text-warning/90">{message}</p>
    {:else if eras.length > 0 && hasCaveats}
      <p class="text-[0.78rem] text-warning/90">
        <span class="font-medium">{eraLabels}:</span> play-event density and data coverage vary for this era. Cross-era comparisons
        may not be directly meaningful.
      </p>
    {:else}
      <p class="text-[0.78rem] text-warning/90">
        Historical data before 1950 has variable play-event coverage. Federal League (1914–15) and Negro Leagues data
        should be interpreted with source caveats in mind.
      </p>
    {/if}
    {#if eras.some((e) => e.caveat)}
      <ul class="mt-1 space-y-0.5">
        {#each eras.filter((e) => e.caveat) as era (era.code)}
          <li class="font-mono text-xxs text-warning/70">
            <span class="font-medium">{eraLabel(era)}</span>: {era.caveat}
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>
