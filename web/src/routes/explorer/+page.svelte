<script lang="ts">
  import { ThreeColLayout } from '$lib';
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Endpoints</div>
  {/snippet}
  {#snippet center()}
    <p class="text-muted text-sm">API explorer — coming soon</p>
  {/snippet}
  {#snippet panel()}
    <div class="panel-label">Schema</div>
  {/snippet}
</ThreeColLayout>
