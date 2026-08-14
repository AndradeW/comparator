(function () {
  "use strict";

  var baseUrl = (typeof API_BASE_URL !== "undefined" && API_BASE_URL) || "http://localhost:8080";

  var els = {
    compareBtn: document.getElementById("compare"),
    verdict: document.getElementById("verdict"),
    ledger: document.getElementById("ledger"),
    results: document.getElementById("results"),
    error: document.getElementById("error"),
    apiBase: document.getElementById("apiBase")
  };

  els.apiBase.textContent = baseUrl;

  var SENTINEL_A_MISSING = "key not found in first JSON";
  var SENTINEL_B_MISSING = "key not found in second JSON";

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
      throw new Error("Falta la URL del Request " + (id === 1 ? "A" : "B"));
    }
    return {
      url: url,
      headers: readJson(document.getElementById("headers" + id).value, "Headers " + (id === 1 ? "A" : "B")),
      params: readJson(document.getElementById("params" + id).value, "Parámetros " + (id === 1 ? "A" : "B"))
    };
  }

  function showError(msg) {
    els.error.textContent = msg;
    els.error.hidden = false;
    els.results.hidden = true;
  }

  function clearFeedback() {
    els.error.hidden = true;
  }

  function fmt(v) {
    if (v === null) {
      return "null";
    }
    if (typeof v === "string") {
      return v;
    }
    if (typeof v === "number" || typeof v === "boolean") {
      return String(v);
    }
    try {
      return JSON.stringify(v);
    } catch (e) {
      return String(v);
    }
  }

  function el(tag, className, text) {
    var node = document.createElement(tag);
    if (className) {
      node.className = className;
    }
    if (text !== undefined && text !== null) {
      node.textContent = text;
    }
    return node;
  }

  function renderVerdict(total) {
    var differ = total > 0;
    var stamp = el("div", "stamp " + (differ ? "stamp-difiere" : "stamp-coinciden"));
    stamp.appendChild(el("span", "stamp-eyebrow", "Veredicto"));
    stamp.appendChild(el("span", "stamp-text", differ ? "Difiere · " + total : "Coinciden"));
    els.verdict.textContent = "";
    els.verdict.appendChild(stamp);
  }

  function renderLedger(data) {
    var statusCodes = data.status_codes || [];
    var headers = data.headers || {};
    var body = data.body_differences || {};

    var headerKeys = Object.keys(headers);
    var bodyKeys = Object.keys(body);
    var statusDiff = statusCodes.length === 2;
    var total = (statusDiff ? 1 : 0) + headerKeys.length + bodyKeys.length;

    renderVerdict(total);

    var root = document.createElement("div");
    els.ledger.textContent = "";
    els.ledger.appendChild(root);

    if (statusCodes.length > 0) {
      var g = group("Status");
      var row = el("div", "row row-status" + (statusDiff ? " row-diff" : " row-equal"));
      row.appendChild(el("span", "cell cell-label", "status"));
      var code = String(statusCodes[0]);
      row.appendChild(el("span", "cell cell-a cell-val", code));
      row.appendChild(el("span", "cell cell-b cell-val", statusDiff ? String(statusCodes[1]) : code));
      g.appendChild(row);
      root.appendChild(g);
    }

    if (headerKeys.length > 0) {
      var gHeaders = group("Headers");
      headerKeys.forEach(function (key) {
        var vals = headers[key];
        gHeaders.appendChild(diffRow(key, vals[0], vals[1]));
      });
      root.appendChild(gHeaders);
    }

    if (bodyKeys.length > 0) {
      var gBody = group("Cuerpo");
      bodyKeys.forEach(function (key) {
        if (key === "error") {
          gBody.appendChild(rawBodyRow(body[key]));
        } else {
          gBody.appendChild(bodyDiffRow(key, body[key]));
        }
      });
      root.appendChild(gBody);
    }

    if (total === 0) {
      root.appendChild(el("p", "ledger-note", "No hay diferencias: status, headers y cuerpo coinciden."));
    }
  }

  function group(label) {
    return el("h3", "group-label", label);
  }

  function diffRow(label, a, b) {
    var row = el("div", "row row-diff");
    row.appendChild(el("span", "cell cell-label", label));
    row.appendChild(el("span", "cell cell-a cell-val", fmt(a)));
    row.appendChild(el("span", "cell cell-b cell-val", fmt(b)));
    return row;
  }

  function bodyDiffRow(key, vals) {
    if (vals.length === 3 && vals[0] === "different lengths") {
      var note = el("div", "row");
      note.appendChild(el("span", "cell cell-note", key + " · longitudes distintas: A=" + vals[1] + ", B=" + vals[2]));
      return note;
    }

    var a = vals[0];
    var b = vals[1];
    var aMissing = a === SENTINEL_A_MISSING;
    var bMissing = b === SENTINEL_B_MISSING;

    var row = el("div", "row row-diff");
    row.appendChild(el("span", "cell cell-label", key));
    row.appendChild(el("span", "cell cell-a cell-val" + (aMissing ? " cell-missing" : ""), aMissing ? "(no existe en A)" : fmt(a)));
    row.appendChild(el("span", "cell cell-b cell-val" + (bMissing ? " cell-missing" : ""), bMissing ? "(no existe en B)" : fmt(b)));
    return row;
  }

  function rawBodyRow(vals) {
    var block = el("div", "raw-block");
    block.appendChild(el("pre", "raw-cell", fmt(vals[0])));
    block.appendChild(el("pre", "raw-cell", fmt(vals[1])));
    return block;
  }

  els.compareBtn.addEventListener("click", async function () {
    clearFeedback();
    var request1, request2;
    try {
      request1 = buildRequest(1);
      request2 = buildRequest(2);
    } catch (e) {
      showError(e.message);
      return;
    }

    els.compareBtn.disabled = true;
    els.compareBtn.textContent = "Comparando…";
    try {
      var resp = await fetch(baseUrl + "/compare", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ request1: request1, request2: request2 })
      });

      if (!resp.ok) {
        var errText = await resp.text();
        throw new Error("Error " + resp.status + " · " + (errText || "no se pudo comparar").trim());
      }

      var data = await resp.json();
      renderLedger(data);
      els.results.hidden = false;
    } catch (e) {
      showError("No se pudo comparar: " + e.message);
    } finally {
      els.compareBtn.disabled = false;
      els.compareBtn.textContent = "Comparar";
    }
  });
})();
