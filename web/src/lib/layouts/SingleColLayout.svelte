<script lang="ts">
  import type { Snippet } from 'svelte';

  let { children }: { children: Snippet } = $props();
</script>

<div class="mx-auto max-w-5xl bg-mantle px-8 py-12">
  {@render children()}
</div>
