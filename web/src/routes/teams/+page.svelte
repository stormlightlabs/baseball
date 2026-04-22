<script lang="ts">
  import { resolve } from '$app/paths';
  import EraRangeChip from '$lib/components/EraRangeChip.svelte';
  import StandingsPanel from '$lib/components/StandingsPanel.svelte';
  import { STATIC_ERAS } from '$lib/eras';
</script>

<div class="flex h-full flex-col gap-4">
  <div class="flex justify-end">
    <a
      href={resolve('/standings')}
      class="inline-flex items-center gap-1 rounded border border-outline px-2.5 py-1 font-mono text-[0.68rem] text-muted no-underline transition-colors hover:bg-surface hover:text-foreground">
      <span class="flex items-center">
        <i class="i-tabler-table"></i>
      </span>
      Full standings
    </a>
  </div>

  <StandingsPanel />

  <div
    class="flex flex-col items-center justify-center gap-3 rounded-lg border border-outline bg-crust p-4 text-center">
    <div class="font-display text-[1.1rem] text-muted">Team & Franchise Explorer</div>
    <p class="max-w-xs font-mono text-[0.78rem] text-muted/60">
      Search for a team or select a franchise in the sidebar.
    </p>
    <div class="mt-2 flex flex-wrap justify-center gap-1.5">
      {#each STATIC_ERAS as era (era.code)}
        <EraRangeChip {era} />
      {/each}
    </div>
  </div>
</div>
