<script lang="ts">
  import { page } from '$app/state';
  import AdvancedTabPanel from '$lib/components/AdvancedTabPanel.svelte';
  import { EP } from '$lib/endpoints';

  let playerId = $derived(page.params.id ?? '');
  let endpoint = $derived(EP.playerSplits(playerId));
</script>

<AdvancedTabPanel {endpoint} label="Splits" />
