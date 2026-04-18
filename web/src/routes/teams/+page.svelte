<script lang="ts">
  import { ThreeColLayout } from '$lib';
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Teams</div>
  {/snippet}
  {#snippet center()}
    <p class="text-sm text-muted">Team & franchise explorer — coming soon</p>
  {/snippet}
  {#snippet panel()}
    <div class="panel-label">API</div>
  {/snippet}
</ThreeColLayout>
