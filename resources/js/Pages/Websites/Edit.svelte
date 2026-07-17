<script lang="ts">
  import { useForm } from '@inertiajs/svelte'
  import WebsiteForm from '@/Components/WebsiteForm.svelte'
  import DashboardLayout from '@/Layouts/DashboardLayout.svelte'
  import * as Card from '@/components/ui/card'
  import { routes } from '@/routes'
  type Website = { ID: string; Name: string; Domain: string }
  let { websites, website }: { websites: Website[]; website: Website } = $props()
  const form = useForm({ name: (() => website.Name)(), domain: (() => website.Domain)() })
</script>

<DashboardLayout {websites} {website} section="settings">
  <div class="grid min-h-[calc(100vh-8rem)] place-items-center">
    <Card.Root class="w-full max-w-2xl"><Card.Header><Card.Title>Edit website</Card.Title><Card.Description>Update the dashboard label or tracked domain.</Card.Description></Card.Header><Card.Content><WebsiteForm {form} label="Save changes" submit={(event) => { event.preventDefault(); $form.put(routes.websiteUpdate(website.ID)) }} /></Card.Content></Card.Root>
  </div>
</DashboardLayout>
