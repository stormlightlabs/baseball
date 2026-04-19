<script lang="ts">
  let { label, active = false, onclick }: { label: string; active?: boolean; onclick?: () => void } = $props();
</script>

<button
  class="rounded-full border px-4 py-1 font-sans text-xs font-semibold transition-colors duration-150 {active
    ? 'border-primary text-primary'
    : 'border-outline text-muted hover:border-primary hover:text-primary'}"
  {onclick}>
  {label}
</button>
