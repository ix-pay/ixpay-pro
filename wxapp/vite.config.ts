import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  const rollupOptions = {
    output: {
      entryFileNames: 'assets/[name].[hash].js',
      chunkFileNames: 'assets/[name].[hash].js',
      assetFileNames: 'assets/[name].[hash].[ext]',
      manualChunks: {
        // 分离 Vue 核心生态
        'vue-core': ['vue', 'vue-router'],
        // 分离 HTTP 请求库
        'http-client': ['axios'],
        // 分离微信 JS-SDK
        'weixin-js-sdk': ['weixin-js-sdk']
      }
    }
  }

  const base = "/wxapp"
  const root = './'
  const outDir = 'dist'

  const config = {
    base: base, // 编译后 js 导入的资源路径
    root: root, // index.html 文件所在位置
    publicDir: 'public', // 静态资源文件夹
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      host: '0.0.0.0',
      open: false,
      port: parseInt(env.VITE_CLI_PORT || '5173', 10),
      proxy: {
        // 把 key 的路径代理到 target 位置
        // detail: https://cli.vuejs.org/config/#devserver-proxy
        '/api': {
          // 需要代理的路径   例如 '/api'
          target: env.VITE_API_BASE_URL || 'http://localhost:8080', // 代理到目标路径
          changeOrigin: true,
          // 不要移除/api前缀，因为后端接口确实需要这个前缀
          rewrite: (path: string) => path
        }
      }
    },
    build: {
      minify: 'terser' as const, // 是否进行压缩，boolean | 'terser' | 'esbuild',默认使用 terser
      manifest: false, // 是否产出 manifest.json
      sourcemap: false, // 是否产出 sourcemap.json
      outDir: outDir, // 产出目录
      assetsDir: 'assets', // 静态资源目录
      assetsInlineLimit: 4096, // 小于 4KB 的资源内联
      cssCodeSplit: true, // 拆分 CSS
      chunkSizeWarningLimit: 1000, // 代码块大小警告阈值（1000 kB）
      terserOptions: {
        compress: {
          // 生产环境时移除 console
          drop_console: true,
          drop_debugger: true,
          pure_funcs: ['console.log', 'console.warn'] // 移除指定函数调用
        },
        output: {
          comments: false // 移除注释
        }
      },
      rollupOptions
    },
    optimizeDeps: {
      include: ['vue', 'vue-router', 'axios', 'weixin-js-sdk'] // 预构建依赖
    }
  }
  return config
})
