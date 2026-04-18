<script lang="ts">
  let {
    label,
    range,
    percent,
    variant = 'primary',
    href,
    tooltip
  }: {
    label: string;
    range: string;
    percent: number;
    variant?: 'primary' | 'secondary' | 'warning';
    href?: string;
    tooltip?: string;
  } = $props();

  const cls = $derived.by(() => {
    switch (variant) {
      case 'primary':
        return 'bg-primary';
      case 'secondary':
        return 'bg-secondary';
      case 'warning':
        return 'bg-warning';
    }
  });
</script>

<div class="mb-2">
  <div class="mb-1 flex justify-between text-[0.72rem] text-muted">
    <span class="inline-flex items-center gap-1.5">
      {#if href}
        <a {href} target="_blank" rel="external" class="text-primary hover:underline">{label}</a>
      {:else}
        <span>{label}</span>
      {/if}
      {#if tooltip}
        <span class="cursor-help font-mono text-[0.62rem] text-muted/80" title={tooltip} aria-label={tooltip}>[?]</span>
      {/if}
    </span>
    <span>{range}</span>
  </div>
  <div class="h-1.5 w-full overflow-hidden rounded-full bg-outline">
    <div
      class="h-full rounded-full {cls} transition-all duration-500"
      style="width: {Math.min(100, Math.max(0, percent))}%">
    </div>
  </div>
</div>
