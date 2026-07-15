<script lang="ts">
  import { Button } from '@/components/ui/button'
  import { Input } from '@/components/ui/input'
  import { Label } from '@/components/ui/label'

  let { form, submit, label }: { form: any; submit: (event: SubmitEvent) => void; label: string } = $props()
</script>

<form class="space-y-5" onsubmit={submit}>
  <div class="space-y-2"><Label for="name">Name</Label><Input id="name" bind:value={$form.name} placeholder="Company website" required aria-invalid={!!$form.errors.name} />{#if $form.errors.name}<p class="text-sm text-destructive">{$form.errors.name}</p>{/if}</div>
  <div class="space-y-2"><Label for="domain">Domain</Label><Input id="domain" bind:value={$form.domain} placeholder="example.com" required aria-invalid={!!$form.errors.domain} />{#if $form.errors.domain}<p class="text-sm text-destructive">{$form.errors.domain}</p>{/if}</div>
  <Button type="submit" disabled={$form.processing}>{$form.processing ? 'Saving…' : label}</Button>
</form>
