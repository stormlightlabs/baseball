<script lang="ts">
  type Props = {
    value?: string;
    options?: number[];
    min?: number;
    max?: number;
    loading?: boolean;
    error?: string;
    disabled?: boolean;
    placeholder?: string;
    loadingMessage?: string;
    emptyMessage?: string;
    dialogLabel?: string;
    oncommit?: (value: string) => void;
  };

  let {
    value = $bindable(''),
    options = [],
    min,
    max,
    loading = false,
    error,
    disabled = false,
    placeholder = 'Select value',
    loadingMessage = 'Loading options…',
    emptyMessage = 'No options available.',
    dialogLabel = 'Range picker',
    oncommit
  }: Props = $props();

  let open = $state(false);

  let displayValue = $derived.by(() => {
    const trimmed = value.trim();
    if (trimmed) return trimmed;
    return placeholder;
  });

  let rangeMin = $derived.by(() => min ?? options[0] ?? null);
  let rangeMax = $derived.by(() => max ?? options.at(-1) ?? null);
  let rangeLabel = $derived.by(() => {
    if (rangeMin == null || rangeMax == null) return null;
    return `${rangeMin}–${rangeMax}`;
  });

  function parseNumericValue(raw: string): number | null {
    const parsed = Number.parseInt(raw, 10);
    if (Number.isNaN(parsed)) return null;
    return parsed;
  }

  function sliderIndexForValue(raw: string): number {
    if (options.length === 0) return 0;

    const parsed = parseNumericValue(raw);
    if (parsed == null) return options.length - 1;

    const exactIndex = options.indexOf(parsed);
    if (exactIndex !== -1) return exactIndex;

    let nearestIndex = 0;
    let nearestDistance = Number.POSITIVE_INFINITY;
    for (const [idx, optionValue] of options.entries()) {
      const distance = Math.abs(optionValue - parsed);
      if (distance < nearestDistance) {
        nearestDistance = distance;
        nearestIndex = idx;
      }
    }
    return nearestIndex;
  }

  function valueForSliderIndex(rawIndex: number): string {
    if (options.length === 0) return '';
    const boundedIndex = Math.max(0, Math.min(options.length - 1, rawIndex));
    return String(options[boundedIndex] ?? options.at(-1));
  }

  function applyFromSlider(rawIndex: number, shouldCommit: boolean): void {
    const nextValue = valueForSliderIndex(rawIndex);
    if (!nextValue) return;
    value = nextValue;
    if (shouldCommit) {
      open = false;
      oncommit?.(nextValue);
    }
  }

  function toggleOpen(): void {
    if (disabled) return;
    open = !open;
  }

  function handleButtonKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      open = false;
    }
  }

  function handlePopoverKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      open = false;
    }
  }

  function handleOutside(node: HTMLElement) {
    function onClick(event: MouseEvent): void {
      if (!node.contains(event.target as Node)) {
        open = false;
      }
    }

    document.addEventListener('click', onClick, true);
    return {
      destroy() {
        document.removeEventListener('click', onClick, true);
      }
    };
  }
</script>

<div class="relative" use:handleOutside>
  <button
    type="button"
    aria-haspopup="dialog"
    aria-expanded={open}
    onclick={toggleOpen}
    onkeydown={handleButtonKeydown}
    {disabled}
    class="w-full cursor-pointer rounded border border-outline bg-mantle px-2 py-1.5 text-left font-mono text-xs text-foreground focus:ring-1 focus:ring-primary/50 focus:outline-none disabled:cursor-not-allowed disabled:opacity-60">
    {displayValue}
  </button>

  {#if open}
    <div
      role="dialog"
      aria-label={dialogLabel}
      tabindex="-1"
      onkeydown={handlePopoverKeydown}
      class="absolute z-20 mt-1.5 w-full rounded-lg border border-outline bg-surface p-3 shadow-lg">
      {#if loading}
        <p class="font-mono text-[0.68rem] text-muted">{loadingMessage}</p>
      {:else if error}
        <p class="font-mono text-[0.68rem] text-warning">{error}</p>
      {:else if options.length === 0}
        <p class="font-mono text-[0.68rem] text-muted">{emptyMessage}</p>
      {:else}
        {#if rangeLabel}
          <div class="mb-2 flex items-center justify-between font-mono text-[0.62rem] text-muted">
            <span>{rangeMin}</span>
            <span>{rangeLabel}</span>
            <span>{rangeMax}</span>
          </div>
        {/if}

        <input
          type="range"
          min="0"
          max={Math.max(0, options.length - 1)}
          step="1"
          value={sliderIndexForValue(value)}
          oninput={(event) => {
            const nextIndex = Number((event.target as HTMLInputElement).value);
            applyFromSlider(nextIndex, false);
          }}
          onchange={(event) => {
            const nextIndex = Number((event.target as HTMLInputElement).value);
            applyFromSlider(nextIndex, true);
          }}
          class="h-2 w-full cursor-pointer appearance-none rounded-lg bg-mantle accent-primary" />

        <div class="mt-2 text-center font-display text-[0.88rem] text-foreground">{value}</div>
      {/if}
    </div>
  {/if}
</div>
