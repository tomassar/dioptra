<script lang="ts">
  import { fetchTableData, type QueryResult } from "./api";

  interface Props {
    schema: string;
    table: string;
  }

  let { schema, table }: Props = $props();

  let result: QueryResult | null = $state(null);
  let meta: any = $state(null);
  let loading = $state(false);
  let error: string | null = $state(null);
  let page = $state(1);
  let pageSize = $state(50);

  async function load() {
    loading = true;
    error = null;
    try {
      const data = await fetchTableData(schema, table, page, pageSize);
      result = data.result;
      meta = data.meta;
    } catch (e: any) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    // Re-fetch when schema/table changes
    schema;
    table;
    page = 1;
    load();
  });

  function prevPage() {
    if (page > 1) {
      page--;
      load();
    }
  }

  function nextPage() {
    if (meta && page < meta.totalPages) {
      page++;
      load();
    }
  }

  function formatCell(val: any): string {
    if (val === null || val === undefined) return "NULL";
    if (typeof val === "object") return JSON.stringify(val);
    return String(val);
  }
</script>

<div class="table-view">
  <div class="table-header">
    <div class="header-left">
      <h2><span class="schema-label">{schema}.</span>{table}</h2>
      {#if meta}
        <span class="row-count">{meta.totalRows.toLocaleString()} rows</span>
      {/if}
    </div>
    <div class="header-right">
      <button class="action-btn">
        <svg
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          width="14"
          height="14"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
          />
        </svg>
        Filter
      </button>
      <button class="action-btn">
        <svg
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          width="14"
          height="14"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M8 9l4-4 4 4m0 6l-4 4-4-4"
          />
        </svg>
        Sort
      </button>
      <div class="header-divider"></div>
      <button
        class="btn-primary"
        onclick={() => alert("Add Insert functionality later!")}
      >
        <svg
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          width="14"
          height="14"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 4v16m8-8H4"
          />
        </svg>
        Insert
      </button>
    </div>
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
        <button onclick={nextPage} disabled={page >= meta.totalPages}
          >Next →</button
        >
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
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    background: var(--bg-primary);
    flex-shrink: 0;
  }

  .header-left {
    display: flex;
    align-items: baseline;
    gap: 12px;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .table-header h2 {
    font-size: 18px;
    font-weight: 400;
    font-family: var(--font-serif);
  }

  .schema-label {
    color: var(--text-muted);
    font-style: italic;
  }

  .row-count {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .action-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--text-secondary);
    font-size: 12px;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    transition: background 0.1s;
  }

  .action-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .action-btn:active {
    background: var(--bg-tertiary);
  }

  .header-divider {
    width: 1px;
    height: 24px;
    background: var(--border);
    margin: 0 8px;
  }

  .btn-primary {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    background: var(--accent);
    border: 1px solid var(--accent);
    border-radius: var(--radius);
    color: white;
    font-size: 12px;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    transition: background 0.1s;
  }

  .btn-primary:hover {
    background: var(--accent-dim);
    border-color: var(--accent-dim);
  }

  .btn-primary:active {
    background: var(--text-primary);
  }

  .error-banner {
    padding: 10px 20px;
    background: rgba(239, 68, 68, 0.1);
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
    background: var(--bg-primary);
  }

  table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0;
    font-size: 13px;
  }

  th {
    position: sticky;
    top: 0;
    background: var(--bg-tertiary);
    padding: 8px 16px;
    text-align: left;
    font-weight: 500;
    color: var(--text-secondary);
    border-bottom: 1px solid var(--border);
    border-right: 1px solid var(--border);
    white-space: nowrap;
    z-index: 10;
    letter-spacing: 0.04em;
    font-size: 11px;
    text-transform: uppercase;
  }

  td {
    padding: 12px 20px;
    border-bottom: 1px solid var(--border);
    border-right: 1px solid var(--border);
    max-width: 300px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--text-primary);
    transition: background 0.2s ease;
  }

  th:last-child,
  td:last-child {
    border-right: none;
  }

  .null-cell {
    color: var(--text-muted);
    font-style: italic;
  }

  .empty-row {
    text-align: center;
    color: var(--text-muted);
    padding: 40px;
    border-right: none;
  }

  tr:hover td {
    background: var(--bg-secondary);
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
