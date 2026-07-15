<script lang="ts">
  import type { Snippet } from 'svelte'
  let { children }: { children: Snippet } = $props()
</script>

<svelte:head><title>Palantir</title></svelte:head>
<main class="grid min-h-screen place-items-center bg-background px-4 py-10 text-foreground">
  <div class="w-full max-w-md">
    <a href="/users/sign-in" class="mb-8 flex items-center justify-center gap-3 text-lg font-semibold">
      <span class="grid size-9 grid-cols-2 gap-1 rounded-md border border-primary/40 bg-card p-1.5">
        <span class="rounded-sm border border-primary"></span><span class="rounded-sm border border-border"></span>
        <span class="rounded-sm border border-border"></span><span class="rounded-sm bg-primary"></span>
      </span>
      Palantir
    </a>
    {@render children()}
  </div>
</main>
