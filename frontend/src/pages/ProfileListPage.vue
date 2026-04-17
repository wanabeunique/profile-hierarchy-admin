<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/lib/api'
import type { Profile } from '@/lib/types'
import Card from '@/components/ui/Card.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Table from '@/components/ui/Table.vue'
import TableHeader from '@/components/ui/TableHeader.vue'
import TableBody from '@/components/ui/TableBody.vue'
import TableRow from '@/components/ui/TableRow.vue'
import TableHead from '@/components/ui/TableHead.vue'
import TableCell from '@/components/ui/TableCell.vue'
import { Plus } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const profiles = ref<Profile[]>([])
const loading = ref(true)
const error = ref('')

const profileType = computed(() => (route.params.type as string) || 'plus')

const typeConfig = computed(() => {
  const configs: Record<string, { label: string; title: string; createLabel: string }> = {
    plus: { label: 'Profile+', title: 'Профили Profile+', createLabel: 'Создать Profile+' },
    standard: { label: 'Profile', title: 'Профили Profile', createLabel: 'Создать Profile' },
    min: { label: 'ProfileMin', title: 'Профили ProfileMin', createLabel: 'Создать ProfileMin' },
  }
  return configs[profileType.value] ?? configs.plus
})

const showSubAccounts = computed(() => profileType.value !== 'min')
const showParent = computed(() => profileType.value !== 'plus')

function formatCurrency(value: number): string {
  return value.toFixed(2)
}

async function loadProfiles() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.get<Profile[]>('/profiles', { params: { type: profileType.value } })
    profiles.value = data
  } catch {
    error.value = 'Ошибка загрузки профилей'
  } finally {
    loading.value = false
  }
}

function goToProfile(id: number) {
  router.push(`/profiles/${id}`)
}

onMounted(loadProfiles)
watch(profileType, loadProfiles)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-[hsl(var(--foreground))]">{{ typeConfig.title }}</h1>
        <p class="text-sm text-[hsl(var(--muted-foreground))]">
          Все профили типа {{ typeConfig.label }}
        </p>
      </div>
      <Button
        v-if="auth.isAdmin"
        @click="router.push(`/profiles/create?type=${profileType}`)"
      >
        <Plus class="h-4 w-4 mr-2" />
        {{ typeConfig.createLabel }}
      </Button>
    </div>

    <div v-if="loading" class="text-[hsl(var(--muted-foreground))]">Загрузка...</div>
    <div v-else-if="error" class="text-red-500">{{ error }}</div>

    <Card v-else>
      <CardContent class="!p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Имя</TableHead>
              <TableHead v-if="showParent">Родитель</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Комиссия</TableHead>
              <TableHead v-if="showParent">Комиссия от родителя</TableHead>
              <TableHead>Заказы</TableHead>
              <TableHead v-if="showSubAccounts">Суб-аккаунты</TableHead>
              <TableHead>Общая сумма</TableHead>
              <TableHead>Оплачено</TableHead>
              <TableHead>Остаток</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="profile in profiles"
              :key="profile.id"
              :clickable="true"
              @click="goToProfile(profile.id)"
            >
              <TableCell class="font-medium">{{ profile.name }}</TableCell>
              <TableCell v-if="showParent">
                <RouterLink
                  v-if="profile.parent"
                  :to="`/profiles/${profile.parent.id}`"
                  class="text-[hsl(var(--primary))] hover:underline"
                  @click.stop
                >
                  {{ profile.parent.name }}
                </RouterLink>
                <span v-else class="text-[hsl(var(--muted-foreground))]">—</span>
              </TableCell>
              <TableCell>{{ profile.email }}</TableCell>
              <TableCell>{{ profile.commission }}%</TableCell>
              <TableCell v-if="showParent">
                <span v-if="profile.parent">{{ profile.parent.commission }}%</span>
                <span v-else class="text-[hsl(var(--muted-foreground))]">—</span>
              </TableCell>
              <TableCell>{{ profile.orderCount }}</TableCell>
              <TableCell v-if="showSubAccounts">
                <Badge variant="secondary">{{ profile.subAccountCount }}</Badge>
              </TableCell>
              <TableCell>{{ formatCurrency(profile.totalAmount) }}</TableCell>
              <TableCell>{{ formatCurrency(profile.totalPaid) }}</TableCell>
              <TableCell>{{ formatCurrency(profile.totalRemaining) }}</TableCell>
            </TableRow>
            <TableRow v-if="profiles.length === 0">
              <TableCell
                :colspan="7 + (showSubAccounts ? 1 : 0) + (showParent ? 2 : 0)"
                class="text-center text-[hsl(var(--muted-foreground))]"
              >
                Нет профилей
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
