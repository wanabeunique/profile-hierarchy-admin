<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import Button from '@/components/ui/Button.vue'

const auth = useAuthStore()
const router = useRouter()

const login = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(login.value, password.value)
    router.push('/')
  } catch (e: unknown) {
    if (e && typeof e === 'object' && 'response' in e) {
      const axiosError = e as { response?: { data?: { message?: string } } }
      error.value = axiosError.response?.data?.message || 'Ошибка авторизации'
    } else {
      error.value = 'Ошибка соединения с сервером'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-[hsl(var(--background))] p-4">
    <Card class="w-full max-w-md">
      <CardHeader>
        <CardTitle>Вход в систему</CardTitle>
        <p class="text-sm text-[hsl(var(--muted-foreground))]">Введите данные для входа</p>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div class="space-y-2">
            <Label for="login">Логин / Email</Label>
            <Input
              id="login"
              v-model="login"
              placeholder="Введите логин или email"
              :disabled="loading"
            />
          </div>
          <div class="space-y-2">
            <Label for="password">Пароль</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              placeholder="Введите пароль"
              :disabled="loading"
            />
          </div>
          <div v-if="error" class="text-sm text-red-500">
            {{ error }}
          </div>
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? 'Вход...' : 'Войти' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
