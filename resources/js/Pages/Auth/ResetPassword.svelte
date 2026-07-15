<script lang="ts">
  import { useForm } from '@inertiajs/svelte'
  import AuthLayout from '@/Layouts/AuthLayout.svelte'
  import { Button } from '@/components/ui/button'
  import * as Card from '@/components/ui/card'
  import { Input } from '@/components/ui/input'
  import { Label } from '@/components/ui/label'
  import { routes } from '@/routes'
  let { token, errors = {} }: { token: string; errors?: Record<string, string> } = $props()
  const form = useForm({ resetPasswordToken: (() => token)(), password: '', confirmPassword: '' })
</script>

<AuthLayout>
  <Card.Root>
    <Card.Header><Card.Title>Choose a new password</Card.Title><Card.Description>Use at least 12 characters.</Card.Description></Card.Header>
    <Card.Content>
      <form class="space-y-5" onsubmit={(event) => { event.preventDefault(); $form.put(routes.passwordUpdate()) }}>
        <div class="space-y-2"><Label for="password">New password</Label><Input id="password" type="password" bind:value={$form.password} autocomplete="new-password" required aria-invalid={!!errors.password} />{#if errors.password}<p class="text-sm text-destructive">{errors.password}</p>{/if}</div>
        <div class="space-y-2"><Label for="confirmPassword">Confirm password</Label><Input id="confirmPassword" type="password" bind:value={$form.confirmPassword} autocomplete="new-password" required aria-invalid={!!errors.confirmPassword} />{#if errors.confirmPassword}<p class="text-sm text-destructive">{errors.confirmPassword}</p>{/if}</div>
        <Button class="w-full" type="submit" disabled={$form.processing}>Reset password</Button>
      </form>
    </Card.Content>
  </Card.Root>
</AuthLayout>
