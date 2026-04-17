<script setup lang="ts">
defineProps<{
  clickable?: boolean
}>()
</script>

<template>
  <tr
    :class="[
      'border-b border-[hsl(var(--border))] transition-colors hover:bg-[hsl(var(--muted))]/50',
      clickable ? 'cursor-pointer' : ''
    ]"
  >
    <slot />
  </tr>
</template>
