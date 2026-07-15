<script lang="ts">
  import { router } from '@inertiajs/svelte'
  import { ChartNoAxesCombined, CodeXml, LogOut, Plus, Settings2 } from '@lucide/svelte'
  import type { Snippet } from 'svelte'
  import { Button } from '@/components/ui/button'
  import * as DropdownMenu from '@/components/ui/dropdown-menu'
  import * as Sidebar from '@/components/ui/sidebar'
  import { routes } from '@/routes'

  type Website = { ID: string; Name: string; Domain: string }
  let { websites, website, section = 'websites', children }: { websites: Website[]; website?: Website; section?: 'websites' | 'dashboard' | 'tracking' | 'settings'; children: Snippet } = $props()

  function chooseWebsite(event: Event) {
    const id = (event.currentTarget as HTMLSelectElement).value
    if (id) router.visit(routes.websiteDashboard(id))
  }
</script>

<Sidebar.Provider style="--sidebar-width: 16rem; --sidebar-width-icon: 3.5rem;">
  <Sidebar.Root variant="inset" collapsible="icon">
    <Sidebar.Header class="p-3">
      <Sidebar.Menu>
        <Sidebar.MenuItem>
          <Sidebar.MenuButton size="lg" tooltipContent="Palantir">
            {#snippet child({ props })}
              <a {...props} href={routes.websiteIndex()}>
                <span class="grid size-8 shrink-0 place-items-center bg-sidebar-primary text-sm font-semibold text-sidebar-primary-foreground">P</span>
                <span class="grid flex-1 text-left leading-tight"><span class="truncate font-heading text-lg font-semibold">Palantir</span><span class="truncate text-xs text-muted-foreground">Private analytics</span></span>
              </a>
            {/snippet}
          </Sidebar.MenuButton>
        </Sidebar.MenuItem>
      </Sidebar.Menu>
    </Sidebar.Header>

    <Sidebar.Content>
      <Sidebar.Group>
        <Sidebar.GroupLabel>Workspace</Sidebar.GroupLabel>
        <Sidebar.GroupContent>
          <Sidebar.Menu>
            <Sidebar.MenuItem><Sidebar.MenuButton isActive={section === 'websites'} tooltipContent="Add website">{#snippet child({ props })}<a {...props} href={routes.websiteNew()}><Plus /><span>Add website</span></a>{/snippet}</Sidebar.MenuButton></Sidebar.MenuItem>
          </Sidebar.Menu>
        </Sidebar.GroupContent>
      </Sidebar.Group>

      {#if website}
        <Sidebar.Separator />
        <Sidebar.Group>
          <Sidebar.GroupLabel>{website.Name}</Sidebar.GroupLabel>
          <Sidebar.GroupContent>
            <Sidebar.Menu>
              <Sidebar.MenuItem><Sidebar.MenuButton isActive={section === 'dashboard'} tooltipContent="Dashboard">{#snippet child({ props })}<a {...props} href={routes.websiteDashboard(website.ID)}><ChartNoAxesCombined /><span>Dashboard</span></a>{/snippet}</Sidebar.MenuButton></Sidebar.MenuItem>
              <Sidebar.MenuItem><Sidebar.MenuButton isActive={section === 'tracking'} tooltipContent="Tracking setup">{#snippet child({ props })}<a {...props} href={routes.websiteShow(website.ID)}><CodeXml /><span>Tracking setup</span></a>{/snippet}</Sidebar.MenuButton></Sidebar.MenuItem>
              <Sidebar.MenuItem><Sidebar.MenuButton isActive={section === 'settings'} tooltipContent="Website settings">{#snippet child({ props })}<a {...props} href={routes.websiteEdit(website.ID)}><Settings2 /><span>Website settings</span></a>{/snippet}</Sidebar.MenuButton></Sidebar.MenuItem>
            </Sidebar.Menu>
          </Sidebar.GroupContent>
        </Sidebar.Group>
      {/if}
    </Sidebar.Content>

    <Sidebar.Footer class="p-3">
      <div class="border border-sidebar-border bg-sidebar-accent/40 p-3 group-data-[collapsible=icon]:hidden">
        <p class="text-xs font-medium">Collecting quietly</p>
        <p class="mt-1 text-xs leading-relaxed text-muted-foreground">No cookies. No invasive profiles.</p>
      </div>
    </Sidebar.Footer>
    <Sidebar.Rail />
  </Sidebar.Root>

  <Sidebar.Inset class="overflow-hidden">
    <header class="sticky top-0 z-20 flex h-16 items-center gap-3 border-b bg-background/90 px-4 backdrop-blur md:px-6">
      <Sidebar.Trigger />
      <div class="h-5 w-px bg-border"></div>
      <label class="sr-only" for="website-switcher">Select website</label>
      <div class="relative min-w-0">
        <select id="website-switcher" class="h-9 max-w-[15rem] appearance-none border border-input bg-background py-1 pl-3 pr-9 text-sm font-medium shadow-xs outline-none transition-colors hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring sm:min-w-52" value={website?.ID ?? ''} onchange={chooseWebsite}>
          <option value="" disabled>{websites.length ? 'Select a website' : 'No websites yet'}</option>
          {#each websites as site}<option value={site.ID}>{site.Name}</option>{/each}
        </select>
        <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">⌄</span>
      </div>
      {#if website}<span class="hidden truncate text-sm text-muted-foreground lg:block">{website.Domain}</span>{/if}
      <div class="ml-auto">
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>{#snippet child({ props })}<Button {...props} variant="ghost" size="sm">Account</Button>{/snippet}</DropdownMenu.Trigger>
          <DropdownMenu.Content align="end" class="w-44"><DropdownMenu.Label>Account</DropdownMenu.Label><DropdownMenu.Separator /><DropdownMenu.Item onclick={() => router.delete(routes.sessionDestroy())}><LogOut /> Sign out</DropdownMenu.Item></DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
    </header>
    <main class="flex-1 bg-background p-4 md:p-6 lg:p-8">{@render children()}</main>
  </Sidebar.Inset>
</Sidebar.Provider>
