<script lang="ts">
  import { router } from '@inertiajs/svelte'
  import DashboardLayout from '@/Layouts/DashboardLayout.svelte'
  import * as AlertDialog from '@/components/ui/alert-dialog'
  import { Button } from '@/components/ui/button'
  import * as Card from '@/components/ui/card'
  import { routes } from '@/routes'
  type Website = { ID: string; Name: string; Domain: string }
  let { websites, website, trackingScriptURL }: { websites: Website[]; website: Website; trackingScriptURL: string } = $props()
  const snippet = $derived(`<script defer data-website-id="${website.ID}" src="${trackingScriptURL}"><\/script>`)
</script>

<DashboardLayout {websites} {website} section="tracking">
  <div class="mx-auto max-w-4xl space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4"><div><h1 class="text-2xl font-semibold">{website.Name}</h1><p class="text-muted-foreground">{website.Domain}</p></div><div class="flex gap-2"><Button href={routes.websiteDashboard(website.ID)}>View dashboard</Button><Button href={routes.websiteEdit(website.ID)} variant="outline">Edit</Button></div></div>
    <Card.Root><Card.Header><Card.Title>Tracking snippet</Card.Title><Card.Description>Add this before the closing &lt;/head&gt; tag on {website.Domain}.</Card.Description></Card.Header><Card.Content><pre class="overflow-x-auto rounded-md border bg-muted p-4 text-sm"><code>{snippet}</code></pre></Card.Content></Card.Root>
    <Card.Root class="border-destructive/40"><Card.Header><Card.Title>Delete website</Card.Title><Card.Description>This permanently removes the website and all analytics data.</Card.Description></Card.Header><Card.Content>
      <AlertDialog.Root><AlertDialog.Trigger>{#snippet child({ props })}<Button {...props} variant="destructive">Delete website</Button>{/snippet}</AlertDialog.Trigger><AlertDialog.Content><AlertDialog.Header><AlertDialog.Title>Delete {website.Name}?</AlertDialog.Title><AlertDialog.Description>This cannot be undone.</AlertDialog.Description></AlertDialog.Header><AlertDialog.Footer><AlertDialog.Cancel>Cancel</AlertDialog.Cancel><AlertDialog.Action onclick={() => router.delete(routes.websiteDestroy(website.ID))}>Delete</AlertDialog.Action></AlertDialog.Footer></AlertDialog.Content></AlertDialog.Root>
    </Card.Content></Card.Root>
  </div>
</DashboardLayout>
