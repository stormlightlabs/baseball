<script lang="ts">
  import type { Snippet } from 'svelte';

  let { text, children }: { text: string; children: Snippet } = $props();
</script>

<span class="group/tooltip relative inline-flex items-center">
  {@render children()}
  <span
    role="tooltip"
    class="pointer-events-none absolute bottom-full left-1/2 z-20 mb-1 w-max max-w-56 -translate-x-1/2 rounded border border-outline bg-mantle px-2 py-1 text-center font-mono text-[0.62rem] leading-tight text-muted opacity-0 shadow-md transition-opacity duration-150 group-focus-within/tooltip:opacity-100 group-hover/tooltip:opacity-100">
    {text}
  </span>
</span>
