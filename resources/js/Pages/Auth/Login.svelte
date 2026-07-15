<script lang="ts">
  import { Link, useForm } from '@inertiajs/svelte'
  import AuthLayout from '@/Layouts/AuthLayout.svelte'
  import { Button } from '@/components/ui/button'
  import * as Card from '@/components/ui/card'
  import { Input } from '@/components/ui/input'
  import { Label } from '@/components/ui/label'
  import { routes } from '@/routes'

  let { errors = {} }: { errors?: Record<string, string> } = $props()
  const form = useForm({ email: '', password: '' })
</script>

<AuthLayout>
  <Card.Root>
    <Card.Header><Card.Title>Sign in</Card.Title><Card.Description>Access your analytics dashboard.</Card.Description></Card.Header>
    <Card.Content>
      <form class="space-y-5" onsubmit={(event) => { event.preventDefault(); $form.post(routes.sessionCreate()) }}>
        <div class="space-y-2"><Label for="email">Email</Label><Input id="email" type="email" bind:value={$form.email} autocomplete="email" required aria-invalid={!!errors.email} />{#if errors.email}<p class="text-sm text-destructive">{errors.email}</p>{/if}</div>
        <div class="space-y-2"><div class="flex justify-between"><Label for="password">Password</Label><Link class="text-sm text-primary hover:underline" href={routes.passwordNew()}>Forgot password?</Link></div><Input id="password" type="password" bind:value={$form.password} autocomplete="current-password" required aria-invalid={!!errors.password} />{#if errors.password}<p class="text-sm text-destructive">{errors.password}</p>{/if}</div>
        <Button class="w-full" type="submit" disabled={$form.processing}>{$form.processing ? 'Signing in…' : 'Sign in'}</Button>
      </form>
    </Card.Content>
  </Card.Root>
</AuthLayout>
