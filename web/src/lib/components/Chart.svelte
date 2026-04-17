<script lang="ts">
  import type { ChartConfiguration, Chart as ChartType } from 'chart.js';
  import { onDestroy, onMount } from 'svelte';

  let { config, height = 120 }: { config: ChartConfiguration; height?: number } = $props();

  let canvas: HTMLCanvasElement;
  let chart: ChartType | null = $state(null);

  onMount(async () => {
    const { Chart, registerables } = await import('chart.js');
    Chart.register(...registerables);
    chart = new Chart(canvas, config);
  });

  onDestroy(() => {
    chart?.destroy();
    chart = null;
  });

  $effect(() => {
    const data = config.data;
    if (chart) {
      chart.data = data;
      chart.update('none');
    }
  });
</script>

<canvas bind:this={canvas} {height}></canvas>
