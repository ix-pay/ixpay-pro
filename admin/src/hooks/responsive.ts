import { ref, onMounted, onUnmounted } from 'vue'

export default function useResponsive(init = false) {
  const device = ref<string>('desktop')
  const screenWidth = ref<number>(window.innerWidth)

  const updateDevice = () => {
    screenWidth.value = window.innerWidth
    if (screenWidth.value < 768) {
      device.value = 'mobile'
    } else if (screenWidth.value < 1024) {
      device.value = 'tablet'
    } else {
      device.value = 'desktop'
    }
  }

  onMounted(() => {
    if (init) {
      updateDevice()
    }
    window.addEventListener('resize', updateDevice)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', updateDevice)
  })

  return { device, screenWidth }
}
