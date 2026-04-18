<script lang="ts">
  import { resolve } from '$app/paths';
  import type { Era } from '$lib/eras';
  import EraBadge from './EraBadge.svelte';

  type Props = {
    era: Era;
    /** When provided, renders as a link to /seasons?year=... */
    year?: number;
    onclick?: () => void;
  };

  let { era, year, onclick }: Props = $props();
</script>

{#if year}
  <a
    href={resolve(`/seasons?year=${year}`)}
    class="inline-flex items-center gap-1.5 rounded-full border border-outline bg-crust px-2.5 py-1 no-underline transition-colors hover:border-primary/50 hover:bg-surface"
    title={era.caveat}>
    <EraBadge {era} size="xs" />
    <span class="font-monospace text-[0.7rem] text-muted">{era.from}–{era.to}</span>
  </a>
{:else}
  <button
    {onclick}
    class="inline-flex cursor-pointer items-center gap-1.5 rounded-full border border-outline bg-crust px-2.5 py-1 transition-colors hover:border-primary/50 hover:bg-surface"
    title={era.caveat}>
    <EraBadge {era} size="xs" />
    <span class="font-monospace text-[0.7rem] text-muted">{era.from}–{era.to}</span>
  </button>
{/if}
