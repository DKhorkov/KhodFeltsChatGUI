import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
    plugins: [vue()],
    server: {
        port: 5173,
        strictPort: true,
    },
    build: {
        rollupOptions: {
            external: [
                '/wails/runtime.js',
                '/wails/ipc.js',
            ],
        },
    },
})