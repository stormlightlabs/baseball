(function () {
  "use strict";

  function loadAPIKeys() {
    const container = document.getElementById("api-keys-list");
    if (!container) {
      return;
    }

    fetch("/v1/auth/keys")
      .then((res) => res.json())
      .then((keys) => {
        if (!Array.isArray(keys) || keys.length === 0) {
          container.innerHTML =
            '<p class="muted" style="padding: 1rem">No API keys yet. Generate one above to get started.</p>';
          return;
        }

        const rows = keys
          .map((key) => {
            const createdAt = new Date(key.created_at).toLocaleString();
            const lastUsed = key.last_used_at ? new Date(key.last_used_at).toLocaleString() : "Never";
            const expires = key.expires_at ? new Date(key.expires_at).toLocaleString() : "Never";
            const status = key.is_active
              ? '<span class="status-pill success">Active</span>'
              : '<span class="status-pill muted">Revoked</span>';

            const revokeButton = key.is_active
              ? `<button class="btn btn-danger btn-compact" onclick="revokeKey('${key.id}')">Revoke</button>`
              : "";

            return `
              <tr>
                <td>${key.name || "Unnamed Key"}</td>
                <td><code>${key.key_prefix}...</code></td>
                <td>${createdAt}</td>
                <td>${lastUsed}</td>
                <td>${expires}</td>
                <td>${status}</td>
                <td style="text-align:right">${revokeButton}</td>
              </tr>
            `;
          })
          .join("");

        container.innerHTML = `
          <table class="data-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Prefix</th>
                <th>Created</th>
                <th>Last Used</th>
                <th>Expires</th>
                <th>Status</th>
                <th></th>
              </tr>
            </thead>
            <tbody>${rows}</tbody>
          </table>
        `;
      })
      .catch((err) => {
        console.error("Error loading keys:", err);
        container.innerHTML = '<p class="alert danger" style="margin: 1rem">Failed to load API keys.</p>';
      });
  }

  function revokeKey(id) {
    if (!confirm("Are you sure you want to revoke this API key? This action cannot be undone.")) {
      return;
    }

    fetch(`/v1/auth/keys/${id}`, { method: "DELETE" })
      .then((res) => res.json())
      .then(() => {
        loadAPIKeys();
      })
      .catch((err) => {
        alert("Failed to revoke key: " + err.message);
      });
  }

  function handleCreateKeySubmit(event) {
    event.preventDefault();

    const nameInput = document.getElementById("key-name");
    const name = nameInput ? nameInput.value || null : null;
    const resultDiv = document.getElementById("new-key-result");

    fetch("/v1/auth/keys", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    })
      .then((res) => res.json())
      .then((data) => {
        if (!resultDiv) {
          return;
        }

        resultDiv.innerHTML = `
          <div class="alert success">
            <strong>API Key Created!</strong>
            <p>Please save this key securely. It will only be shown once.</p>
            <div style="display: grid; grid-template-columns: 1fr auto; gap: 0.5rem; margin: 1rem 0;">
              <input type="text" value="${data.key}" id="new-key-value" readonly>
              <button class="btn btn-compact" type="button" onclick="copyKey(event)">Copy</button>
            </div>
            <p class="muted" style="margin: 0">${data.warning}</p>
          </div>
        `;

        if (nameInput) {
          nameInput.value = "";
        }
        loadAPIKeys();

        setTimeout(() => {
          resultDiv.innerHTML = "";
        }, 60000);
      })
      .catch((err) => {
        if (!resultDiv) {
          return;
        }
        resultDiv.innerHTML = `<div class="alert danger">Failed to create API key: ${err.message}</div>`;
      });
  }

  function copyKey(event) {
    const input = document.getElementById("new-key-value");
    if (!input) {
      return;
    }

    input.select();
    document.execCommand("copy");

    if (event && event.target) {
      const btn = event.target;
      const originalText = btn.textContent;
      btn.textContent = "Copied!";
      setTimeout(() => {
        btn.textContent = originalText;
      }, 2000);
    }
  }

  document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("create-key-form");
    if (form) {
      form.addEventListener("submit", handleCreateKeySubmit);
    }

    if (document.body && document.body.dataset.authenticated === "true") {
      loadAPIKeys();
    }
  });

  window.revokeKey = revokeKey;
  window.copyKey = copyKey;
})();
