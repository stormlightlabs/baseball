<script lang="ts">
  import PromoBanner from '$lib/components/PromoBanner.svelte';
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Compare</div>
  {/snippet}
  {#snippet center()}
    <p class="text-sm text-muted">Compare mode — coming soon</p>
  {/snippet}
  {#snippet panel()}
    <div class="panel-label">API</div>
    <div class="mt-3">
      <PromoBanner variant="vertical" />
    </div>
  {/snippet}
</ThreeColLayout>
