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

<section class="mx-auto max-w-6xl px-4 pb-6 sm:px-6 lg:px-8">
  <div class="rounded-lg border border-outline bg-crust p-4">
    {@render children()}

    <div class="mt-3 border-t border-outline/70 pt-2.5">
      <div class="truncate text-muted">
        <span class="font-display text-[0.7rem] tracking-[0.08em] uppercase">ENDPOINT:</span>
        <span class="ml-1 font-mono text-[0.63rem]">{endpoint}</span>
      </div>
      <div class="mt-1.5 flex flex-wrap gap-1.5">
        <CopyButton text={copyUrl} label="copy URL" />
        <CopyButton text={curlText} label="copy curl" />
      </div>
    </div>
  </div>
</section>
