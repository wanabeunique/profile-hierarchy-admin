<script setup lang="ts">
import { useRoute } from 'vue-router'
import AdminLayout from '@/layouts/AdminLayout.vue'

const route = useRoute()
</script>

<template>
  <component :is="route.path === '/login' ? 'div' : AdminLayout">
    <RouterView />
  </component>
</template>
