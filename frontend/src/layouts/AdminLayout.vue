<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Users, User, UserMinus, LogOut, Menu, X } from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const sidebarOpen = ref(false)

const adminNavItems = [
  { to: '/profiles/type/plus', label: 'Profile+', icon: Users },
  { to: '/profiles/type/standard', label: 'Profile', icon: User },
  { to: '/profiles/type/min', label: 'ProfileMin', icon: UserMinus },
]

function isActive(path: string): boolean {
  return route.path === path
}

function handleLogout() {
  auth.logout()
  router.push('/login')
}

function closeSidebar() {
  sidebarOpen.value = false
}
</script>

<template>
  <div class="min-h-screen flex bg-[hsl(var(--background))]">
    <!-- Mobile overlay -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 z-40 bg-black/50 lg:hidden"
      @click="closeSidebar"
    />

    <!-- Sidebar -->
    <aside
      :class="[
        'fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-[hsl(var(--border))] bg-[hsl(var(--card))] transition-transform lg:static lg:translate-x-0',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full'
      ]"
    >
      <!-- Logo -->
      <div class="flex h-14 items-center border-b border-[hsl(var(--border))] px-4">
        <span class="text-lg font-semibold text-[hsl(var(--foreground))]">Админ-панель</span>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 space-y-1 p-4">
        <template v-if="auth.isAdmin">
          <RouterLink
            v-for="item in adminNavItems"
            :key="item.to"
            :to="item.to"
            :class="[
              'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
              isActive(item.to)
                ? 'bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))]'
                : 'text-[hsl(var(--muted-foreground))] hover:bg-[hsl(var(--secondary))] hover:text-[hsl(var(--foreground))]'
            ]"
            @click="closeSidebar"
          >
            <component :is="item.icon" class="h-4 w-4" />
            {{ item.label }}
          </RouterLink>
        </template>
        <template v-else>
          <RouterLink
            :to="`/profiles/${auth.profileId}`"
            class="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-[hsl(var(--muted-foreground))] hover:bg-[hsl(var(--secondary))] hover:text-[hsl(var(--foreground))] transition-colors"
            @click="closeSidebar"
          >
            <User class="h-4 w-4" />
            Мой профиль
          </RouterLink>
        </template>
      </nav>

      <!-- User info -->
      <div class="border-t border-[hsl(var(--border))] p-4">
        <div class="flex items-center gap-3 mb-3">
          <div class="flex h-8 w-8 items-center justify-center rounded-full bg-[hsl(var(--secondary))]">
            <User class="h-4 w-4 text-[hsl(var(--muted-foreground))]" />
          </div>
          <div class="flex-1 min-w-0">
            <Badge variant="secondary">{{ auth.displayRole }}</Badge>
          </div>
        </div>
        <Button variant="ghost" size="sm" class="w-full justify-start gap-2" @click="handleLogout">
          <LogOut class="h-4 w-4" />
          Выйти
        </Button>
      </div>
    </aside>

    <!-- Main content -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Mobile header -->
      <header class="flex h-14 items-center border-b border-[hsl(var(--border))] px-4 lg:hidden">
        <button @click="sidebarOpen = !sidebarOpen" class="text-[hsl(var(--foreground))]">
          <Menu v-if="!sidebarOpen" class="h-5 w-5" />
          <X v-else class="h-5 w-5" />
        </button>
        <span class="ml-3 text-lg font-semibold text-[hsl(var(--foreground))]">Админ-панель</span>
      </header>

      <!-- Page content -->
      <main class="flex-1 p-6 overflow-auto">
        <slot />
      </main>
    </div>
  </div>
</template>
