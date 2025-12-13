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
          container.innerHTML = '<p class="text-muted">No API keys yet. Generate one above to get started.</p>';
          return;
        }

        const html = keys
          .map((key) => {
            const createdAt = new Date(key.created_at).toLocaleString();
            const lastUsed = key.last_used_at ? new Date(key.last_used_at).toLocaleString() : "Never";
            const expires = key.expires_at ? new Date(key.expires_at).toLocaleString() : "Never";
            const status = key.is_active
              ? '<span class="badge bg-success">Active</span>'
              : '<span class="badge bg-danger">Revoked</span>';

            return `
                        <div class="card mb-2">
                            <div class="card-body">
                                <div class="d-flex justify-content-between align-items-start">
                                    <div>
                                        <h6 class="card-title">${key.name || "Unnamed Key"}</h6>
                                        <p class="card-text text-muted mb-1">
                                            <small>Prefix: <code>${key.key_prefix}...</code></small><br>
                                            <small>Created: ${createdAt}</small><br>
                                            <small>Last Used: ${lastUsed}</small><br>
                                            <small>Expires: ${expires}</small>
                                        </p>
                                        ${status}
                                    </div>
                                    ${
                                      key.is_active
                                        ? `<button class="btn btn-sm btn-danger" onclick="revokeKey('${key.id}')">Revoke</button>`
                                        : ""
                                    }
                                </div>
                            </div>
                        </div>
                    `;
          })
          .join("");

        container.innerHTML = html;
      })
      .catch((err) => {
        console.error("Error loading keys:", err);
        container.innerHTML = '<p class="text-danger">Failed to load API keys.</p>';
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
                    <div class="alert alert-success">
                        <h6>API Key Created!</h6>
                        <p class="mb-2">Please save this key securely. It will only be shown once:</p>
                        <div class="input-group">
                            <input type="text" class="form-control" value="${data.key}" id="new-key-value" readonly>
                            <button class="btn btn-outline-secondary" type="button" onclick="copyKey(event)">Copy</button>
                        </div>
                        <small class="text-muted">${data.warning}</small>
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
        resultDiv.innerHTML = `
                    <div class="alert alert-danger">Failed to create API key: ${err.message}</div>
                `;
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
