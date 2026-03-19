export interface TableInfo {
  schema: string
  name: string
  rows: number
}

export interface QueryResult {
  columns: string[]
  rows: any[][]
}

interface ApiResponse<T> {
  data?: T
  error?: string
  meta?: any
}

async function request<T>(url: string, opts?: RequestInit): Promise<ApiResponse<T>> {
  const res = await fetch(url, opts)
  return res.json()
}

export async function fetchSchemas(): Promise<string[]> {
  const res = await request<string[]>('/api/schemas')
  if (res.error) throw new Error(res.error)
  return res.data ?? []
}

export async function fetchTables(schema: string): Promise<TableInfo[]> {
  const res = await request<TableInfo[]>(`/api/tables?schema=${encodeURIComponent(schema)}`)
  if (res.error) throw new Error(res.error)
  return res.data ?? []
}

export async function fetchTableData(
  schema: string,
  table: string,
  page: number = 1,
  pageSize: number = 50,
  sortCol: string = '',
  sortDir: string = '',
  filterCol: string = '',
  filterVal: string = ''
): Promise<{ result: QueryResult; meta: any }> {
  const params = new URLSearchParams()
  params.set('page', String(page))
  params.set('pageSize', String(pageSize))
  if (sortCol) params.set('sortCol', sortCol)
  if (sortDir) params.set('sortDir', sortDir)
  if (filterCol) params.set('filterCol', filterCol)
  if (filterVal) params.set('filterVal', filterVal)

  const res = await request<QueryResult>(
    `/api/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}?${params.toString()}`
  )
  if (res.error) throw new Error(res.error)
  return { result: res.data!, meta: res.meta }
}

export async function insertRow(schema: string, table: string, data: Record<string, string>): Promise<void> {
  const res = await request<void>(`/api/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}/insert`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (res.error) throw new Error(res.error)
}

export async function updateRow(schema: string, table: string, pkValues: Record<string, any>, updates: Record<string, any>): Promise<void> {
  const res = await request<void>(`/api/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}/update`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pkValues, updates }),
  })
  if (res.error) throw new Error(res.error)
}

export async function runQuery(sql: string): Promise<QueryResult> {
  const res = await request<QueryResult>('/api/query', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ sql }),
  })
  if (res.error) throw new Error(res.error)
  return res.data!
}

export async function fetchStatus(): Promise<{ readOnly: boolean }> {
  const res = await request<{ readOnly: boolean }>('/api/status')
  return res.data!
}
