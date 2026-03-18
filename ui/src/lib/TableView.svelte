<script lang="ts">
  import { untrack } from "svelte";
  import { fetchTableData, insertRow, type QueryResult } from "./api";

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

  let sortCol = $state("");
  let sortDir = $state("ASC");
  let filterCol = $state("");
  let filterVal = $state("");

  let showFilter = $state(false);
  let showSort = $state(false);
  let showInsert = $state(false);
  let insertData: Record<string, string> = $state({});

  async function load() {
    loading = true;
    error = null;
    try {
      const data = await fetchTableData(
        schema,
        table,
        page,
        pageSize,
        sortCol,
        sortDir,
        filterCol,
        filterVal,
      );
      result = data.result;
      meta = data.meta;
    } catch (e: any) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  async function handleInsert() {
    try {
      await insertRow(schema, table, insertData);
      showInsert = false;
      insertData = {};
      load();
    } catch (e: any) {
      error = e.message;
    }
  }

  $effect(() => {
    schema;
    table;

    untrack(() => {
      page = 1;
      sortCol = "";
      sortDir = "ASC";
      filterCol = "";
      filterVal = "";
      showFilter = false;
      showSort = false;
      showInsert = false;
      insertData = {};
      load();
    });
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
      <button
        class="action-btn"
        class:active={showFilter}
        onclick={() => {
          showFilter = !showFilter;
          showSort = false;
          showInsert = false;
        }}
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
            d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
          />
        </svg>
        Filter
      </button>
      <button
        class="action-btn"
        class:active={showSort}
        onclick={() => {
          showSort = !showSort;
          showFilter = false;
          showInsert = false;
        }}
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
            d="M8 9l4-4 4 4m0 6l-4 4-4-4"
          />
        </svg>
        Sort
      </button>
      <div class="header-divider"></div>
      <button
        class="btn-primary"
        class:active={showInsert}
        onclick={() => {
          showInsert = !showInsert;
          showFilter = false;
          showSort = false;
          insertData = {};
        }}
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

  {#if showFilter || showSort}
    <div class="toolbar-panel">
      {#if showFilter}
        <div class="toolbar-group">
          <span class="toolbar-label">Where</span>
          <select bind:value={filterCol}>
            <option value="" disabled selected>Select column</option>
            {#each result?.columns || [] as col}
              <option value={col}>{col}</option>
            {/each}
          </select>
          <span class="operator">contains</span>
          <input
            type="text"
            bind:value={filterVal}
            placeholder="Value..."
            onkeydown={(e) => {
              if (e.key === "Enter") {
                page = 1;
                load();
              }
            }}
          />
          <button
            class="action-btn"
            disabled={!filterCol || !filterVal}
            onclick={() => {
              page = 1;
              load();
            }}>Apply</button
          >
          {#if filterCol || filterVal}
            <button
              class="ghost-btn"
              onclick={() => {
                filterCol = "";
                filterVal = "";
                page = 1;
                load();
              }}>Clear</button
            >
          {/if}
        </div>
      {/if}

      {#if showSort}
        <div class="toolbar-group">
          <span class="toolbar-label">Order by</span>
          <select bind:value={sortCol}>
            <option value="" disabled selected>Select column</option>
            {#each result?.columns || [] as col}
              <option value={col}>{col}</option>
            {/each}
          </select>
          <select bind:value={sortDir}>
            <option value="ASC">Ascending</option>
            <option value="DESC">Descending</option>
          </select>
          <button
            class="action-btn"
            disabled={!sortCol}
            onclick={() => {
              page = 1;
              load();
            }}>Apply</button
          >
          {#if sortCol}
            <button
              class="ghost-btn"
              onclick={() => {
                sortCol = "";
                sortDir = "ASC";
                page = 1;
                load();
              }}>Clear</button
            >
          {/if}
        </div>
      {/if}
    </div>
  {/if}

  {#if showInsert}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="drawer-overlay" onclick={() => (showInsert = false)}></div>
    <div class="drawer">
      <div class="drawer-header">
        <h3>Insert row</h3>
        <button class="ghost-btn" onclick={() => (showInsert = false)}>✕</button
        >
      </div>
      <div class="drawer-content">
        {#each result?.columns || [] as col}
          <div class="field-group">
            <label for="insert-{col}">{col}</label>
            <input
              id="insert-{col}"
              type="text"
              bind:value={insertData[col]}
              placeholder="Leave empty for default"
            />
          </div>
        {/each}
      </div>
      <div class="drawer-footer">
        <button class="ghost-btn" onclick={() => (showInsert = false)}
          >Cancel</button
        >
        <button class="btn-primary" onclick={handleInsert}>Save Record</button>
      </div>
    </div>
  {/if}

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

  .action-btn.active {
    background: var(--bg-tertiary);
    border-color: var(--text-primary);
    color: var(--text-primary);
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

  .btn-primary:active:not(:disabled) {
    background: var(--text-primary);
  }

  .action-btn:disabled,
  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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

  .toolbar-panel {
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border);
    padding: 12px 16px;
    flex-shrink: 0;
  }

  .toolbar-group {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .toolbar-label {
    font-size: 13px;
    font-weight: 500;
    margin-right: 8px;
    color: var(--text-primary);
  }

  .toolbar-group input,
  .toolbar-group select {
    background: var(--bg-primary);
    border: 1px solid var(--border);
    color: var(--text-primary);
    padding: 6px 12px;
    border-radius: var(--radius);
    font-size: 12px;
    outline: none;
  }

  .toolbar-group input:focus,
  .toolbar-group select:focus {
    border-color: var(--text-secondary);
  }

  .operator {
    font-size: 13px;
    color: var(--text-secondary);
    margin: 0 4px;
    font-style: italic;
  }

  .ghost-btn {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    padding: 6px 12px;
    border-radius: var(--radius);
  }

  .ghost-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .drawer-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.1);
    z-index: 100;
  }

  .drawer {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    width: 400px;
    background: var(--bg-primary);
    border-left: 1px solid var(--border);
    box-shadow: var(--shadow-float);
    z-index: 101;
    display: flex;
    flex-direction: column;
    animation: slideIn 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes slideIn {
    from {
      transform: translateX(100%);
    }
    to {
      transform: translateX(0);
    }
  }

  .drawer-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 24px;
    border-bottom: 1px solid var(--border);
  }

  .drawer-header h3 {
    font-size: 16px;
    font-weight: 400;
    font-family: var(--font-serif);
  }

  .drawer-content {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .drawer-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
    padding: 16px 24px;
    border-top: 1px solid var(--border);
    background: var(--bg-tertiary);
  }

  .field-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .field-group label {
    font-size: 12px;
    font-weight: 500;
    color: var(--text-secondary);
  }

  .field-group input {
    background: var(--bg-primary);
    border: 1px solid var(--border);
    color: var(--text-primary);
    padding: 8px 12px;
    border-radius: var(--radius);
    font-size: 13px;
    outline: none;
    transition: border-color 0.1s;
  }

  .field-group input:focus {
    border-color: var(--text-primary);
  }
</style>
