<script setup lang="ts">
import { ref } from 'vue'
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
const profileId = route.params.id as string

const number = ref('')
const amount = ref('')
const loading = ref(false)
const error = ref('')

async function handleSubmit() {
  error.value = ''
  loading.value = true

  try {
    await api.post(`/profiles/${profileId}/orders`, {
      number: number.value,
      amount: parseFloat(amount.value) || 0,
    })
    router.push(`/profiles/${profileId}`)
  } catch (e: unknown) {
    if (e && typeof e === 'object' && 'response' in e) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Ошибка создания заказа'
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
      <h1 class="text-2xl font-bold text-[hsl(var(--foreground))]">Создание заказа</h1>
      <p class="text-sm text-[hsl(var(--muted-foreground))]">Заполните данные нового заказа</p>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>Данные заказа</CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div class="space-y-2">
            <Label for="number">Номер заказа</Label>
            <Input id="number" v-model="number" placeholder="Введите номер заказа" :disabled="loading" />
          </div>

          <div class="space-y-2">
            <Label for="amount">Сумма</Label>
            <Input id="amount" v-model="amount" type="number" placeholder="0.00" :disabled="loading" />
          </div>

          <div v-if="error" class="text-sm text-red-500">{{ error }}</div>

          <div class="flex gap-3">
            <Button type="button" variant="outline" @click="router.back()" :disabled="loading">
              Отмена
            </Button>
            <Button type="submit" :disabled="loading">
              {{ loading ? 'Создание...' : 'Создать заказ' }}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
