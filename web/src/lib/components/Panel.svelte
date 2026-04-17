<script lang="ts">
  import type { Snippet } from 'svelte';
  let { label, children }: { label?: string; children: Snippet } = $props();
</script>

<div class="rounded-lg border border-outline bg-crust p-5">
  {#if label}
    <div class="mb-3 border-b border-outline pb-2 font-monospace text-[0.65rem] tracking-widest text-muted uppercase">
      {label}
    </div>
  {/if}
  {@render children()}
</div>
