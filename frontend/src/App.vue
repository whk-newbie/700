<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()

onMounted(() => {
  // 检查本地token
  authStore.checkAuth()
})
</script>

<style scoped>
#app {
  min-height: 100vh;
}
</style>
