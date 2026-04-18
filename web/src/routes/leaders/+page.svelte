<script lang="ts">
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Leaders</div>
  {/snippet}
  {#snippet center()}
    <p class="text-sm text-muted">Stat leaders — coming soon</p>
  {/snippet}
  {#snippet panel()}
    <div class="panel-label">API</div>
  {/snippet}
</ThreeColLayout>
