<script lang="ts">
  import CopyButton from './CopyButton.svelte';

  let { method = 'GET', url }: { method?: string; url: string } = $props();

  let curlText = $derived(`curl -s "${url}"`);
</script>

<div class="mt-px flex items-center gap-6 rounded-lg border border-outline bg-crust px-5 py-4">
  <span class="font-mono text-[0.72rem] text-muted">{method}</span>
  <span class="min-w-0 flex-1 overflow-x-auto font-mono text-xs whitespace-nowrap text-primary">
    {url}
  </span>
  <CopyButton text={url} label="copy URL" />
  <CopyButton text={curlText} label="copy curl" />
</div>
