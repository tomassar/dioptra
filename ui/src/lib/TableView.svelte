<script lang="ts">
  import { fetchTableData, type QueryResult } from './api'

  interface Props {
    schema: string
    table: string
  }

  let { schema, table }: Props = $props()

  let result: QueryResult | null = $state(null)
  let meta: any = $state(null)
  let loading = $state(false)
  let error: string | null = $state(null)
  let page = $state(1)
  let pageSize = $state(50)

  async function load() {
    loading = true
    error = null
    try {
      const data = await fetchTableData(schema, table, page, pageSize)
      result = data.result
      meta = data.meta
    } catch (e: any) {
      error = e.message
    } finally {
      loading = false
    }
  }

  $effect(() => {
    // Re-fetch when schema/table changes
    schema; table;
    page = 1
    load()
  })

  function prevPage() {
    if (page > 1) { page--; load() }
  }

  function nextPage() {
    if (meta && page < meta.totalPages) { page++; load() }
  }

  function formatCell(val: any): string {
    if (val === null || val === undefined) return 'NULL'
    if (typeof val === 'object') return JSON.stringify(val)
    return String(val)
  }
</script>

<div class="table-view">
  <div class="table-header">
    <h2>
      <span class="schema-label">{schema}.</span>{table}
    </h2>
    {#if meta}
      <span class="row-count">{meta.totalRows.toLocaleString()} rows</span>
    {/if}
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  {#if loading && !result}
    <div class="loading">Loading...</div>
  {:else if result}
    <div class="table-scroll">
      <table>
        <thead>
          <tr>
            {#each result.columns as col}
              <th>{col}</th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each result.rows as row}
            <tr>
              {#each row as cell}
                <td class:null-cell={cell === null || cell === undefined}>
                  {formatCell(cell)}
                </td>
              {/each}
            </tr>
          {:else}
            <tr>
              <td colspan={result.columns.length} class="empty-row">No rows</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    {#if meta && meta.totalPages > 1}
      <div class="pagination">
        <button onclick={prevPage} disabled={page <= 1}>← Prev</button>
        <span class="page-info">
          Page {page} of {meta.totalPages}
        </span>
        <button onclick={nextPage} disabled={page >= meta.totalPages}>Next →</button>
      </div>
    {/if}
  {/if}
</div>

<style>
  .table-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }

  .table-header {
    display: flex;
    align-items: baseline;
    gap: 12px;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
  }

  .table-header h2 {
    font-size: 16px;
    font-weight: 600;
    font-family: var(--font-mono);
  }

  .schema-label {
    color: var(--text-muted);
  }

  .row-count {
    font-size: 12px;
    color: var(--text-secondary);
    font-family: var(--font-mono);
  }

  .error-banner {
    padding: 10px 20px;
    background: rgba(248, 81, 73, 0.1);
    color: var(--error);
    font-size: 13px;
    border-bottom: 1px solid var(--border);
  }

  .loading {
    padding: 40px 20px;
    color: var(--text-muted);
    text-align: center;
  }

  .table-scroll {
    overflow: auto;
    flex: 1;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-family: var(--font-mono);
    font-size: 12px;
  }

  th {
    position: sticky;
    top: 0;
    background: var(--bg-tertiary);
    padding: 8px 12px;
    text-align: left;
    font-weight: 600;
    color: var(--text-secondary);
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  td {
    padding: 6px 12px;
    border-bottom: 1px solid var(--border);
    max-width: 300px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--text-primary);
  }

  .null-cell {
    color: var(--text-muted);
    font-style: italic;
  }

  .empty-row {
    text-align: center;
    color: var(--text-muted);
    padding: 40px;
  }

  tr:hover td {
    background: var(--bg-hover);
  }

  .pagination {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 12px;
    border-top: 1px solid var(--border);
    flex-shrink: 0;
  }

  .pagination button {
    padding: 6px 14px;
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    font-size: 12px;
  }

  .pagination button:hover:not(:disabled) {
    background: var(--bg-hover);
    border-color: var(--text-muted);
  }

  .pagination button:disabled {
    opacity: 0.4;
    cursor: default;
  }

  .page-info {
    font-size: 12px;
    color: var(--text-secondary);
    font-family: var(--font-mono);
  }
</style>
