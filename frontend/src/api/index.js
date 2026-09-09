import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: true
})

let refreshRequest = null

api.interceptors.response.use(
  response => response,
  async (error) => {
    const request = error.config
    const isAuthRequest = request?.url?.startsWith('/auth/')

    if (error.response?.status !== 401 || request?._retry || isAuthRequest) {
      return Promise.reject(error)
    }

    request._retry = true
    try {
      refreshRequest ||= axios.post('/api/auth/refresh', {}, { withCredentials: true })
        .then(response => {
          return response.data.access_token
        })
        .finally(() => {
          refreshRequest = null
        })

      await refreshRequest
      return api(request)
    } catch (refreshError) {
      if (!window.location.pathname.startsWith('/login')) {
        const redirect = encodeURIComponent(window.location.pathname + window.location.search)
        window.location.assign(`/login?redirect=${redirect}`)
      }
      return Promise.reject(refreshError)
    }
  }
)

export const authApi = {
  register: (data) => api.post('/auth/register', data),
  login: (data) => api.post('/auth/login', data),
  refresh: () => api.post('/auth/refresh', {}, { withCredentials: true }),
  logout: () => api.post('/auth/logout'),
  me: () => api.get('/auth/me')
}

export const userAdminApi = {
  list: (params) => api.get('/admin/users', { params }),
  create: (data) => api.post('/admin/users', data),
  updateRole: (id, role) => api.patch(`/admin/users/${id}/role`, { role }),
  updateStatus: (id, isActive) => api.patch(`/admin/users/${id}/status`, { is_active: isActive }),
  updateEmail: (id, email) => api.put(`/admin/users/${id}/email`, { email }),
  delete: (id) => api.delete(`/admin/users/${id}`)
}

export const rbacApi = {
  listPolicies: () => api.get('/admin/rbac/policies'),
  getResources: () => api.get('/admin/rbac/resources'),
  addPolicy: (data) => api.post('/admin/rbac/policies', data),
  deletePolicy: (data) => api.delete('/admin/rbac/policies', { data })
}

export const productApi = {
  list: (params) => api.get('/products', { params }),
  getById: (id) => api.get(`/products/${id}`),
  getBySlug: (slug) => api.get(`/products/slug/${slug}`)
}

export const categoryApi = {
  list: () => api.get('/categories'),
  getById: (id) => api.get(`/categories/${id}`)
}

export const flashSaleApi = {
  list: (params) => api.get('/flash-sales', { params }),
  getById: (id) => api.get(`/flash-sales/${id}`),
  adminList: (params) => api.get('/admin/flash-sales', { params }),
  adminGetById: (id) => api.get(`/admin/flash-sales/${id}`),
  createBulk: (data) => api.post('/admin/flash-sales/bulk', data)
}

export const promoApi = {
  list: (params) => api.get('/promos', { params }),
  getById: (id) => api.get(`/promos/${id}`),
  apply: (code, userId, totalAmount) => api.post(`/promos/apply/${code}/${userId}`, { total_amount: totalAmount })
}

export const orderApi = {
  create: (data) => api.post('/orders', data),
  getById: (id) => api.get(`/orders/${id}`),
  list: (params) => api.get('/orders', { params }),
  listByUser: (userId, params) => api.get(`/orders/user/${userId}`, { params })
}

export const cartApi = {
  getByUserId: (userId) => api.get(`/carts/user/${userId}`),
  addItem: (data) => api.post('/carts/items', data),
  updateItem: (id, data) => api.put(`/carts/items/${id}`, data),
  deleteItem: (id, userId) => api.delete(`/carts/items/${id}`, { params: { user_id: userId } }),
  clear: (userId) => api.delete(`/carts/user/${userId}`)
}

export const paymentApi = {
  create: (orderId, data) => api.post(`/payments/${orderId}`, data)
}

export const addressApi = {
  list: (params) => api.get('/addresses', { params }),
  getById: (id) => api.get(`/addresses/${id}`),
  create: (data) => api.post('/addresses', data)
}

export const adminApi = {
  categories: {
    list: (params) => api.get('/admin/categories', { params }),
    create: (data) => api.post('/admin/categories', data),
    update: (id, data) => api.put(`/admin/categories/${id}`, data),
    delete: (id) => api.delete(`/admin/categories/${id}`)
  },
  products: {
    list: (params) => api.get('/admin/products', { params }),
    create: (data) => api.post('/admin/products', data),
    update: (id, data) => api.put(`/admin/products/${id}`, data),
    delete: (id) => api.delete(`/admin/products/${id}`),
    uploadImage: (id, formData) => api.post(`/admin/products/${id}/images`, formData, { headers: { 'Content-Type': 'multipart/form-data' } }),
    deleteImage: (productId, imageId) => api.delete(`/admin/products/${productId}/images/${imageId}`)
  }
}

export default api
