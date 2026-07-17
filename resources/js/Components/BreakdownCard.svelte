<script lang="ts">
  import { geoNaturalEarth1, geoPath } from 'd3-geo'
  import iso from 'iso-3166-1'
  import { feature } from 'topojson-client'
  import type { GeometryCollection, Topology } from 'topojson-specification'
  import world from 'world-atlas/countries-110m.json'
  import * as Card from '@/components/ui/card'

  type Item = { name: string; views: number; code?: string }
  type Group = { title: string; label: string; items: Item[]; map?: boolean }

  const topology = world as unknown as Topology
  const countries = feature(topology, topology.objects.countries as GeometryCollection<{ name: string }>)
  const mapPath = geoPath(geoNaturalEarth1().fitSize([1010, 500], countries))

  let { groups }: { groups: Group[] } = $props()
  let selected = $state(0)
  const items = $derived(groups[selected]?.items ?? [])
  const max = $derived(Math.max(...items.map((item) => item.views), 1))
  const countryViews = (id: string | number | undefined) => {
    const code = iso.whereNumeric(String(id))?.alpha2
    return items.find((item) => item.code?.toUpperCase() === code)?.views ?? 0
  }
</script>

<Card.Root size="sm" class="h-full min-h-80 gap-0 py-0">
  <div class="flex min-h-12 items-stretch gap-4 border-b px-4" role="tablist">
    {#each groups as group, index}
      <button
        type="button"
        role="tab"
        aria-selected={selected === index}
        class={`border-b-2 px-2 text-xs font-semibold uppercase tracking-wide ${selected === index ? 'border-primary bg-primary/10 text-primary' : 'border-transparent text-muted-foreground hover:bg-muted hover:text-foreground'}`}
        onclick={() => selected = index}
      >{group.title}</button>
    {/each}
  </div>

  <Card.Content class="p-4">
    {#if groups[selected]?.map}
      <svg viewBox="0 0 1010 500" class="mt-2 w-full" role="img" aria-label="Pageviews by country">
        {#each countries.features as country}
          {@const views = countryViews(country.id)}
          <path
            d={mapPath(country) ?? ''}
            class={views ? 'fill-primary stroke-background' : 'fill-muted stroke-background'}
            style={`fill-opacity: ${views ? 0.3 + (views / max) * 0.7 : 1}`}
            stroke-width="0.8"
          ><title>{country.properties?.name}: {views.toLocaleString()} views</title></path>
        {/each}
      </svg>
    {:else}
      <div class="mb-2 flex items-center justify-between text-xs text-muted-foreground">
        <span>{groups[selected]?.label ?? ''}</span><span>Views</span>
      </div>
      {#if items.length === 0}
        <p class="py-16 text-center text-sm text-muted-foreground">No data for this period.</p>
      {:else}
        <ol class="space-y-1">
          {#each items.slice(0, 8) as item}
            <li class="relative flex h-8 items-center justify-between gap-3 overflow-hidden px-2 text-sm">
              <div class="absolute inset-y-0 left-0 bg-muted" style={`width: ${Math.max((item.views / max) * 74, 3)}%`}></div>
              <span class="relative truncate font-medium">{item.name || 'Unknown'}</span>
              <span class="relative shrink-0 text-xs tabular-nums text-muted-foreground">{item.views.toLocaleString()}</span>
            </li>
          {/each}
        </ol>
      {/if}
    {/if}
  </Card.Content>
</Card.Root>
