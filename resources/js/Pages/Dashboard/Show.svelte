<script lang="ts">
  import { router } from '@inertiajs/svelte'
  import { onMount } from 'svelte'
  import { LineChart } from 'layerchart'
  import BreakdownCard from '@/Components/BreakdownCard.svelte'
  import DashboardLayout from '@/Layouts/DashboardLayout.svelte'
  import { Button } from '@/components/ui/button'
  import * as Card from '@/components/ui/card'
  import * as Chart from '@/components/ui/chart'
  import { Input } from '@/components/ui/input'
  import { Label } from '@/components/ui/label'
  import { routes } from '@/routes'

  type Website = { ID: string; Name: string; Domain: string }
  type Bucket = { time: string; count: number }
  type Item = { name: string; views: number; code?: string }
  type Stats = {
    TotalPageviews: number; TotalUniqueVisitors: number; ViewsPerVisitor: number; BounceRate: number
    PageviewsChange: number; UniqueVisitorsChange: number; ViewsPerVisitorChange: number; BounceRateChange: number
    PageviewsOverTime: Bucket[]; VisitorsOverTime: Bucket[]; EventsOverTime: Bucket[]
    TopPages: Item[]; TopReferrers: Item[]; Browsers: Item[]; OSes: Item[]; Devices: Item[]; TopCountries: Item[]; TopCities: Item[]; TopEvents: Item[]
  }
  let { websites, website, stats, period, start = '', end = '', bucket }: { websites: Website[]; website: Website; stats: Stats; period: string; start?: string; end?: string; bucket: string } = $props()
  let customStart = $state((() => start)())
  let customEnd = $state((() => end)())
  let refreshing = $state(false)
  let refreshFailed = $state(false)
  const chartConfig = { pageviews: { label: 'Pageviews', color: 'var(--chart-1)' }, visitors: { label: 'Visitors', color: 'var(--chart-2)' } } satisfies Chart.ChartConfig
  const eventChartConfig = { events: { label: 'Events', color: 'var(--chart-3)' } } satisfies Chart.ChartConfig
  const chartData = $derived(stats.PageviewsOverTime.map((point, index) => ({
    label: new Date(point.time).toLocaleDateString(undefined, bucket === 'hour' ? { day: 'numeric', hour: 'numeric' } : { month: 'short', day: 'numeric' }),
    pageviews: point.count,
    visitors: stats.VisitorsOverTime[index]?.count ?? 0,
  })))
  const eventChartData = $derived(stats.EventsOverTime.map((point) => ({
    label: new Date(point.time).toLocaleDateString(undefined, bucket === 'hour' ? { day: 'numeric', hour: 'numeric' } : { month: 'short', day: 'numeric' }),
    events: point.count,
  })))
  const eventTotal = $derived(stats.EventsOverTime.reduce((total, point) => total + point.count, 0))
  const dashboard = $derived(routes.websiteDashboard(website.ID))
  const periodURL = (value: string) => `${dashboard}?period=${value}`
  const metrics = $derived([
    { label: 'Unique visitors', value: stats.TotalUniqueVisitors.toLocaleString(), change: stats.UniqueVisitorsChange },
    { label: 'Pageviews', value: stats.TotalPageviews.toLocaleString(), change: stats.PageviewsChange },
    { label: 'Views per visitor', value: stats.ViewsPerVisitor.toFixed(1), change: stats.ViewsPerVisitorChange },
    { label: 'Bounce rate', value: `${stats.BounceRate.toFixed(1)}%`, change: stats.BounceRateChange },
  ])

  onMount(() => {
    const timer = window.setInterval(() => router.reload({
      only: ['stats'],
      onStart: () => { refreshing = true; refreshFailed = false },
      onError: () => { refreshFailed = true },
      onFinish: () => { refreshing = false },
    }), 15000)
    return () => window.clearInterval(timer)
  })
</script>

