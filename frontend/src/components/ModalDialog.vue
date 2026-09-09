<template>
  <TransitionRoot :show="show" appear as="template">
    <Dialog as="div" class="relative z-50" @close="$emit('close')">
      <TransitionChild
        as="template"
        enter="duration-200 ease-out"
        enter-from="opacity-0"
        enter-to="opacity-100"
        leave="duration-150 ease-in"
        leave-from="opacity-100"
        leave-to="opacity-0"
      >
        <div class="fixed inset-0 bg-black bg-opacity-50" />
      </TransitionChild>

      <div class="fixed inset-0 overflow-y-auto">
        <div class="flex min-h-full items-center justify-center p-4">
          <TransitionChild
            as="template"
            appear
            enter="duration-200 ease-out"
            enter-from="opacity-0 scale-95"
            enter-to="opacity-100 scale-100"
            leave="duration-150 ease-in"
            leave-from="opacity-100 scale-100"
            leave-to="opacity-0 scale-95"
          >
            <DialogPanel
              class="bg-white rounded-lg shadow-xl w-full transform transition-all"
              :class="sizeClass"
            >
              <div class="px-6 py-4 border-b flex justify-between items-center">
                <DialogTitle as="h3" class="text-lg font-semibold text-gray-800" :id="ariaId">
                  <slot name="title">{{ title }}</slot>
                </DialogTitle>
                <button
                  type="button"
                  class="text-gray-400 hover:text-gray-600"
                  aria-label="Close dialog"
                  @click="$emit('close')"
                >
                  <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div :class="bodyClass">
                <slot />
              </div>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup>
import { computed } from 'vue'
import {
  Dialog,
  DialogPanel,
  DialogTitle,
  TransitionRoot,
  TransitionChild
} from '@headlessui/vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  title: { type: String, default: '' },
  size: { type: String, default: 'md' },
  ariaId: { type: String, default: 'modal-title' },
  scrollBody: { type: Boolean, default: false }
})

defineEmits(['close'])

const sizeClass = computed(() => ({
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-2xl',
  xl: 'max-w-3xl'
}[props.size] || 'max-w-md'))

const bodyClass = computed(() => props.scrollBody ? 'px-6 py-4 max-h-[70vh] overflow-y-auto' : 'px-6 py-4')
</script>
