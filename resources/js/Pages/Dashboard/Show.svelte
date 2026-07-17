<script lang="ts">
  import { router } from '@inertiajs/svelte'
  import { onMount } from 'svelte'
  import { Area, AreaChart, LineChart, LinearGradient } from 'layerchart'
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
  type RateBucket = { time: string; rate: number }
  type MetricKey = 'visitors' | 'pageviews' | 'viewsPerVisitor' | 'bounceRate'
  type Item = { name: string; views: number; code?: string }
  type Stats = {
    TotalPageviews: number; TotalUniqueVisitors: number; ViewsPerVisitor: number; BounceRate: number
    PageviewsChange: number; UniqueVisitorsChange: number; ViewsPerVisitorChange: number; BounceRateChange: number
    PageviewsOverTime: Bucket[]; VisitorsOverTime: Bucket[]; EventsOverTime: Bucket[]; BounceRateOverTime: RateBucket[]
    TopPages: Item[]; TopReferrers: Item[]; Browsers: Item[]; OSes: Item[]; Devices: Item[]; TopCountries: Item[]; TopCities: Item[]; TopEvents: Item[]
  }
  let { websites, website, stats, period, start = '', end = '', bucket }: { websites: Website[]; website: Website; stats: Stats; period: string; start?: string; end?: string; bucket: string } = $props()
  let customStart = $state((() => start)())
  let customEnd = $state((() => end)())
  let refreshing = $state(false)
  let refreshFailed = $state(false)
  let activeMetric = $state<MetricKey>('visitors')
  const chartConfig = { value: { label: 'Value', color: 'var(--chart-2)' } } satisfies Chart.ChartConfig
  const eventChartConfig = { events: { label: 'Events', color: 'var(--chart-3)' } } satisfies Chart.ChartConfig
  const xTickStride = $derived(Math.max(1, Math.ceil(stats.PageviewsOverTime.length / 8)))
  const countChartProps = $derived({ xAxis: { ticks: xTickStride }, yAxis: { format: 'integer' as const, ticks: 5 } })
  const trafficChartProps = $derived({ xAxis: { ticks: xTickStride }, yAxis: { format: activeMetric === 'viewsPerVisitor' ? 'decimal' as const : 'integer' as const, ticks: 5 } })
  const chartPadding = { top: 12, right: 16, bottom: 24, left: 8 }
  const chartLabel = (time: string) => bucket === 'hour'
    ? new Date(time).toLocaleTimeString(undefined, { hour: 'numeric' })
    : new Date(time).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
  const metrics = $derived([
    { key: 'visitors' as const, label: 'Unique visitors', value: stats.TotalUniqueVisitors.toLocaleString(), change: stats.UniqueVisitorsChange },
    { key: 'pageviews' as const, label: 'Pageviews', value: stats.TotalPageviews.toLocaleString(), change: stats.PageviewsChange },
    { key: 'viewsPerVisitor' as const, label: 'Views per visitor', value: stats.ViewsPerVisitor.toFixed(1), change: stats.ViewsPerVisitorChange },
    { key: 'bounceRate' as const, label: 'Bounce rate', value: `${stats.BounceRate.toFixed(1)}%`, change: stats.BounceRateChange },
  ])
  const activeMetricLabel = $derived(metrics.find((metric) => metric.key === activeMetric)?.label ?? '')
  const chartData = $derived(stats.PageviewsOverTime.map((point, index) => {
    const visitors = stats.VisitorsOverTime[index]?.count ?? 0
    return {
      label: chartLabel(point.time),
      value: activeMetric === 'visitors' ? visitors
        : activeMetric === 'viewsPerVisitor' ? (visitors ? point.count / visitors : 0)
        : activeMetric === 'bounceRate' ? (stats.BounceRateOverTime[index]?.rate ?? 0)
        : point.count,
    }
  }))
  const eventChartData = $derived(stats.EventsOverTime.map((point) => ({
    label: chartLabel(point.time),
    events: point.count,
  })))
  const eventTotal = $derived(stats.EventsOverTime.reduce((total, point) => total + point.count, 0))
  const dashboard = $derived(routes.websiteDashboard(website.ID))
  const periodURL = (value: string) => `${dashboard}?period=${value}`

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
  <div class="mx-auto max-w-7xl space-y-4">
    <section class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
      <div class="max-w-2xl">
        <p class="mb-2 text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">{website.Name}</p>
        <h1 class="font-heading text-3xl font-semibold tracking-tight sm:text-4xl">{website.Domain}</h1>
        <p class="mt-2 text-sm text-muted-foreground">{refreshing ? 'Refreshing data…' : refreshFailed ? 'Live refresh failed. Existing data is still shown.' : 'Live data refreshes every 15 seconds.'}</p>
      </div>
      <nav class="flex items-center border bg-card p-1 shadow-xs" aria-label="Date range">
        {#each [['today', 'Today'], ['7d', '7 days'], ['30d', '30 days'], ['month', 'Month']] as option}
          <Button href={periodURL(option[0])} variant={period === option[0] ? 'default' : 'ghost'} size="sm" class="px-3">{option[1]}</Button>
        {/each}
        <details class="group relative border-l">
          <summary class="flex h-9 cursor-pointer list-none items-center px-3 text-sm font-medium hover:bg-muted">Custom</summary>
          <form class="absolute right-0 top-12 z-30 grid w-72 gap-4 border bg-popover p-4 text-popover-foreground shadow-xl" onsubmit={(event) => { event.preventDefault(); router.get(dashboard, { period: 'custom', start: customStart, end: customEnd }) }}>
            <div class="space-y-2"><Label for="start">Start date</Label><Input id="start" type="date" bind:value={customStart} required /></div>
            <div class="space-y-2"><Label for="end">End date</Label><Input id="end" type="date" bind:value={customEnd} required /></div>
            <Button type="submit">Apply range</Button>
          </form>
        </details>
      </nav>
    </section>

    <Card.Root size="sm" class="gap-0 py-0">
      <section class="grid gap-px overflow-hidden border-b bg-border sm:grid-cols-2 xl:grid-cols-4" aria-label="Summary metrics">
        {#each metrics as item}
          <button
            type="button"
            aria-pressed={activeMetric === item.key}
            class={`p-3 text-left transition-colors sm:p-4 ${activeMetric === item.key ? 'bg-muted' : 'bg-card hover:bg-muted/50'}`}
            onclick={() => activeMetric = item.key}
          >
            <p class="text-xs text-muted-foreground">{item.label}</p>
            <div class="mt-1 flex items-end justify-between gap-3">
              <p class="font-heading text-xl font-semibold tabular-nums sm:text-2xl">{item.value}</p>
              <p class:text-accent-foreground={item.change >= 0} class:text-destructive={item.change < 0} class="pb-0.5 text-xs font-medium tabular-nums">{item.change > 0 ? '+' : ''}{item.change.toFixed(0)}%</p>
            </div>
          </button>
        {/each}
      </section>
      <Card.Content class="p-4 sm:p-5">
        {#if stats.TotalPageviews === 0}
          <div class="grid h-80 place-items-center"><div class="text-center"><p class="font-heading text-lg font-medium">No activity yet</p><p class="mt-1 text-sm text-muted-foreground">Try another period or check the tracking setup.</p></div></div>
        {:else}
          <Chart.Container config={chartConfig} class="h-80 w-full">
            <AreaChart data={chartData} x="label" series={[{ key: 'value', label: activeMetricLabel, color: 'var(--color-value)' }]} axis props={trafficChartProps} padding={chartPadding}>
              {#snippet marks()}
                <LinearGradient id="traffic-area-gradient" vertical stops={[
                  ['0%', 'color-mix(in srgb, var(--color-value) 30%, transparent)'],
                  ['55%', 'color-mix(in srgb, var(--color-value) 10%, transparent)'],
                  ['100%', 'transparent'],
                ]}>
                  {#snippet children({ gradient })}<Area seriesKey="value" fill={gradient} line />{/snippet}
                </LinearGradient>
              {/snippet}
            </AreaChart>
          </Chart.Container>
        {/if}
      </Card.Content>
    </Card.Root>

    <section class="grid gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]" aria-label="Events">
      <Card.Root size="sm" class="[--card-spacing:--spacing(3)]">
        <Card.Header class="border-b sm:flex sm:flex-row sm:items-start sm:justify-between">
          <div class="space-y-0.5"><Card.Title class="font-heading text-lg">Event activity</Card.Title><Card.Description>Tracked interactions across the selected period.</Card.Description></div>
          <div class="mt-2 sm:mt-0 sm:text-right"><p class="font-heading text-xl font-semibold tabular-nums">{eventTotal.toLocaleString()}</p><p class="text-xs text-muted-foreground">total events</p></div>
        </Card.Header>
        <Card.Content class="pt-3">
          {#if eventTotal === 0}
            <div class="grid h-64 place-items-center"><div class="text-center"><p class="font-heading text-base font-medium">No events yet</p><p class="mt-1 text-sm text-muted-foreground">Add data-palantir-event to an element to track interactions.</p></div></div>
          {:else}
            <Chart.Container config={eventChartConfig} class="h-64 w-full"><LineChart data={eventChartData} x="label" series={[{ key: 'events', label: 'Events', color: 'var(--color-events)' }]} axis props={countChartProps} padding={chartPadding} /></Chart.Container>
          {/if}
        </Card.Content>
      </Card.Root>

      <Card.Root size="sm" class="[--card-spacing:--spacing(3)]">
        <Card.Header class="gap-0.5 border-b"><Card.Title class="font-heading text-lg">Top events</Card.Title><Card.Description>Occurrences and share of all events.</Card.Description></Card.Header>
        <Card.Content class="pt-3">
          {#if stats.TopEvents.length === 0}
            <p class="py-8 text-center text-sm text-muted-foreground">No data for this period.</p>
          {:else}
            <ol class="space-y-3">
              {#each stats.TopEvents as item}
                <li>
                  <div class="mb-1.5 flex items-center justify-between gap-3 text-sm"><span class="truncate font-medium">{item.name || 'Unknown'}</span><span class="shrink-0 text-xs tabular-nums text-muted-foreground">{item.views.toLocaleString()} · {((item.views / eventTotal) * 100).toFixed(0)}%</span></div>
                  <div class="h-1 overflow-hidden bg-muted"><div class="h-full bg-foreground/55" style={`width: ${Math.max((item.views / eventTotal) * 100, 3)}%`}></div></div>
                </li>
              {/each}
            </ol>
          {/if}
        </Card.Content>
      </Card.Root>
    </section>

    <section class="grid gap-4 xl:grid-cols-2">
      <BreakdownCard groups={[{ title: 'Referrers', label: 'Source', items: stats.TopReferrers }]} />
      <BreakdownCard groups={[{ title: 'Top pages', label: 'Page', items: stats.TopPages }]} />
      <BreakdownCard groups={[
        { title: 'Map', label: 'Location', items: stats.TopCountries, map: true },
        { title: 'Countries', label: 'Location', items: stats.TopCountries },
        { title: 'Cities', label: 'Location', items: stats.TopCities },
      ]} />
      <BreakdownCard groups={[
        { title: 'Browsers', label: 'Technology', items: stats.Browsers },
        { title: 'Operating systems', label: 'Technology', items: stats.OSes },
        { title: 'Devices', label: 'Technology', items: stats.Devices },
      ]} />
    </section>
  </div>
</DashboardLayout>
