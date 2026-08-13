(function () {
  "use strict";

  var baseUrl = (typeof API_BASE_URL !== "undefined" && API_BASE_URL) || "http://localhost:8080";
  var compareBtn = document.getElementById("compare");
  var output = document.getElementById("output");
  var results = document.getElementById("results");
  var errorEl = document.getElementById("error");
  var apiBaseEl = document.getElementById("apiBase");

  apiBaseEl.textContent = baseUrl;

  function readJson(value, name) {
    if (!value || !value.trim()) {
      return {};
    }
    try {
      var parsed = JSON.parse(value);
      if (typeof parsed !== "object" || Array.isArray(parsed)) {
        throw new Error("debe ser un objeto");
      }
      return parsed;
    } catch (e) {
      throw new Error(name + " no es un JSON de objeto válido: " + e.message);
    }
  }

  function buildRequest(id) {
    var url = document.getElementById("url" + id).value.trim();
    if (!url) {
      throw new Error("Falta la URL del Request " + id);
    }
    return {
      url: url,
      headers: readJson(document.getElementById("headers" + id).value, "Headers " + id),
      params: readJson(document.getElementById("params" + id).value, "Params " + id)
    };
  }

  function showError(msg) {
    errorEl.textContent = msg;
    errorEl.classList.remove("hidden");
    results.classList.add("hidden");
  }

  function hideError() {
    errorEl.classList.add("hidden");
  }

  compareBtn.addEventListener("click", async function () {
    hideError();
    var request1, request2;
    try {
      request1 = buildRequest(1);
      request2 = buildRequest(2);
    } catch (e) {
      showError(e.message);
      return;
    }

    compareBtn.disabled = true;
    compareBtn.textContent = "Comparando...";
    try {
      var resp = await fetch(baseUrl + "/compare", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ request1: request1, request2: request2 })
      });

      if (!resp.ok) {
        var errText = await resp.text();
        throw new Error("Error " + resp.status + ": " + errText);
      }

      var data = await resp.json();
      output.textContent = JSON.stringify(data, null, 2);
      results.classList.remove("hidden");
    } catch (e) {
      showError("No se pudo comparar: " + e.message);
    } finally {
      compareBtn.disabled = false;
      compareBtn.textContent = "Comparar";
    }
  });
})();
