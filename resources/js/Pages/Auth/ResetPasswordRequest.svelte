<script lang="ts">
  import { Link, useForm } from '@inertiajs/svelte'
  import AuthLayout from '@/Layouts/AuthLayout.svelte'
  import { Button } from '@/components/ui/button'
  import * as Card from '@/components/ui/card'
  import { Input } from '@/components/ui/input'
  import { Label } from '@/components/ui/label'
  import { routes } from '@/routes'
  let { errors = {} }: { errors?: Record<string, string> } = $props()
  const form = useForm({ email: '' })
</script>

<AuthLayout>
  <Card.Root>
    <Card.Header><Card.Title>Reset password</Card.Title><Card.Description>We will email you a secure reset link.</Card.Description></Card.Header>
    <Card.Content>
      <form class="space-y-5" onsubmit={(event) => { event.preventDefault(); $form.post(routes.passwordCreate()) }}>
        <div class="space-y-2"><Label for="email">Email</Label><Input id="email" type="email" bind:value={$form.email} autocomplete="email" required aria-invalid={!!errors.email} />{#if errors.email}<p class="text-sm text-destructive">{errors.email}</p>{/if}</div>
        <Button class="w-full" type="submit" disabled={$form.processing}>Send reset link</Button>
      </form>
      <p class="mt-5 text-center text-sm text-muted-foreground"><Link class="text-primary hover:underline" href={routes.sessionNew()}>Back to sign in</Link></p>
    </Card.Content>
  </Card.Root>
</AuthLayout>
