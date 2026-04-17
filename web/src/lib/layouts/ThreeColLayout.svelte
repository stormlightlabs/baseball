<script lang="ts">
  import type { Snippet } from 'svelte';

  let { sidebar, center, panel }: { sidebar: Snippet; center: Snippet; panel: Snippet } = $props();
</script>

<div class="grid h-[calc(100vh-3.5rem)]" style="grid-template-columns: 280px 1fr 320px">
  <aside class="overflow-y-auto border-r border-outline p-5">
    {@render sidebar()}
  </aside>
  <main class="overflow-y-auto p-6">
    {@render center()}
  </main>
  <aside class="overflow-y-auto border-l border-outline p-5">
    {@render panel()}
  </aside>
</div>
