import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import Components from 'unplugin-vue-components/vite'
import AutoImport from 'unplugin-auto-import/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  const rollupOptions = {
    output: {
      entryFileNames: 'assets/087AC4D233B64EB0[name].[hash].js',
      chunkFileNames: 'assets/087AC4D233B64EB0[name].[hash].js',
      assetFileNames: 'assets/087AC4D233B64EB0[name].[hash].[ext]',
      manualChunks: {
        // 分离 Vue 核心生态
        'vue-core': ['vue', 'vue-router', 'pinia'],
        // 分离 HTTP 请求库
        'http-client': ['axios'],
        // 分离 UI 组件库
        'element-plus': ['element-plus'],
        // 分离工具库
        'vueuse': ['@vueuse/core'],
        // 分离图标库
        'element-icons': ['@element-plus/icons-vue']
      }
    }
  }

  const base = "/"
  const root = "./"
  const outDir = "dist"

  const config = {
    base: base, // 编译后 js 导入的资源路径
    root: root, // index.html 文件所在位置
    publicDir: 'public', // 静态资源文件夹
    plugins: [
      vue(),
      vueDevTools(),
      // Element Plus 按需引入
      Components({
        resolvers: [ElementPlusResolver()],
        dts: 'src/components.d.ts', // 自动生成类型声明
      }),
      AutoImport({
        resolvers: [ElementPlusResolver()],
        dts: 'src/auto-imports.d.ts', // 自动生成类型声明
        imports: [
          'vue',
          'vue-router',
          'pinia',
        ],
      }),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        '@/app': fileURLToPath(new URL('./src/app', import.meta.url)),
        '@/assets': fileURLToPath(new URL('./src/assets', import.meta.url)),
        '@/components': fileURLToPath(new URL('./src/components', import.meta.url)),
        '@/core': fileURLToPath(new URL('./src/core', import.meta.url)),
        '@/features': fileURLToPath(new URL('./src/features', import.meta.url)),
        '@/hooks': fileURLToPath(new URL('./src/hooks', import.meta.url)),
        '@/services': fileURLToPath(new URL('./src/services', import.meta.url)),
        '@/stores': fileURLToPath(new URL('./src/stores', import.meta.url)),
        '@/utils': fileURLToPath(new URL('./src/utils', import.meta.url)),
        '@/types': fileURLToPath(new URL('./src/types', import.meta.url))
      },
    },
    server: {
      open: false,
      port: parseInt(env.VITE_CLI_PORT || '8080', 10),
      proxy: {
        // 把key的路径代理到target位置
        // detail: https://cli.vuejs.org/config/#devserver-proxy
        [env.VITE_BASE_API as string]: {
          // 需要代理的路径   例如 '/api'
          target: `${env.VITE_BASE_PATH}:${env.VITE_SERVER_PORT}/`, // 代理到 目标路径
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
          //生产环境时移除console
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
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: `/* 全局样式重置 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  margin: 0;
  padding: 0;
  display: block;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

html, body {
  height: 100%;
  overflow: hidden;
}

#app {
  height: 100%;
}` // 全局导入 SCSS 变量
        }
      }
    },
    optimizeDeps: {
      include: ['vue', 'vue-router', 'pinia', 'axios'] // 预构建依赖（移除 element-plus，改用按需引入）
    }
  }
  return config
})

