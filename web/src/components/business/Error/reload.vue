<template>
  <div />
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

defineOptions({
  name: 'ErrorReload',
})

const router = useRouter()
router.go(-1)
</script>
