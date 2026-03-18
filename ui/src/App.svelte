<script lang="ts">
  import SchemaNav from "./lib/SchemaNav.svelte";
  import TableView from "./lib/TableView.svelte";
  import QueryEditor from "./lib/QueryEditor.svelte";
  import { fetchStatus } from "./lib/api";

  let activeView: "table" | "query" = $state("table");
  let selectedSchema: string | null = $state(null);
  let selectedTable: string | null = $state(null);
  let readOnly = $state(true);

  function handleSelectTable(schema: string, table: string) {
    selectedSchema = schema;
    selectedTable = table;
    activeView = "table";
  }

  fetchStatus()
    .then((s) => {
      readOnly = s.readOnly;
    })
    .catch(() => {});
</script>

<div class="app">
  <SchemaNav onselect={handleSelectTable} />

  <main class="main">
    <div class="tab-bar">
      <button
        class="tab"
        class:active={activeView === "table"}
        onclick={() => (activeView = "table")}
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <rect x="3" y="3" width="18" height="18" rx="2" />
          <line x1="3" y1="9" x2="21" y2="9" />
          <line x1="9" y1="3" x2="9" y2="21" />
        </svg>
        Browse
      </button>
      <button
        class="tab"
        class:active={activeView === "query"}
        onclick={() => (activeView = "query")}
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <polyline points="16,18 22,12 16,6" />
          <polyline points="8,6 2,12 8,18" />
        </svg>
        Query
      </button>

      <div class="tab-spacer"></div>

      <div class="status-badge" class:read-only={readOnly}>
        {readOnly ? "Read-only" : "Read-write"}
      </div>
    </div>

    <div class="content">
      {#if activeView === "query"}
        <QueryEditor />
      {:else if selectedSchema && selectedTable}
        <TableView schema={selectedSchema} table={selectedTable} />
      {:else}
        <div class="empty-state">
          <svg
            width="48"
            height="48"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1"
            opacity="0.3"
          >
            <ellipse cx="12" cy="5" rx="9" ry="3" />
            <path d="M3 5V19A9 3 0 0 0 21 19V5" />
            <path d="M3 12A9 3 0 0 0 21 12" />
          </svg>
          <p>Select a table from the sidebar to browse its data</p>
        </div>
      {/if}
    </div>
  </main>
</div>

<style>
  .app {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
  }

  .main {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
  }

  .tab-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 0 24px;
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
    height: 52px;
  }

  .tab {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 16px;
    height: 100%;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-secondary);
    font-size: 14px;
    font-weight: 500;
    transition: color 0.15s;
  }

  .tab:hover {
    color: var(--text-primary);
  }

  .tab.active {
    color: var(--accent);
    border-bottom-color: var(--accent);
  }

  .tab-spacer {
    flex: 1;
  }

  .status-badge {
    font-size: 12px;
    font-weight: 500;
    padding: 4px 10px;
    border-radius: var(--radius);
    font-family: var(--font-sans);
    border: 1px solid var(--border);
  }

  .status-badge.read-only {
    background: var(--bg-secondary);
    color: var(--warning);
    border-color: rgba(245, 158, 11, 0.4);
  }

  .status-badge:not(.read-only) {
    background: var(--accent);
    color: white;
    border-color: var(--accent-dim);
  }

  .content {
    flex: 1;
    overflow: hidden;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 16px;
    color: var(--text-muted);
    font-size: 14px;
  }
</style>
