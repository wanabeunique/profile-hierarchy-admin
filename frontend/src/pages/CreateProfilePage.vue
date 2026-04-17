<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/lib/api'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import Button from '@/components/ui/Button.vue'

const route = useRoute()
const router = useRouter()

const profileType = computed(() => (route.query.type as string) || 'plus')
const parentId = computed(() => route.query.parent_id ? Number(route.query.parent_id) : null)

const name = ref('')
const email = ref('')
const password = ref('')
const commission = ref('')
const paymentDeadline = ref('')
const loading = ref(false)
const error = ref('')

const typeLabel = computed(() => {
  if (profileType.value === 'plus') return 'Profile+'
  if (profileType.value === 'standard') return 'Profile'
  return 'ProfileMin'
})

async function handleSubmit() {
  error.value = ''
  loading.value = true

  try {
    const payload: Record<string, unknown> = {
      name: name.value,
      email: email.value,
      password: password.value,
      type: profileType.value,
      commission: parseFloat(commission.value) || 0,
    }
    if (parentId.value) {
      payload.parentId = parentId.value
    }
    if (paymentDeadline.value) {
      payload.paymentDeadline = new Date(paymentDeadline.value).toISOString()
    }

    const { data } = await api.post<{ id: number }>('/profiles', payload)
    router.push(`/profiles/${data.id}`)
  } catch (e: unknown) {
    if (e && typeof e === 'object' && 'response' in e) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Ошибка создания профиля'
    } else {
      error.value = 'Ошибка соединения с сервером'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-lg mx-auto space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-[hsl(var(--foreground))]">Создание {{ typeLabel }}</h1>
      <p class="text-sm text-[hsl(var(--muted-foreground))]">Заполните данные нового профиля</p>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>Данные профиля</CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div class="space-y-2">
            <Label for="name">Имя</Label>
            <Input id="name" v-model="name" placeholder="Введите имя" :disabled="loading" />
          </div>

          <div class="space-y-2">
            <Label for="email">Email</Label>
            <Input id="email" v-model="email" type="email" placeholder="Введите email" :disabled="loading" />
          </div>

          <div class="space-y-2">
            <Label for="password">Пароль</Label>
            <Input id="password" v-model="password" type="password" placeholder="Введите пароль" :disabled="loading" />
          </div>

          <div class="space-y-2">
            <Label for="commission">Комиссия (%)</Label>
            <Input id="commission" v-model="commission" type="number" placeholder="0" :disabled="loading" />
          </div>

          <div class="space-y-2">
            <Label for="deadline">Срок оплаты</Label>
            <Input id="deadline" v-model="paymentDeadline" type="date" :disabled="loading" />
          </div>

          <div v-if="error" class="text-sm text-red-500">{{ error }}</div>

          <div class="flex gap-3">
            <Button type="button" variant="outline" @click="router.back()" :disabled="loading">
              Отмена
            </Button>
            <Button type="submit" :disabled="loading">
              {{ loading ? 'Создание...' : 'Создать' }}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
