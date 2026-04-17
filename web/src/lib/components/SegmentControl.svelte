<script lang="ts">
  let { options, value = $bindable() }: { options: Array<{ id: string; label: string }>; value: string } = $props();
</script>

<div class="flex rounded-lg border border-outline bg-surface p-0.5">
  {#each options as option (`${option.id}:${option.label}`)}
    <button
      class="flex-1 rounded px-3 py-1 font-display text-[0.78rem] transition-all duration-150 {value === option.id
        ? 'bg-crust text-foreground shadow-sm'
        : 'text-muted hover:text-foreground'}"
      onclick={() => (value = option.id)}>
      {option.label}
    </button>
  {/each}
</div>
