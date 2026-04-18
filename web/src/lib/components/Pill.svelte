<script lang="ts">
  let { label, active = false, onclick }: { label: string; active?: boolean; onclick?: () => void } = $props();
</script>

<button
  class="rounded-full border px-3 py-0.75 font-mono text-[0.75rem] transition-colors duration-150 {active
    ? 'border-primary text-primary'
    : 'border-outline text-muted hover:border-primary hover:text-primary'}"
  {onclick}>
  {label}
</button>
