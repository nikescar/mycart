<script lang="ts">
  let message = '';

  async function enableMaintenanceAndRestart() {
    if (!confirm('This will enable maintenance mode and restart the server. Continue?')) {
      return;
    }

    message = 'Enabling maintenance mode and restarting...';

    try {
      const res = await fetch('/_/api/maintenance/enable-and-restart', {
        method: 'POST'
      });
      const data = await res.json();
      message = data.message || data.error;
    } catch (err) {
      message = 'Server is restarting...';
    }
  }
</script>

<div class="settings-maintenance">
  <h2>Maintenance Mode</h2>

  <p>
    Enable maintenance mode to perform database operations, backups, and system maintenance.
    Only localhost will have access during maintenance mode.
  </p>

  {#if message}
    <div class="message">{message}</div>
  {/if}

  <button on:click={enableMaintenanceAndRestart} class="btn-danger">
    Enable Maintenance Mode & Restart
  </button>

  <div class="info">
    <h3>What happens when you click this button:</h3>
    <ul>
      <li>Server enters maintenance mode</li>
      <li>Only localhost can access the application</li>
      <li>Server restarts gracefully</li>
      <li>Access maintenance panel at: http://localhost:8080/_/maintenance</li>
    </ul>

    <h3>To disable maintenance mode:</h3>
    <p>Run this command via SSH/terminal:</p>
    <code>./mycart maintenance disable</code>
  </div>
</div>

<style>
  .settings-maintenance {
    padding: 2rem;
    max-width: 800px;
  }

  .btn-danger {
    padding: 0.75rem 1.5rem;
    background: #dc3545;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 1rem;
    font-weight: bold;
  }

  .btn-danger:hover {
    background: #c82333;
  }

  .message {
    padding: 1rem;
    margin: 1rem 0;
    background: #f0f0f0;
    border-radius: 4px;
  }

  .info {
    margin-top: 2rem;
    padding: 1rem;
    background: #f8f9fa;
    border-left: 4px solid #007bff;
  }

  .info h3 {
    margin-top: 1rem;
  }

  .info code {
    display: block;
    padding: 0.5rem;
    margin: 0.5rem 0;
    background: #e9ecef;
    border-radius: 4px;
    font-family: monospace;
  }

  ul {
    margin: 0.5rem 0;
  }

  li {
    margin: 0.25rem 0;
  }
</style>
