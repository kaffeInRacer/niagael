import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      // Auth Service :8085 (before generic /api/admin)
      '/api/auth': {
        target: 'http://localhost:8085',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/auth/, '/auth'),
        // Forward cookies
        configure: (proxy) => {
          proxy.on('proxyReq', (proxyReq) => {
            // Ensure cookies are forwarded
          })
        }
      },
      '/api/admin/users': {
        target: 'http://localhost:8085',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/admin\/users/, '/admin/users')
      },

      // Dynamic Pricing Service :8083 (specific routes first)
      '/api/admin/flash-sales': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/admin\/flash-sales/, '/admin/flash-sales')
      },
      '/api/admin/promos': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/admin\/promos/, '/admin/promos')
      },
      '/api/flash-sales': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/flash-sales/, '/flash-sales')
      },
      '/api/promos': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/promos/, '/promos')
      },

      // Product Service :8080
      '/api/products': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/products/, '/products')
      },
      '/api/categories': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/categories/, '/categories')
      },
      '/api/variants': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/variants/, '/variants')
      },
      '/api/admin': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/admin/, '/admin')
      },

      // Order Service :8082
      '/api/carts': {
        target: 'http://localhost:8082',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/carts/, '/carts')
      },
      '/api/orders': {
        target: 'http://localhost:8082',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/orders/, '/orders')
      },
      '/api/payments': {
        target: 'http://localhost:8082',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/payments/, '/payments')
      },
      '/api/addresses': {
        target: 'http://localhost:8082',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/addresses/, '/addresses')
      }
    }
  }
})
