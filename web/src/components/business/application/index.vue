<template>
  <error-preview
    v-if="showError && errorInfo"
    :error-data="errorInfo"
    @close="handleClose"
    @confirm="handleConfirm"
  />
</template>

<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import { emitter } from '@/utils/bus'
import ErrorPreview from '@/components/business/errorPreview/index.vue'

defineOptions({ name: 'ApplicationIndex' })

interface ErrorData {
  code: number | string
  message?: string
  fn?: (code: string | number) => void
}

const showError = ref(false)
const errorInfo = ref<ErrorData | null>(null)
let cb: ((code: string | number) => void) | null = null

const showErrorDialog = (data: ErrorData) => {
  // 这玩意同时只允许存在一个
  if (showError.value) return

  errorInfo.value = data
  showError.value = true
  cb = data?.fn || null
}

const handleClose = () => {
  showError.value = false
  errorInfo.value = null
  cb = null
}

const handleConfirm = (code: string | number) => {
  if (cb) cb(code)
  handleClose()
}

emitter.on('show-error', showErrorDialog)

onUnmounted(() => {
  emitter.off('show-error', showErrorDialog)
})
</script>
