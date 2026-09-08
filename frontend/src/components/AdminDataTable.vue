<script setup>
import { computed, ref, onMounted, watch, nextTick } from 'vue'
import DataTable from 'datatables.net-vue3'
import DataTablesCore from 'datatables.net-dt'

DataTable.use(DataTablesCore)

const props = defineProps({
  columns: {
    type: Array,
    required: true
  },
  data: {
    type: Array,
    default: () => []
  },
  ajax: {
    type: [String, Object, Function],
    default: null
  },
  options: {
    type: Object,
    default: () => ({})
  },
  serverSide: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['draw', 'xhr', 'ready', 'processing'])

const tableRef = ref(null)
const dtInstance = ref(null)

const normalizedColumns = computed(() => props.columns.map((column) => (
  column.data == null
    ? { defaultContent: '', ...column }
    : column
)))

const defaultOptions = {
  autoWidth: false,
  processing: true,
  searchDelay: 350,
  pageLength: 10,
  lengthMenu: [10, 25, 50, 100],
  language: {
    search: 'Search:',
    lengthMenu: 'Show _MENU_ entries',
    info: 'Showing _START_ to _END_ of _TOTAL_ entries',
    infoEmpty: 'No entries available',
    infoFiltered: '(filtered from _MAX_ total entries)',
    paginate: {
      first: 'First',
      last: 'Last',
      next: 'Next',
      previous: 'Previous'
    },
    emptyTable: 'No data available',
    zeroRecords: 'No matching records found'
  },
  layout: {
    topStart: 'pageLength',
    topEnd: 'search',
    bottomStart: 'info',
    bottomEnd: 'paging'
  },
  drawCallback: function() {
    emit('draw')
  }
}

const mergeOptions = () => {
  return {
    ...defaultOptions,
    ...props.options,
    serverSide: props.serverSide,
    ajax: props.ajax || undefined,
    data: !props.ajax ? props.data : undefined
  }
}

onMounted(async () => {
  await nextTick()
  if (tableRef.value) {
    dtInstance.value = tableRef.value.dt
    emit('ready', dtInstance.value)
  }
})

watch(() => props.loading, (isLoading) => {
  if (dtInstance.value) {
    if (isLoading) {
      dtInstance.value.processing(true)
    } else {
      dtInstance.value.processing(false)
    }
  }
})

const getInstance = () => dtInstance.value
const reload = (resetPaging = false) => {
  dtInstance.value?.ajax?.reload(null, resetPaging)
}

defineExpose({ getInstance, reload })
</script>

<template>
  <div class="relative">
    <div v-if="loading" class="absolute inset-0 bg-white bg-opacity-80 flex items-center justify-center z-10">
      <div class="flex items-center gap-2 text-gray-500">
        <svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span>Loading...</span>
      </div>
    </div>
    <DataTable
      ref="tableRef"
      :columns="normalizedColumns"
      :options="mergeOptions()"
      :ajax="ajax"
      :data="!ajax ? data : undefined"
      class="display nowrap"
    >
      <template
        v-for="(_, slotName) in $slots"
        #[slotName]="slotProps"
      >
        <slot :name="slotName" v-bind="slotProps || {}" />
      </template>
    </DataTable>
  </div>
</template>
