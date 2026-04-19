<script lang="ts">
  import { tick } from 'svelte';
  import type { Snippet } from 'svelte';

  let { text, children }: { text: string; children: Snippet } = $props();

  let anchorEl = $state<HTMLElement | null>(null);
  let tooltipEl = $state<HTMLElement | null>(null);
  let visible = $state(false);
  let leftPx = $state(0);
  let topPx = $state(0);

  function hideTooltip() {
    visible = false;
  }

  async function showTooltip() {
    visible = true;
    await tick();
    updatePosition();
  }

  function handleFocusOut(event: FocusEvent) {
    const next = event.relatedTarget;
    if (next instanceof Node && anchorEl?.contains(next)) return;
    hideTooltip();
  }

  function handleViewportChange() {
    if (!visible) return;
    updatePosition();
  }

  function updatePosition() {
    if (!visible || !anchorEl || !tooltipEl) return;

    const anchor = anchorEl.getBoundingClientRect();
    const tip = tooltipEl.getBoundingClientRect();
    const margin = 8;

    let nextLeft = anchor.left + anchor.width / 2 - tip.width / 2;
    nextLeft = Math.max(margin, Math.min(nextLeft, window.innerWidth - tip.width - margin));

    let nextTop = anchor.top - tip.height - margin;
    if (nextTop < margin) {
      nextTop = anchor.bottom + margin;
    }

    leftPx = Math.round(nextLeft);
    topPx = Math.round(nextTop);
  }
</script>

<svelte:window onresize={handleViewportChange} onscroll={handleViewportChange} />

<span
  bind:this={anchorEl}
  role="group"
  class="inline-flex items-center"
  title={text}
  onmouseenter={() => void showTooltip()}
  onmouseleave={hideTooltip}
  onfocusin={() => void showTooltip()}
  onfocusout={handleFocusOut}>
  {@render children()}
</span>

{#if visible}
  <span
    bind:this={tooltipEl}
    role="tooltip"
    class="pointer-events-none fixed z-[120] w-max max-w-56 rounded border border-outline bg-mantle px-2 py-1 text-center font-mono text-[0.62rem] leading-tight text-muted shadow-md"
    style:left={`${leftPx}px`}
    style:top={`${topPx}px`}>
    {text}
  </span>
{/if}
