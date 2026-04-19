<script lang="ts">
  import CopyButton from './CopyButton.svelte';

  let { endpoint, url, sampleJson }: { endpoint: string; url: string; sampleJson?: string } = $props();

  let curlText = $derived(`curl -s "${url}"`);
  let jsonCollapsed = $state(false);
</script>

<div class="flex flex-col gap-4">
  <div>
    <div class="panel-label">Endpoint</div>
    <div class="rounded-md bg-surface px-3 py-2.5">
      <div class="overflow-x-auto">
        <div class="font-mono text-[0.72rem] leading-relaxed whitespace-nowrap text-primary">{endpoint}</div>
      </div>
    </div>
  </div>

  <div>
    <div class="panel-label">Generated URL</div>
    <div class="rounded-md bg-surface px-3 py-2.5">
      <div class="overflow-x-auto">
        <div class="font-mono text-[0.72rem] leading-relaxed whitespace-nowrap text-primary">{url}</div>
      </div>
    </div>
    <div class="mt-1.5 flex gap-1.5">
      <CopyButton text={url} label="copy URL" />
    </div>
  </div>

  <div>
    <div class="panel-label">curl</div>
    <div class="rounded-md bg-surface px-3 py-3">
      <pre class="m-0 overflow-x-auto font-mono text-[0.68rem] leading-relaxed whitespace-nowrap text-[#a5b4fc]"><code
          >{curlText}</code></pre>
    </div>
    <div class="mt-1.5">
      <CopyButton text={curlText} label="copy curl" />
    </div>
  </div>

  {#if sampleJson}
    <div>
      <button
        class="panel-label flex w-full cursor-pointer items-center gap-1 transition-colors hover:text-foreground"
        onclick={() => (jsonCollapsed = !jsonCollapsed)}>
        Sample response <span class="ml-auto">{jsonCollapsed ? '▶' : '▼'}</span>
      </button>
      {#if !jsonCollapsed}
        <pre
          class="max-h-60 overflow-y-auto rounded-md bg-surface px-3 py-3 font-mono text-[0.68rem] leading-relaxed text-[#86efac]">{sampleJson}</pre>
      {/if}
    </div>
  {/if}
</div>
