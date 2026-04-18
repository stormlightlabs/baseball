<script lang="ts">
  import type { Era } from '$lib/eras';

  let { era, size = 'sm' }: { era: Era; size?: 'xs' | 'sm' } = $props();

  const colorCls = $derived.by(() => {
    switch (era.color) {
      case 'primary': {
        return 'bg-primary/15 text-primary border-primary/25';
      }
      case 'secondary': {
        return 'bg-secondary/15 text-secondary border-secondary/25';
      }
      case 'warning': {
        return 'bg-warning/15 text-warning border-warning/25';
      }
      default: {
        return 'bg-outline text-muted border-outline';
      }
    }
  });

  const sizeCls = $derived(size === 'xs' ? 'text-[0.6rem] px-1 py-px' : 'text-[0.65rem] px-1.5 py-0.5');
</script>

<span
  class="inline-block rounded border font-mono leading-none font-medium {colorCls} {sizeCls}"
  title="{era.label} ({era.from}–{era.to}){era.caveat ? ' · ' + era.caveat : ''}">
  {era.code}
</span>
