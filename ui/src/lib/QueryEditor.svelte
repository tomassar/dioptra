<script lang="ts">
  import { runQuery, type QueryResult } from './api'

  let sql = $state('')
  let result: QueryResult | null = $state(null)
  let error: string | null = $state(null)
  let loading = $state(false)
  let elapsed = $state(0)

  async function execute() {
    if (!sql.trim()) return
    loading = true
    error = null
    result = null
    const start = performance.now()

    try {
      result = await runQuery(sql)
    } catch (e: any) {
      error = e.message
    } finally {
      elapsed = Math.round(performance.now() - start)
      loading = false
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault()
      execute()
    }
  }

  function formatCell(val: any): string {
    if (val === null || val === undefined) return 'NULL'
    if (typeof val === 'object') return JSON.stringify(val)
    return String(val)
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
      <button class="run-btn" onclick={execute} disabled={loading || !sql.trim()}>
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
      <span>{result.rows.length} row{result.rows.length !== 1 ? 's' : ''}</span>
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
    padding: 8px 20px;
    background: var(--bg-secondary);
  }

  .run-btn {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 16px;
    background: var(--accent-dim);
    color: white;
    border: none;
    border-radius: var(--radius);
    font-size: 13px;
    font-weight: 500;
  }

  .run-btn:hover:not(:disabled) {
    background: var(--accent);
  }

  .run-btn:disabled {
    opacity: 0.5;
    cursor: default;
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
    padding: 8px 20px;
    font-size: 12px;
    color: var(--text-secondary);
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border);
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

  tr:hover td {
    background: var(--bg-hover);
  }
</style>
