<script setup lang="ts">
import SessionExpiredDialog from '@/components/SessionExpiredDialog.vue'
import { useAuth } from '@/composables/useAuth'

const { sessionExpired } = useAuth()
</script>

<template>
  <router-view />
  <SessionExpiredDialog v-if="sessionExpired" @close="sessionExpired = false" />
</template>
