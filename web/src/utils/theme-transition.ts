/**
 * 主题切换圆形扩散动画
 *
 * 原理：调用方先切换主题，本函数创建一个旧主题色的全屏遮罩，
 * 然后遮罩从全屏缩小到 0，制造"新主题从点击位置向外揭开"的视觉效果。
 *
 * 用法：
 *   1. 调用方先改变主题（修改 store 中的 darkMode / isDark）
 *   2. 立即调用本函数，传入点击位置和切换后的主题
 *
 * @param x 扩散起始点 X 坐标（相对于视口）
 * @param y 扩散起始点 Y 坐标（相对于视口）
 * @param toDark 切换后的主题是否为暗色（用于决定遮罩使用旧主题的背景色）
 */
export function animateThemeTransition(x: number, y: number, toDark: boolean): void {
  // 计算覆盖整个视口所需的最大半径
  const maxRadius = Math.hypot(
    Math.max(x, window.innerWidth - x),
    Math.max(y, window.innerHeight - y),
  )

  // 遮罩使用旧主题的背景色：切换到暗色 → 旧主题是浅色 → 遮罩用浅色背景
  const oldThemeBg = toDark ? '#f8fafc' : '#0f172a'

  // 创建遮罩层，初始覆盖整个屏幕
  const overlay = document.createElement('div')
  overlay.style.cssText = `
    position: fixed;
    inset: 0;
    z-index: 99999;
    pointer-events: none;
    clip-path: circle(${maxRadius}px at ${x}px ${y}px);
    background: ${oldThemeBg};
    transition: clip-path 1s cubic-bezier(0.4, 0, 0.2, 1);
  `
  document.body.appendChild(overlay)

  // 下一帧触发收缩动画：从全屏缩小到 0（消失于点击位置）
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      overlay.style.clipPath = `circle(0px at ${x}px ${y}px)`
    })
  })

  // 动画结束后清理遮罩
  overlay.addEventListener('transitionend', () => {
    overlay.remove()
  }, { once: true })

  // 兜底清理：1.1 秒后强制移除
  setTimeout(() => {
    if (overlay.parentNode) {
      overlay.remove()
    }
  }, 1100)
}