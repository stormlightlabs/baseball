<script lang="ts">
  import type { Snippet } from 'svelte';
  import CopyButton from '$lib/components/CopyButton.svelte';

  let { endpoint, children }: { endpoint: string; children: Snippet } = $props();

  let copyUrl = $derived.by(() => {
    const origin = globalThis.window?.location.origin;

    if (!origin || endpoint.startsWith('http://') || endpoint.startsWith('https://')) {
      return endpoint;
    }

    const normalizedEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
    return `${origin}${normalizedEndpoint}`;
  });
  let curlText = $derived(`curl -s "${copyUrl}"`);
</script>

<div class="rounded-md border border-outline/60 bg-crust/40">
  {@render children()}

  <div class="border-t border-outline/60 px-2 py-1.5">
    <div class="truncate font-mono text-xxs text-muted">{endpoint}</div>
    <div class="mt-1.5 flex flex-wrap gap-1.5">
      <CopyButton text={copyUrl} label="copy URL" />
      <CopyButton text={curlText} label="copy curl" />
    </div>
  </div>
</div>
