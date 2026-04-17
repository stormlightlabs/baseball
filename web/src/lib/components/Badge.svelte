<script lang="ts">
  let { label, variant = 'default' }: { label: string; variant?: 'default' | 'primary' | 'secondary' | 'warning' } =
    $props();

  const cls = $derived.by(() => {
    switch (variant) {
      case 'primary':
        return 'bg-primary/20 text-primary';
      case 'secondary':
        return 'bg-secondary/20 text-secondary';
      case 'warning':
        return 'bg-warning/20 text-warning';
      default:
        return 'bg-outline text-muted';
    }
  });
</script>

<span class="inline-block rounded px-1.5 py-0.5 font-monospace text-[0.65rem] {cls}">
  {label}
</span>
