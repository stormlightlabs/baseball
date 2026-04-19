<script lang="ts">
  import type { Snippet } from 'svelte';

  let { sidebar, center, panel }: { sidebar: Snippet; center: Snippet; panel: Snippet } = $props();
</script>

<div class="grid h-full min-h-full bg-mantle" style="grid-template-columns: 280px 1fr 320px">
  <aside class="min-w-0 overflow-y-auto border-r border-outline p-5">
    {@render sidebar()}
  </aside>
  <main class="min-w-0 overflow-y-auto p-6">
    {@render center()}
  </main>
  <aside class="min-w-0 overflow-y-auto border-l border-outline p-5">
    {@render panel()}
  </aside>
</div>
