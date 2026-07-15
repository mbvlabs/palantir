<script lang="ts">
  import * as Card from '@/components/ui/card'
  let { title, items }: { title: string; items: Array<{ name: string; views: number }> } = $props()
  const max = $derived(Math.max(...items.map((item) => item.views), 1))
</script>

<Card.Root>
  <Card.Header class="border-b pb-4"><Card.Title class="font-heading text-lg">{title}</Card.Title></Card.Header>
  <Card.Content class="pt-5">
    {#if items.length === 0}
      <p class="py-8 text-center text-sm text-muted-foreground">No data for this period.</p>
    {:else}
      <ol class="space-y-4">
        {#each items.slice(0, 8) as item}
          <li>
            <div class="mb-1.5 flex items-center justify-between gap-4 text-sm"><span class="truncate font-medium">{item.name || 'Unknown'}</span><span class="shrink-0 tabular-nums text-muted-foreground">{item.views.toLocaleString()}</span></div>
            <div class="h-1 overflow-hidden bg-muted"><div class="h-full bg-foreground/55" style={`width: ${Math.max((item.views / max) * 100, 3)}%`}></div></div>
          </li>
        {/each}
      </ol>
    {/if}
  </Card.Content>
</Card.Root>
