<script lang="ts">
  let {
    label,
    range,
    percent,
    variant = 'primary'
  }: { label: string; range: string; percent: number; variant?: 'primary' | 'secondary' | 'warning' } = $props();

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
    <span>{label}</span>
    <span>{range}</span>
  </div>
  <div class="h-1.5 w-full overflow-hidden rounded-full bg-outline">
    <div
      class="h-full rounded-full {cls} transition-all duration-500"
      style="width: {Math.min(100, Math.max(0, percent))}%">
    </div>
  </div>
</div>
