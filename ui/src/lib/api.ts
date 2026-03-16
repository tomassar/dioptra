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
  pageSize: number = 50
): Promise<{ result: QueryResult; meta: any }> {
  const res = await request<QueryResult>(
    `/api/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}?page=${page}&pageSize=${pageSize}`
  )
  if (res.error) throw new Error(res.error)
  return { result: res.data!, meta: res.meta }
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
