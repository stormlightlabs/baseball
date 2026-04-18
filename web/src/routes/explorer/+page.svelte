<script lang="ts">
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Endpoints</div>
  {/snippet}
  {#snippet center()}
    <p class="text-sm text-muted">API explorer — coming soon</p>
  {/snippet}
  {#snippet panel()}
    <div class="panel-label">Schema</div>
  {/snippet}
</ThreeColLayout>
