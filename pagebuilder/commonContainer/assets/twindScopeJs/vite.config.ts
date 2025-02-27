import { defineConfig } from 'vite'
import cp from 'vite-plugin-cp'

export default defineConfig({
  build: {
    lib: {
      entry: './lib/main.ts',
      name: 'common-container-scope',
      fileName: 'common-container-scope',
    },
    terserOptions: {
      compress: {
        drop_console: false,
        drop_debugger: false,
      },
    },
  },
  plugins: [
    cp({
      targets: [
        {
          src: './dist/common-container-scope.js',
          dest: '../js',
        },
      ],
    }),
  ],
})
