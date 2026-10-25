/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class', // 使用 class 模式控制暗黑模式
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        // 主色调 - 科技蓝
        primary: {
          DEFAULT: 'var(--primary-color)',
          light: 'var(--primary-light)',
          dark: 'var(--primary-dark)',
          glow: 'var(--primary-glow)',
        },
        // 成功色 - 翡翠绿
        success: {
          DEFAULT: 'var(--success-color)',
          light: 'var(--success-light)',
          dark: 'var(--success-dark)',
          glow: 'var(--success-glow)',
        },
        // 警告色 - 琥珀色
        warning: {
          DEFAULT: 'var(--warning-color)',
          light: 'var(--warning-light)',
          dark: 'var(--warning-dark)',
          glow: 'var(--warning-glow)',
        },
        // 危险色 - 玫瑰红
        danger: {
          DEFAULT: 'var(--danger-color)',
          light: 'var(--danger-light)',
          dark: 'var(--danger-dark)',
          glow: 'var(--danger-glow)',
        },
        // 信息色 - 青蓝色
        info: {
          DEFAULT: 'var(--info-color)',
          light: 'var(--info-light)',
          dark: 'var(--info-dark)',
          glow: 'var(--info-glow)',
        },
        // 中性色 - 背景
        bg: {
          primary: 'var(--bg-primary)',
          secondary: 'var(--bg-secondary)',
          tertiary: 'var(--bg-tertiary)',
        },
        // 中性色 - 文字
        text: {
          primary: 'var(--text-primary)',
          secondary: 'var(--text-secondary)',
          tertiary: 'var(--text-tertiary)',
          placeholder: 'var(--text-placeholder)',
        },
        // 中性色 - 边框
        border: {
          primary: 'var(--border-primary)',
          secondary: 'var(--border-secondary)',
          light: 'var(--border-light)',
        },
        // 布局组件
        sidebar: {
          bg: 'var(--sidebar-bg-solid)',
          text: 'var(--sidebar-text)',
          hover: 'var(--sidebar-hover-bg)',
          active: {
            bg: 'var(--sidebar-active-bg)',
            text: 'var(--sidebar-active-text)',
          },
        },
        header: {
          bg: 'var(--header-bg)',
          text: 'var(--header-text)',
          border: 'var(--header-border)',
        },
      },
      fontFamily: {
        sans: [
          'var(--font-sans)',
          'Inter',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Oxygen',
          'Ubuntu',
          'Cantarell',
          'Fira Sans',
          'Droid Sans',
          'Helvetica Neue',
          'Noto Sans SC',
          'PingFang SC',
          'Microsoft YaHei',
          'sans-serif',
        ],
        mono: [
          'var(--font-mono)',
          'JetBrains Mono',
          'Fira Code',
          'monospace',
        ],
        chinese: [
          'var(--font-chinese)',
          'Noto Sans SC',
          'PingFang SC',
          'Microsoft YaHei',
          'sans-serif',
        ],
      },
      fontSize: {
        xs: 'var(--text-xs)',     // 12px
        sm: 'var(--text-sm)',      // 14px
        base: 'var(--text-base)',  // 16px
        lg: 'var(--text-lg)',      // 18px
        xl: 'var(--text-xl)',      // 20px
        '2xl': 'var(--text-2xl)',  // 24px
        '3xl': 'var(--text-3xl)',  // 30px
        '4xl': 'var(--text-4xl)',  // 36px
      },
      spacing: {
        // 标准 Tailwind 间距值 (0-96)
        px: '1px',
        0: '0',
        0.5: '0.125rem',
        1: '0.25rem',
        1.5: '0.375rem',
        2: '0.5rem',
        2.5: '0.625rem',
        3: '0.75rem',
        3.5: '0.875rem',
        4: '1rem',
        5: '1.25rem',
        6: '1.5rem',
        7: '1.75rem',
        8: '2rem',
        9: '2.25rem',
        10: '2.5rem',
        11: '2.75rem',
        12: '3rem',
        14: '3.5rem',
        16: '4rem',
        20: '5rem',
        24: '6rem',
        28: '7rem',
        32: '8rem',
        36: '9rem',
        40: '10rem',
        44: '11rem',
        48: '12rem',
        52: '13rem',
        56: '14rem',
        60: '15rem',
        64: '16rem',
        72: '18rem',
        80: '20rem',
        96: '24rem',
        // 自定义间距值（使用 CSS 变量）
        'space-1': 'var(--space-1)',   // 4px
        'space-2': 'var(--space-2)',   // 8px
        'space-3': 'var(--space-3)',   // 12px
        'space-4': 'var(--space-4)',   // 16px
        'space-5': 'var(--space-5)',   // 20px
        'space-6': 'var(--space-6)',   // 24px
        'space-8': 'var(--space-8)',   // 32px
        'space-10': 'var(--space-10)', // 40px
        'space-12': 'var(--space-12)', // 48px
        'space-16': 'var(--space-16)', // 64px
        'space-20': 'var(--space-20)', // 80px
        'space-24': 'var(--space-24)', // 96px
      },
      borderRadius: {
        none: '0',
        sm: 'var(--radius-sm)',    // 6px
        md: 'var(--radius-md)',    // 8px
        lg: 'var(--radius-lg)',    // 12px
        xl: 'var(--radius-xl)',    // 16px
        '2xl': 'var(--radius-2xl)', // 24px
        full: 'var(--radius-full)', // 9999px
      },
      boxShadow: {
        sm: 'var(--shadow-sm)',
        md: 'var(--shadow-md)',
        lg: 'var(--shadow-lg)',
        xl: 'var(--shadow-xl)',
        '2xl': 'var(--shadow-2xl)',
        'primary-glow': 'var(--shadow-primary-glow)',
        'success-glow': 'var(--shadow-success-glow)',
        'warning-glow': 'var(--shadow-warning-glow)',
        'danger-glow': 'var(--shadow-danger-glow)',
        'info-glow': 'var(--shadow-info-glow)',
      },
      transitionDuration: {
        fast: 'var(--duration-fast)',    // 150ms
        normal: 'var(--duration-normal)', // 250ms
        slow: 'var(--duration-slow)',     // 350ms
        slower: 'var(--duration-slower)', // 500ms
      },
      transitionTimingFunction: {
        'in-out': 'var(--ease-in-out)',
        out: 'var(--ease-out)',
        in: 'var(--ease-in)',
        bounce: 'var(--ease-bounce)',
      },
      backgroundImage: {
        'gradient-primary': 'var(--primary-gradient)',
        'gradient-success': 'var(--success-gradient)',
        'gradient-warning': 'var(--warning-gradient)',
        'gradient-danger': 'var(--danger-gradient)',
        'gradient-info': 'var(--info-gradient)',
        'gradient-bg': 'var(--bg-gradient)',
        'gradient-sidebar': 'var(--sidebar-bg)',
      },
      backdropBlur: {
        glass: '10px',
        'glass-dark': '12px',
      },
      keyframes: {
        'fade-in': {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        'fade-in-up': {
          '0%': {
            opacity: '0',
            transform: 'translateY(20px)',
          },
          '100%': {
            opacity: '1',
            transform: 'translateY(0)',
          },
        },
        'fade-in-down': {
          '0%': {
            opacity: '0',
            transform: 'translateY(-20px)',
          },
          '100%': {
            opacity: '1',
            transform: 'translateY(0)',
          },
        },
        'scale-in': {
          '0%': {
            opacity: '0',
            transform: 'scale(0.95)',
          },
          '100%': {
            opacity: '1',
            transform: 'scale(1)',
          },
        },
        'slide-in-right': {
          '0%': {
            opacity: '0',
            transform: 'translateX(20px)',
          },
          '100%': {
            opacity: '1',
            transform: 'translateX(0)',
          },
        },
        'slide-in-left': {
          '0%': {
            opacity: '0',
            transform: 'translateX(-20px)',
          },
          '100%': {
            opacity: '1',
            transform: 'translateX(0)',
          },
        },
        shimmer: {
          '0%': { backgroundPosition: '-1000px 0' },
          '100%': { backgroundPosition: '1000px 0' },
        },
        'pulse-glow': {
          '0%, 100%': { boxShadow: '0 0 20px var(--primary-glow)' },
          '50%': { boxShadow: '0 0 40px var(--primary-glow)' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-10px)' },
        },
      },
      animation: {
        'fade-in': 'fade-in var(--duration-normal) var(--ease-out)',
        'fade-in-up': 'fade-in-up var(--duration-slow) var(--ease-out)',
        'fade-in-down': 'fade-in-down var(--duration-slow) var(--ease-out)',
        'scale-in': 'scale-in var(--duration-normal) var(--ease-out)',
        'slide-in-right': 'slide-in-right var(--duration-normal) var(--ease-out)',
        'slide-in-left': 'slide-in-left var(--duration-normal) var(--ease-out)',
        shimmer: 'shimmer 2s linear infinite',
        'pulse-glow': 'pulse-glow 2s ease-in-out infinite',
        float: 'float 3s ease-in-out infinite',
        'spin-slow': 'spin 2s linear infinite',
      },
    },
  },
  plugins: [],
  corePlugins: {
    // 禁用与 Element Plus 冲突的预检
    preflight: false,
  },
}
