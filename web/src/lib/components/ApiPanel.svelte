<script lang="ts">
  import CopyButton from './CopyButton.svelte';

  let { endpoint, url, sampleJson }: { endpoint: string; url: string; sampleJson?: string } = $props();

  let curlText = $derived(`curl -s \\\n  "${url}" \\\n  | jq .`);
  let jsonCollapsed = $state(false);
</script>

<div class="flex flex-col gap-4">
  <div>
    <div class="ap-label">Endpoint</div>
    <div class="rounded-md bg-surface px-3 py-2.5 font-monospace text-[0.72rem] leading-relaxed break-all text-primary">
      {endpoint}
    </div>
  </div>

  <div>
    <div class="ap-label">Generated URL</div>
    <div class="rounded-md bg-surface px-3 py-2.5 font-monospace text-[0.72rem] leading-relaxed break-all text-primary">
      {url}
    </div>
    <div class="mt-1.5 flex gap-1.5">
      <CopyButton text={url} label="copy URL" />
    </div>
  </div>

  <div>
    <div class="ap-label">curl</div>
    <pre
      class="rounded-md bg-surface px-3 py-3 font-monospace text-[0.68rem] leading-relaxed break-all whitespace-pre-wrap text-[#a5b4fc]">{curlText}</pre>
    <div class="mt-1.5">
      <CopyButton text={curlText} label="copy curl" />
    </div>
  </div>

  {#if sampleJson}
    <div>
      <button
        class="ap-label flex w-full cursor-pointer items-center gap-1 transition-colors hover:text-foreground"
        onclick={() => (jsonCollapsed = !jsonCollapsed)}>
        Sample response <span class="ml-auto">{jsonCollapsed ? '▶' : '▼'}</span>
      </button>
      {#if !jsonCollapsed}
        <pre
          class="max-h-60 overflow-y-auto rounded-md bg-surface px-3 py-3 font-monospace text-[0.68rem] leading-relaxed text-[#86efac]">{sampleJson}</pre>
      {/if}
    </div>
  {/if}
</div>

<style>
  /* TODO: this could be moved to the layout.css file as an @utility */
  .ap-label {
    font-family: var(--font-monospace);
    font-size: 0.65rem;
    color: var(--color-muted);
    letter-spacing: 0.08em;
    text-transform: uppercase;
    margin-bottom: 0.5rem;
    padding-bottom: 0.375rem;
    border-bottom: 1px solid var(--color-outline);
  }
</style>
