<script lang="ts">
  import { useForm } from '@inertiajs/svelte'
  import WebsiteForm from '@/Components/WebsiteForm.svelte'
  import DashboardLayout from '@/Layouts/DashboardLayout.svelte'
  import * as Breadcrumb from '@/components/ui/breadcrumb'
  import * as Card from '@/components/ui/card'
  import { routes } from '@/routes'
  type Website = { ID: string; Name: string; Domain: string }
  let { websites }: { websites: Website[] } = $props()
  const form = useForm({ name: '', domain: '' })
</script>

<DashboardLayout {websites}>
  <div class="mx-auto max-w-2xl space-y-6">
    <Breadcrumb.Root><Breadcrumb.List><Breadcrumb.Item><Breadcrumb.Link href={routes.websiteIndex()}>Websites</Breadcrumb.Link></Breadcrumb.Item><Breadcrumb.Separator /><Breadcrumb.Item><Breadcrumb.Page>Add website</Breadcrumb.Page></Breadcrumb.Item></Breadcrumb.List></Breadcrumb.Root>
    <Card.Root><Card.Header><Card.Title>Add website</Card.Title><Card.Description>Enter the public domain you want to track.</Card.Description></Card.Header><Card.Content><WebsiteForm {form} label="Add website" submit={(event) => { event.preventDefault(); $form.post(routes.websiteCreate()) }} /></Card.Content></Card.Root>
  </div>
</DashboardLayout>
