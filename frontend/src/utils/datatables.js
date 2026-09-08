export function createServerSideAjax({
  fetchPage,
  orderColumns = {},
  defaultOrderBy = 'created_at',
  defaultOrderDir = 'desc',
  getFilters = () => ({}),
  onError = () => {}
}) {
  return async (request, callback) => {
    const pageSize = request.length > 0 ? request.length : 10
    const order = request.order?.[0]
    const orderBy = orderColumns[order?.column] || defaultOrderBy
    const orderDir = ['asc', 'desc'].includes(order?.dir) ? order.dir : defaultOrderDir

    const params = {
      page: Math.floor(request.start / pageSize),
      page_size: pageSize,
      search: request.search?.value?.trim() || undefined,
      order_by: orderBy,
      order_dir: orderDir,
      ...getFilters()
    }

    Object.keys(params).forEach((key) => {
      if (params[key] === '' || params[key] == null) delete params[key]
    })

    try {
      const response = await fetchPage(params)
      const count = Number(response.data.count || 0)

      callback({
        draw: request.draw,
        data: response.data.data || [],
        recordsTotal: count,
        recordsFiltered: count
      })
    } catch (error) {
      onError(error)
      callback({
        draw: request.draw,
        data: [],
        recordsTotal: 0,
        recordsFiltered: 0
      })
    }
  }
}
