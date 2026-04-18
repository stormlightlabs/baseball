<script lang="ts">
  let { text, label = 'copy' }: { text: string; label?: string } = $props();

  let copied = $state(false);

  async function copy() {
    await navigator.clipboard.writeText(text);
    copied = true;
    setTimeout(() => (copied = false), 1500);
  }
</script>

<button
  onclick={copy}
  class="rounded bg-outline px-2.5 py-1 font-mono text-[0.7rem] text-muted transition-colors duration-150 hover:text-foreground">
  {copied ? 'Copied!' : label}
</button>
