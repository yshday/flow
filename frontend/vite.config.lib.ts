import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'
import dts from 'vite-plugin-dts'

// Library build configuration for @flow/issue-tracker package
export default defineConfig({
  plugins: [
    react(),
    dts({
      include: ['src/lib'],
      outDir: 'dist',
      tsconfigPath: './tsconfig.json',
      copyDtsFiles: true,
    }),
  ],
  css: {
    // CSS will be extracted to a separate file
  },
  build: {
    lib: {
      entry: resolve(__dirname, 'src/lib/index.ts'),
      name: 'FlowIssueTracker',
      formats: ['es', 'cjs'],
      fileName: (format) => `flow-issue-tracker.${format === 'es' ? 'mjs' : 'cjs'}`,
    },
    cssCodeSplit: false,
    rollupOptions: {
      // Externalize deps that shouldn't be bundled
      external: [
        'react',
        'react-dom',
        'react/jsx-runtime',
        'react/jsx-dev-runtime',
        '@tanstack/react-query',
        // Externalize all React-dependent libraries to avoid duplicate React
        /^react\/.*/,
        /^@tanstack\/.*/,
      ],
      output: {
        globals: {
          react: 'React',
          'react-dom': 'ReactDOM',
          'react/jsx-runtime': 'jsxRuntime',
          '@tanstack/react-query': 'ReactQuery',
        },
        assetFileNames: (assetInfo) => {
          if (assetInfo.name === 'style.css') return 'styles.css'
          return assetInfo.name || 'asset'
        },
      },
    },
    outDir: 'dist',
    sourcemap: true,
    minify: false,
  },
})
