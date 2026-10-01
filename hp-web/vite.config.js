import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  // echarts 从 core / charts / components / renderers 多入口按需引入时，
  // Vite 会把它们各自预构建成独立 entry，共享模块出现多份副本，
  // 导致 instanceof 校验失效（Invalid data provider）和重复注册（component exists）。
  // 这里强制去重，保证全局只有一份 echarts 内部模块。
  resolve: {
    dedupe: ['echarts'],
  },
  optimizeDeps: {
    include: ['echarts/core', 'echarts/charts', 'echarts/components', 'echarts/renderers'],
  },
})
