<script lang="ts">
  import { ThreeColLayout } from '$lib';
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Games</div>
  {/snippet}
  {#snippet center()}
    <p class="text-sm text-muted">Game finder — coming soon</p>
  {/snippet}
  {#snippet panel()}
    <div class="panel-label">API</div>
  {/snippet}
</ThreeColLayout>
