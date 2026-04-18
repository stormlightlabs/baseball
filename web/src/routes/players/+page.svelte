<script lang="ts">
  import EraRangeChip from '$lib/components/EraRangeChip.svelte';
  import { STATIC_ERAS } from '$lib/eras';
</script>

<div class="flex h-full flex-col items-center justify-center gap-3 text-center">
  <div class="font-display text-[1.1rem] text-muted">Player Explorer</div>
  <p class="max-w-xs font-mono text-[0.78rem] text-muted/60">
    Search for a player in the sidebar to explore career stats, awards, game logs, and more.
  </p>
  <div class="mt-2 flex flex-wrap justify-center gap-1.5">
    {#each STATIC_ERAS as era (era.code)}
      <EraRangeChip {era} />
    {/each}
  </div>
</div>