<DashboardLayout {websites} {website} section="dashboard">
  <div class="mx-auto max-w-7xl space-y-8">
    <section class="flex flex-col gap-5 xl:flex-row xl:items-end xl:justify-between">
      <div class="max-w-2xl">
        <p class="mb-2 text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">{website.Name}</p>
        <h1 class="font-heading text-3xl font-semibold tracking-tight sm:text-4xl">Analytics overview</h1>
        <p class="mt-2 text-sm text-muted-foreground">{refreshing ? 'Refreshing data…' : refreshFailed ? 'Live refresh failed. Existing data is still shown.' : 'Live data refreshes every 15 seconds.'}</p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <nav class="flex items-center border bg-card p-1 shadow-xs" aria-label="Date range">
          {#each [['today', 'Today'], ['7d', '7 days'], ['30d', '30 days'], ['month', 'Month']] as option}
            <Button href={periodURL(option[0])} variant={period === option[0] ? 'default' : 'ghost'} size="sm" class="px-3">{option[1]}</Button>
          {/each}
        </nav>
        <details class="group relative">
          <summary class="flex h-9 cursor-pointer list-none items-center border bg-background px-3 text-sm font-medium shadow-xs hover:bg-muted">Custom</summary>
          <form class="absolute right-0 top-12 z-30 grid w-72 gap-4 border bg-popover p-4 text-popover-foreground shadow-xl" onsubmit={(event) => { event.preventDefault(); router.get(dashboard, { period: 'custom', start: customStart, end: customEnd }) }}>
            <div class="space-y-2"><Label for="start">Start date</Label><Input id="start" type="date" bind:value={customStart} required /></div>
            <div class="space-y-2"><Label for="end">End date</Label><Input id="end" type="date" bind:value={customEnd} required /></div>
            <Button type="submit">Apply range</Button>
          </form>
        </details>
      </div>
    </section>

    <section class="grid gap-px overflow-hidden border bg-border sm:grid-cols-2 xl:grid-cols-4" aria-label="Summary metrics">
      {#each metrics as item}
        <article class="bg-card p-5 sm:p-6">
          <p class="text-sm text-muted-foreground">{item.label}</p>
          <div class="mt-3 flex items-end justify-between gap-3">
            <p class="font-heading text-3xl font-semibold tabular-nums sm:text-4xl">{item.value}</p>
            <p class:text-accent-foreground={item.change >= 0} class:text-destructive={item.change < 0} class="pb-1 text-xs font-medium tabular-nums">{item.change > 0 ? '+' : ''}{item.change.toFixed(0)}%</p>
          </div>
          <p class="mt-1 text-xs text-muted-foreground">versus previous period</p>
        </article>
      {/each}
    </section>

    <Card.Root>
      <Card.Header class="border-b"><Card.Title class="font-heading text-xl">Traffic</Card.Title><Card.Description>Pageviews and unique visitors across the selected period.</Card.Description></Card.Header>
      <Card.Content class="pt-6">
        {#if stats.TotalPageviews === 0}
          <div class="grid min-h-80 place-items-center"><div class="text-center"><p class="font-heading text-lg font-medium">No activity yet</p><p class="mt-1 text-sm text-muted-foreground">Try another period or check the tracking setup.</p></div></div>
        {:else}
          <Chart.Container config={chartConfig} class="min-h-80 w-full"><LineChart data={chartData} x="label" series={[{ key: 'pageviews', label: 'Pageviews', color: 'var(--color-pageviews)' }, { key: 'visitors', label: 'Visitors', color: 'var(--color-visitors)' }]} axis legend /></Chart.Container>
        {/if}
      </Card.Content>
    </Card.Root>

    <section class="grid gap-5 xl:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]" aria-label="Events">
      <Card.Root>
        <Card.Header class="border-b sm:flex-row sm:items-start sm:justify-between">
          <div><Card.Title class="font-heading text-xl">Event activity</Card.Title><Card.Description>Tracked interactions across the selected period.</Card.Description></div>
          <div class="mt-3 sm:mt-0 sm:text-right"><p class="font-heading text-2xl font-semibold tabular-nums">{eventTotal.toLocaleString()}</p><p class="text-xs text-muted-foreground">total events</p></div>
        </Card.Header>
        <Card.Content class="pt-6">
          {#if eventTotal === 0}
            <div class="grid min-h-72 place-items-center"><div class="text-center"><p class="font-heading text-lg font-medium">No events yet</p><p class="mt-1 text-sm text-muted-foreground">Add data-palantir-event to an element to track interactions.</p></div></div>
          {:else}
            <Chart.Container config={eventChartConfig} class="min-h-72 w-full"><LineChart data={eventChartData} x="label" series={[{ key: 'events', label: 'Events', color: 'var(--color-events)' }]} axis /></Chart.Container>
          {/if}
        </Card.Content>
      </Card.Root>

      <Card.Root>
        <Card.Header class="border-b pb-4"><Card.Title class="font-heading text-lg">Top events</Card.Title><Card.Description>Occurrences and share of all events.</Card.Description></Card.Header>
        <Card.Content class="pt-5">
          {#if stats.TopEvents.length === 0}
            <p class="py-8 text-center text-sm text-muted-foreground">No data for this period.</p>
          {:else}
            <ol class="space-y-4">
              {#each stats.TopEvents as item}
                <li>
                  <div class="mb-1.5 flex items-center justify-between gap-4 text-sm"><span class="truncate font-medium">{item.name || 'Unknown'}</span><span class="shrink-0 tabular-nums text-muted-foreground">{item.views.toLocaleString()} · {((item.views / eventTotal) * 100).toFixed(0)}%</span></div>
                  <div class="h-1 overflow-hidden bg-muted"><div class="h-full bg-foreground/55" style={`width: ${Math.max((item.views / eventTotal) * 100, 3)}%`}></div></div>
                </li>
              {/each}
            </ol>
          {/if}
        </Card.Content>
      </Card.Root>
    </section>

    <section class="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
      <BreakdownCard title="Top pages" items={stats.TopPages} />
      <BreakdownCard title="Referrers" items={stats.TopReferrers} />
      <BreakdownCard title="Countries" items={stats.TopCountries} />
      <BreakdownCard title="Cities" items={stats.TopCities} />
      <BreakdownCard title="Browsers" items={stats.Browsers} />
      <BreakdownCard title="Operating systems" items={stats.OSes} />
      <BreakdownCard title="Devices" items={stats.Devices} />
    </section>
  </div>
</DashboardLayout>
