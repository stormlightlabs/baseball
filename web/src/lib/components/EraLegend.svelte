<script lang="ts">
  import { STATIC_ERAS } from '$lib/eras';
  import EraBadge from './EraBadge.svelte';

  let { compact = false }: { compact?: boolean } = $props();
</script>

<div class={compact ? 'space-y-1' : 'space-y-1.5'}>
  {#each STATIC_ERAS as era (era.code)}
    <div class="flex items-start gap-2.5">
      <EraBadge {era} size={compact ? 'xs' : 'sm'} />
      <div class="min-w-0 flex-1">
        <span class="text-[0.78rem] text-foreground">{era.label}</span>
        <span class="ml-1.5 font-mono text-xxs text-muted">{era.from}–{era.to}</span>
        {#if era.caveat}
          <span
            class="ml-1.5 inline-block rounded bg-warning/10 px-1 font-mono text-[0.6rem] text-warning"
            title={era.caveat}>
            ⚠ caveat
          </span>
        {/if}
      </div>
    </div>
  {/each}
</div>
