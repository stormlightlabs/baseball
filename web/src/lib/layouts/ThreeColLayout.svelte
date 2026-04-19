<script lang="ts">
  import type { Snippet } from 'svelte';

  let { sidebar, center, panel }: { sidebar: Snippet; center: Snippet; panel: Snippet } = $props();
</script>

<div
  class="grid h-full max-h-full min-h-0 w-full overflow-hidden bg-mantle"
  style="grid-template-columns: 280px 1fr 320px">
  <aside class="h-full min-h-0 min-w-0 overflow-y-auto overscroll-y-contain border-r border-outline p-5">
    {@render sidebar()}
  </aside>
  <main class="h-full min-h-0 min-w-0 overflow-y-auto overscroll-y-contain p-6">
    {@render center()}
  </main>
  <aside class="h-full min-h-0 min-w-0 overflow-y-auto overscroll-y-contain border-l border-outline p-5">
    {@render panel()}
  </aside>
</div>
