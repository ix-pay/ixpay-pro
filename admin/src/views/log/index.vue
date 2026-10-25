<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <router-view />
  </div>
</template>

<script setup lang="ts">
defineOptions({
  name: 'LogManagement',
})
</script>
