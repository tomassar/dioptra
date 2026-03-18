<script lang="ts">
  import { fetchSchemas, fetchTables, type TableInfo } from "./api";

  interface Props {
    onselect: (schema: string, table: string) => void;
  }

  let { onselect }: Props = $props();

  let schemas: string[] = $state([]);
  let tablesBySchema: Record<string, TableInfo[]> = $state({});
  let expandedSchema: string | null = $state(null);
  let activeTable: string | null = $state(null);
  let loading = $state(true);
  let error: string | null = $state(null);

  async function loadSchemas() {
    try {
      schemas = await fetchSchemas();
      if (schemas.length > 0) {
        await toggleSchema(schemas.includes("public") ? "public" : schemas[0]);
      }
    } catch (e: any) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  async function toggleSchema(schema: string) {
    if (expandedSchema === schema) {
      expandedSchema = null;
      return;
    }
    expandedSchema = schema;
    if (!tablesBySchema[schema]) {
      tablesBySchema[schema] = await fetchTables(schema);
    }
  }

  function selectTable(schema: string, table: string) {
    activeTable = `${schema}.${table}`;
    onselect(schema, table);
  }

  loadSchemas();
</script>

<nav class="sidebar">
  <div class="sidebar-header">
    <svg
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
    >
      <ellipse cx="12" cy="5" rx="9" ry="3" />
      <path d="M3 5V19A9 3 0 0 0 21 19V5" />
      <path d="M3 12A9 3 0 0 0 21 12" />
    </svg>
    <span class="sidebar-title">Tables</span>
  </div>

  {#if loading}
    <div class="sidebar-empty">Loading...</div>
  {:else if error}
    <div class="sidebar-error">{error}</div>
  {:else}
    <div class="schema-list">
      {#each schemas as schema}
        <div class="schema-group">
          <button
            class="schema-toggle"
            class:expanded={expandedSchema === schema}
            onclick={() => toggleSchema(schema)}
          >
            <svg
              class="chevron"
              width="12"
              height="12"
              viewBox="0 0 12 12"
              fill="currentColor"
            >
              <path
                d="M4.5 2L8.5 6L4.5 10"
                stroke="currentColor"
                stroke-width="1.5"
                fill="none"
              />
            </svg>
            {schema}
          </button>

          {#if expandedSchema === schema && tablesBySchema[schema]}
            <div class="table-list">
              {#each tablesBySchema[schema] as table}
                <button
                  class="table-item"
                  class:active={activeTable === `${schema}.${table.name}`}
                  onclick={() => selectTable(schema, table.name)}
                >
                  <svg
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <rect x="3" y="3" width="18" height="18" rx="2" />
                    <line x1="3" y1="9" x2="21" y2="9" />
                    <line x1="9" y1="3" x2="9" y2="21" />
                  </svg>
                  <span class="table-name">{table.name}</span>
                  <span class="table-rows">{table.rows.toLocaleString()}</span>
                </button>
              {:else}
                <div class="sidebar-empty">No tables</div>
              {/each}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</nav>

<style>
  .sidebar {
    width: 260px;
    min-width: 260px;
    height: 100%;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    overflow-y: auto;
  }

  .sidebar-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 16px;
    color: var(--text-primary);
    border-bottom: 1px solid var(--border);
    font-weight: 600;
  }

  .sidebar-title {
    font-size: 14px;
  }

  .sidebar-empty {
    padding: 16px;
    color: var(--text-muted);
    font-size: 13px;
  }

  .sidebar-error {
    padding: 16px;
    color: var(--error);
    font-size: 13px;
  }

  .schema-list {
    padding: 4px 0;
  }

  .schema-group {
    margin-bottom: 2px;
  }

  .schema-toggle {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 16px;
    background: none;
    border: none;
    color: var(--text-secondary);
    font-size: 13px;
    font-weight: 500;
    text-align: left;
  }

  .schema-toggle:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .chevron {
    transition: transform 0.15s ease;
    flex-shrink: 0;
  }

  .schema-toggle.expanded .chevron {
    transform: rotate(90deg);
  }

  .table-list {
    padding: 2px 0;
  }

  .table-item {
    width: calc(100% - 16px);
    margin: 2px 8px;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px 8px 24px;
    background: none;
    border: none;
    border-radius: var(--radius);
    color: var(--text-secondary);
    font-size: 13px;
    text-align: left;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .table-item:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
    transform: translateX(2px);
  }

  .table-item.active {
    background: var(--accent-light);
    color: var(--accent-dim);
    font-weight: 600;
  }

  .table-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
  }

  .table-rows {
    font-size: 11px;
    color: var(--text-muted);
    font-family: var(--font-mono);
  }
</style>
