<script lang="ts">
  import { ThreeColLayout } from '$lib';
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Seasons</div>
  {/snippet}
  {#snippet center()}
    <p class="text-muted text-sm">Season hub — coming soon</p>
  {/snippet}
  {#snippet panel()}
    <div class="panel-label">API</div>
  {/snippet}
</ThreeColLayout>
