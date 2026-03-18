<script lang="ts">
  import { runQuery, type QueryResult } from "./api";

  let sql = $state("");
  let result: QueryResult | null = $state(null);
  let error: string | null = $state(null);
  let loading = $state(false);
  let elapsed = $state(0);

  async function execute() {
    if (!sql.trim()) return;
    loading = true;
    error = null;
    result = null;
    const start = performance.now();

    try {
      result = await runQuery(sql);
    } catch (e: any) {
      error = e.message;
    } finally {
      elapsed = Math.round(performance.now() - start);
      loading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
      e.preventDefault();
      execute();
    }
  }

  function formatCell(val: any): string {
    if (val === null || val === undefined) return "NULL";
    if (typeof val === "object") return JSON.stringify(val);
    return String(val);
  }
</script>

<div class="query-editor">
  <div class="editor-area">
    <textarea
      bind:value={sql}
      onkeydown={handleKeydown}
      placeholder="SELECT * FROM ..."
      spellcheck="false"
    ></textarea>
    <div class="editor-actions">
      <button
        class="run-btn"
        onclick={execute}
        disabled={loading || !sql.trim()}
      >
        {#if loading}
          Running...
        {:else}
          Run
          <kbd>⌘↵</kbd>
        {/if}
      </button>
    </div>
  </div>

  {#if error}
    <div class="query-error">{error}</div>
  {/if}

  {#if result}
    <div class="query-meta">
      <span>{result.rows.length} row{result.rows.length !== 1 ? "s" : ""}</span>
      <span class="dot">·</span>
      <span>{elapsed}ms</span>
    </div>

    <div class="result-scroll">
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
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .query-editor {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }

  .editor-area {
    display: flex;
    flex-direction: column;
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
    overflow: hidden;
    background: var(--bg-primary);
  }

  textarea {
    width: 100%;
    min-height: 120px;
    max-height: 300px;
    resize: vertical;
    padding: 16px 20px;
    background: var(--bg-primary);
    color: var(--text-primary);
    border: none;
    outline: none;
    font-family: var(--font-mono);
    font-size: 13px;
    line-height: 1.6;
  }

  textarea::placeholder {
    color: var(--text-muted);
  }

  .editor-actions {
    display: flex;
    justify-content: flex-end;
    padding: 12px 20px;
    background: var(--bg-secondary);
    border-top: 1px solid var(--border);
  }

  .run-btn {
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

  .run-btn:hover:not(:disabled) {
    background: var(--accent-dim);
    border-color: var(--accent-dim);
  }

  .run-btn:active:not(:disabled) {
    background: var(--text-primary);
  }

  .run-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    background: var(--bg-secondary);
    color: var(--text-muted);
    border-color: var(--border);
    box-shadow: none;
  }

  kbd {
    font-family: var(--font-sans);
    font-size: 11px;
    opacity: 0.7;
  }

  .query-error {
    padding: 12px 20px;
    background: rgba(248, 81, 73, 0.1);
    color: var(--error);
    font-size: 13px;
    font-family: var(--font-mono);
    border-bottom: 1px solid var(--border);
    white-space: pre-wrap;
  }

  .query-meta {
    padding: 12px 24px;
    font-size: 13px;
    color: var(--text-secondary);
    background: transparent;
    font-family: var(--font-mono);
    flex-shrink: 0;
  }

  .dot {
    margin: 0 4px;
    color: var(--text-muted);
  }

  .result-scroll {
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

  tr:hover td {
    background: var(--bg-secondary);
  }
</style>
