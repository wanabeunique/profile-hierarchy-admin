<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/lib/api'
import type { Profile, Order } from '@/lib/types'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import Dialog from '@/components/ui/Dialog.vue'
import Table from '@/components/ui/Table.vue'
import TableHeader from '@/components/ui/TableHeader.vue'
import TableBody from '@/components/ui/TableBody.vue'
import TableRow from '@/components/ui/TableRow.vue'
import TableHead from '@/components/ui/TableHead.vue'
import TableCell from '@/components/ui/TableCell.vue'
import {
  Plus,
  ChevronRight,
  DollarSign,
  CreditCard,
  AlertCircle,
  ShoppingCart,
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const profile = ref<Profile | null>(null)
const orders = ref<Order[]>([])
const children = ref<Profile[]>([])
const breadcrumbs = ref<{ id: number; name: string; type: string }[]>([])
const loading = ref(true)
const error = ref('')
const payDialogOpen = ref(false)
const payingOrder = ref<Order | null>(null)
const payAmount = ref('')
const payError = ref('')
const payLoading = ref(false)

function formatCurrency(value: number): string {
  return value.toFixed(2)
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  const day = String(d.getDate()).padStart(2, '0')
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const year = d.getFullYear()
  return `${day}.${month}.${year}`
}

function typeBadgeLabel(type: string): string {
  if (type === 'plus') return 'Profile+'
  if (type === 'standard') return 'Profile'
  if (type === 'min') return 'ProfileMin'
  return type
}

function statusBadgeVariant(status: string): 'destructive' | 'secondary' | 'success' {
  if (status === 'Не оплачено') return 'destructive'
  if (status === 'Частично оплачено') return 'secondary'
  return 'success'
}

const canCreateChild = computed(() => {
  if (!profile.value) return false
  if (auth.isAdmin) return profile.value.type === 'plus' || profile.value.type === 'standard'
  return auth.canCreateChildren
})

const childType = computed(() => {
  if (!profile.value) return 'standard'
  if (profile.value.type === 'plus') return 'standard'
  return 'min'
})

const hasAccess = computed(() => {
  if (auth.isAdmin) return true
  if (!profile.value) return false
  return auth.profileId === profile.value.id
})

async function buildBreadcrumbs(p: Profile) {
  const crumbs: { id: number; name: string; type: string }[] = []
  crumbs.push({ id: p.id, name: p.name, type: p.type })

  if (auth.isAdmin) {
    // Admin can fetch parent profiles via API
    let current = p
    while (current.parentId) {
      try {
        const { data } = await api.get<Profile>(`/profiles/${current.parentId}`)
        crumbs.unshift({ id: data.id, name: data.name, type: data.type })
        current = data
      } catch {
        break
      }
    }
  } else {
    // Non-admin: use already loaded parent data (no extra API calls)
    if (p.parent) {
      crumbs.unshift({ id: p.parent.id, name: p.parent.name, type: p.parent.type })
    }
  }
  breadcrumbs.value = crumbs
}

async function loadProfile() {
  const id = route.params.id as string
  loading.value = true
  error.value = ''

  try {
    const { data } = await api.get<Profile>(`/profiles/${id}`)
    profile.value = data

    if (!hasAccess.value && !auth.isAdmin) {
      error.value = 'У вас нет доступа к этому профилю'
      loading.value = false
      return
    }

    await buildBreadcrumbs(data)

    const [ordersRes, childrenRes] = await Promise.all([
      api.get<Order[]>(`/profiles/${id}/orders`),
      data.type !== 'min'
        ? api.get<Profile[]>(`/profiles?parent_id=${id}`)
        : Promise.resolve({ data: [] as Profile[] }),
    ])

    orders.value = ordersRes.data
    children.value = childrenRes.data
  } catch {
    error.value = 'Ошибка загрузки профиля'
  } finally {
    loading.value = false
  }
}

function openPayDialog(order: Order) {
  payingOrder.value = order
  payAmount.value = ''
  payError.value = ''
  payDialogOpen.value = true
}

async function submitPayment() {
  if (!payingOrder.value) return
  const amount = parseFloat(payAmount.value)
  if (isNaN(amount) || amount <= 0) {
    payError.value = 'Введите корректную сумму'
    return
  }
  if (amount > payingOrder.value.remaining) {
    payError.value = `Сумма не может превышать остаток (${formatCurrency(payingOrder.value.remaining)})`
    return
  }

  payLoading.value = true
  payError.value = ''
  try {
    await api.patch(`/orders/${payingOrder.value.id}/pay`, { amount })
    payDialogOpen.value = false
    await loadProfile()
  } catch {
    payError.value = 'Ошибка при оплате'
  } finally {
    payLoading.value = false
  }
}

function goToChild(id: number) {
  router.push(`/profiles/${id}`)
}

onMounted(loadProfile)

watch(() => route.params.id, loadProfile)
</script>

<template>
  <div class="space-y-6">
    <!-- Loading / Error -->
    <div v-if="loading" class="text-[hsl(var(--muted-foreground))]">Загрузка...</div>
    <div v-else-if="error" class="text-red-500">{{ error }}</div>

    <template v-else-if="profile">
      <!-- Breadcrumbs -->
      <nav class="flex items-center gap-1 text-sm text-[hsl(var(--muted-foreground))]">
        <RouterLink
          v-if="auth.isAdmin"
          to="/"
          class="hover:text-[hsl(var(--foreground))] transition-colors"
        >
          Дашборд
        </RouterLink>
        <template v-for="(crumb, index) in breadcrumbs" :key="crumb.id">
          <ChevronRight class="h-3 w-3" />
          <RouterLink
            v-if="index < breadcrumbs.length - 1 && auth.isAdmin"
            :to="`/profiles/${crumb.id}`"
            class="hover:text-[hsl(var(--foreground))] transition-colors"
          >
            {{ crumb.name }}
          </RouterLink>
          <span v-else class="text-[hsl(var(--foreground))]">{{ crumb.name }}</span>
        </template>
      </nav>

      <!-- Header -->
      <div class="flex items-center gap-3">
        <h1 class="text-2xl font-bold text-[hsl(var(--foreground))]">{{ profile.name }}</h1>
        <Badge variant="secondary">{{ typeBadgeLabel(profile.type) }}</Badge>
      </div>

      <!-- Info card -->
      <Card>
        <CardHeader>
          <CardTitle>Информация о профиле</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 text-sm">
            <div>
              <span class="text-[hsl(var(--muted-foreground))]">Email</span>
              <p class="font-medium">{{ profile.email }}</p>
            </div>
            <div>
              <span class="text-[hsl(var(--muted-foreground))]">Комиссия</span>
              <p class="font-medium">{{ profile.commission }}%</p>
            </div>
            <div v-if="profile.parent">
              <span class="text-[hsl(var(--muted-foreground))]">
                Комиссия от {{ profile.parent.type === 'plus' ? 'Profile+' : 'Profile' }}
              </span>
              <p class="font-medium">
                <RouterLink
                  v-if="auth.isAdmin"
                  :to="`/profiles/${profile.parent.id}`"
                  class="text-[hsl(var(--primary))] hover:underline"
                >
                  {{ profile.parent.name }}
                </RouterLink>
                <span v-else>{{ profile.parent.name }}</span>
                — {{ profile.parent.commission }}%
              </p>
            </div>
            <div>
              <span class="text-[hsl(var(--muted-foreground))]">Срок оплаты</span>
              <p class="font-medium">{{ profile.paymentDeadline ? formatDate(profile.paymentDeadline) : 'Не указан' }}</p>
            </div>
            <div>
              <span class="text-[hsl(var(--muted-foreground))]">Дата создания</span>
              <p class="font-medium">{{ formatDate(profile.createdAt) }}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Stats -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent class="!p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-[hsl(var(--secondary))]">
                <DollarSign class="h-5 w-5 text-[hsl(var(--muted-foreground))]" />
              </div>
              <div>
                <p class="text-sm text-[hsl(var(--muted-foreground))]">Общая сумма</p>
                <p class="text-xl font-bold">{{ formatCurrency(profile.totalAmount) }}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="!p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-green-900/30">
                <CreditCard class="h-5 w-5 text-green-500" />
              </div>
              <div>
                <p class="text-sm text-[hsl(var(--muted-foreground))]">Оплачено</p>
                <p class="text-xl font-bold">{{ formatCurrency(profile.totalPaid) }}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="!p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-red-900/30">
                <AlertCircle class="h-5 w-5 text-red-500" />
              </div>
              <div>
                <p class="text-sm text-[hsl(var(--muted-foreground))]">Остаток</p>
                <p class="text-xl font-bold">{{ formatCurrency(profile.totalRemaining) }}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="!p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-[hsl(var(--secondary))]">
                <ShoppingCart class="h-5 w-5 text-[hsl(var(--muted-foreground))]" />
              </div>
              <div>
                <p class="text-sm text-[hsl(var(--muted-foreground))]">Заказы</p>
                <p class="text-xl font-bold">{{ profile.orderCount }}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Sub-accounts -->
      <Card v-if="profile.type !== 'min'">
        <CardHeader>
          <div class="flex items-center justify-between">
            <CardTitle>Суб-аккаунты</CardTitle>
            <Button
              v-if="canCreateChild"
              size="sm"
              @click="router.push(`/profiles/create?type=${childType}&parent_id=${profile.id}`)"
            >
              <Plus class="h-4 w-4 mr-1" />
              Создать {{ childType === 'standard' ? 'Profile' : 'ProfileMin' }}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Имя</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Комиссия</TableHead>
                <TableHead>Заказы</TableHead>
                <TableHead>Суб-аккаунты</TableHead>
                <TableHead>Общая сумма</TableHead>
                <TableHead>Оплачено</TableHead>
                <TableHead>Остаток</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow
                v-for="child in children"
                :key="child.id"
                :clickable="true"
                @click="goToChild(child.id)"
              >
                <TableCell class="font-medium">
                  {{ child.name }}
                  <Badge variant="secondary" class="ml-2">{{ typeBadgeLabel(child.type) }}</Badge>
                </TableCell>
                <TableCell>{{ child.email }}</TableCell>
                <TableCell>{{ child.commission }}%</TableCell>
                <TableCell>{{ child.orderCount }}</TableCell>
                <TableCell>
                  <Badge variant="secondary">{{ child.subAccountCount }}</Badge>
                </TableCell>
                <TableCell>{{ formatCurrency(child.totalAmount) }}</TableCell>
                <TableCell>{{ formatCurrency(child.totalPaid) }}</TableCell>
                <TableCell>{{ formatCurrency(child.totalRemaining) }}</TableCell>
              </TableRow>
              <TableRow v-if="children.length === 0">
                <TableCell colspan="8" class="text-center text-[hsl(var(--muted-foreground))]">
                  Нет суб-аккаунтов
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <!-- Orders -->
      <Card>
        <CardHeader>
          <div class="flex items-center justify-between">
            <CardTitle>Заказы</CardTitle>
            <Button
              size="sm"
              @click="router.push(`/profiles/${profile.id}/orders/create`)"
            >
              <Plus class="h-4 w-4 mr-1" />
              Создать заказ
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Номер</TableHead>
                <TableHead>Сумма</TableHead>
                <TableHead>Оплачено</TableHead>
                <TableHead>Остаток</TableHead>
                <TableHead>Статус</TableHead>
                <TableHead>Действия</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="order in orders" :key="order.id">
                <TableCell class="font-medium">{{ order.number }}</TableCell>
                <TableCell>{{ formatCurrency(order.amount) }}</TableCell>
                <TableCell>{{ formatCurrency(order.paid) }}</TableCell>
                <TableCell>{{ formatCurrency(order.remaining) }}</TableCell>
                <TableCell>
                  <Badge :variant="statusBadgeVariant(order.status)">
                    {{ order.status }}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Button
                    v-if="order.status !== 'Оплачено'"
                    variant="outline"
                    size="sm"
                    @click="openPayDialog(order)"
                  >
                    Оплатить
                  </Button>
                </TableCell>
              </TableRow>
              <TableRow v-if="orders.length === 0">
                <TableCell colspan="6" class="text-center text-[hsl(var(--muted-foreground))]">
                  Нет заказов
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </template>

    <!-- Pay dialog -->
    <Dialog
      :open="payDialogOpen"
      :title="'Оплата заказа ' + (payingOrder?.number ?? '')"
      description="Введите сумму для оплаты"
      @update:open="payDialogOpen = $event"
    >
      <div class="space-y-4">
        <div v-if="payingOrder" class="text-sm text-[hsl(var(--muted-foreground))]">
          Остаток к оплате: {{ formatCurrency(payingOrder.remaining) }}
        </div>
        <div class="space-y-2">
          <Label for="pay-amount">Сумма</Label>
          <Input
            id="pay-amount"
            v-model="payAmount"
            type="number"
            placeholder="Введите сумму"
            :disabled="payLoading"
          />
        </div>
        <div v-if="payError" class="text-sm text-red-500">{{ payError }}</div>
        <div class="flex justify-end gap-2">
          <Button variant="outline" @click="payDialogOpen = false" :disabled="payLoading">
            Отмена
          </Button>
          <Button @click="submitPayment" :disabled="payLoading">
            {{ payLoading ? 'Оплата...' : 'Оплатить' }}
          </Button>
        </div>
      </div>
    </Dialog>
  </div>
</template>
