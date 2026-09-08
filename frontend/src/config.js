export const DEMO_USER_ID = import.meta.env.VITE_USER_ID || '00000000-0000-4000-8000-000000000001'
export const PRODUCT_IMAGE_BASE_URL = import.meta.env.VITE_PRODUCT_IMAGE_BASE_URL || 'http://localhost:9000/products'

export function productImageUrl(image) {
  if (!image) return ''
  if (image.url) return image.url
  if (!image.file_name) return ''
  return `${PRODUCT_IMAGE_BASE_URL}/${image.file_name}`
}
