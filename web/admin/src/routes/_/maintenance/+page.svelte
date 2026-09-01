<script lang="ts">
  import { onMount } from 'svelte';

  let status = {
    maintenance_mode: false,
    database: { type: '', path: '' },
    storage: { type: '', base_path: '' },
    wrangler: { installed: false }
  };

  let backupPath = '';
  let restorePath = '';
  let switchTarget = '';
  let message = '';

  async function loadStatus() {
    const res = await fetch('/_/api/maintenance/status');
    status = await res.json();
  }

  async function installWrangler() {
    message = 'Installing wrangler...';
    const res = await fetch('/_/api/maintenance/wrangler/install', { method: 'POST' });
    const data = await res.json();
    message = data.message || data.error;
    await loadStatus();
  }

  async function uninstallWrangler() {
    message = 'Uninstalling wrangler...';
    const res = await fetch('/_/api/maintenance/wrangler/uninstall', { method: 'POST' });
    const data = await res.json();
    message = data.message || data.error;
    await loadStatus();
  }

  async function backup() {
    message = 'Creating backup...';
    const res = await fetch('/_/api/maintenance/backup', { method: 'POST' });
    const data = await res.json();
    message = data.message || data.error;
    backupPath = data.backup_path || '';
  }

  async function restore() {
    if (!restorePath) {
      message = 'Please enter backup path';
      return;
    }
    message = 'Restoring...';
    const res = await fetch('/_/api/maintenance/restore', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ backup_path: restorePath })
    });
    const data = await res.json();
    message = data.message || data.error;
  }

  async function switchDatabase() {
    if (!switchTarget) {
      message = 'Please select target database';
      return;
    }
    message = 'Switching database...';
    const res = await fetch('/_/api/maintenance/switch', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ target_type: switchTarget })
    });
    const data = await res.json();
    message = data.message || data.error;
  }

  onMount(loadStatus);
</script>

<div class="maintenance-panel">
  <h1>Maintenance Mode</h1>

  {#if message}
    <div class="message">{message}</div>
  {/if}

  <section>
    <h2>Status</h2>
    <p>Database: {status.database.type} ({status.database.path})</p>
    <p>Storage: {status.storage.type} ({status.storage.base_path})</p>
    <p>Wrangler: {status.wrangler.installed ? 'Installed' : 'Not installed'}</p>
  </section>

  <section>
    <h2>Wrangler Management</h2>
    {#if !status.wrangler.installed}
      <button on:click={installWrangler}>Install Wrangler</button>
    {:else}
      <button on:click={uninstallWrangler}>Uninstall Wrangler</button>
    {/if}
  </section>

  <section>
    <h2>Database Operations</h2>
    <div>
      <button on:click={backup}>Create Backup</button>
      {#if backupPath}
        <p>Backup created: {backupPath}</p>
      {/if}
    </div>

    <div>
      <input type="text" bind:value={restorePath} placeholder="Backup path" />
      <button on:click={restore}>Restore from Backup</button>
    </div>

    <div>
      <select bind:value={switchTarget}>
        <option value="">Select database</option>
        <option value="sqlite">SQLite</option>
        <option value="d1">Cloudflare D1</option>
      </select>
      <button on:click={switchDatabase}>Switch Database</button>
    </div>
  </section>
</div>

<style>
  .maintenance-panel {
    padding: 2rem;
    max-width: 800px;
  }

  section {
    margin: 2rem 0;
    padding: 1rem;
    border: 1px solid #ccc;
    border-radius: 4px;
  }

  button {
    margin: 0.5rem 0.5rem 0.5rem 0;
    padding: 0.5rem 1rem;
    background: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }

  button:hover {
    background: #0056b3;
  }

  input, select {
    margin: 0.5rem 0.5rem 0.5rem 0;
    padding: 0.5rem;
    border: 1px solid #ccc;
    border-radius: 4px;
  }

  .message {
    padding: 1rem;
    margin: 1rem 0;
    background: #f0f0f0;
    border-radius: 4px;
  }
</style>
