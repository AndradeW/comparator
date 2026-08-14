(function () {
  "use strict";

  var baseUrl = (typeof API_BASE_URL !== "undefined" && API_BASE_URL) || "http://localhost:8080";

  var els = {
    compareBtn: document.getElementById("compare"),
    verdict: document.getElementById("verdict"),
    ledger: document.getElementById("ledger"),
    results: document.getElementById("results"),
    error: document.getElementById("error"),
    volatileToggle: document.getElementById("volatile-headers")
  };

  var SENTINEL_A_MISSING = "key not found in first JSON";
  var SENTINEL_B_MISSING = "key not found in second JSON";

  var VOLATILE_HEADERS = ["date", "age", "cf-ray", "set-cookie", "via", "expires", "x-cache", "x-cache-hits", "x-served-by", "x-request-id"];

  function isVolatileHeader(key) {
    return VOLATILE_HEADERS.indexOf(key.toLowerCase()) !== -1;
  }

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
    var method = document.getElementById("method" + id).value;
    var body = document.getElementById("body" + id).value;
    return {
      url: url,
      method: method,
      headers: readJson(document.getElementById("headers" + id).value, "Headers " + (id === 1 ? "A" : "B")),
      params: readJson(document.getElementById("params" + id).value, "Parámetros " + (id === 1 ? "A" : "B")),
      body: body
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

  // ---------- Render de valores como JSON con sintaxis ----------

  function appendJson(node, v, depth) {
    if (v === null) {
      node.appendChild(el("span", "tok tok-bool", "null"));
      return;
    }
    if (typeof v === "string") {
      node.appendChild(el("span", "tok tok-str", JSON.stringify(v)));
      return;
    }
    if (typeof v === "number") {
      node.appendChild(el("span", "tok tok-num", String(v)));
      return;
    }
    if (typeof v === "boolean") {
      node.appendChild(el("span", "tok tok-bool", String(v)));
      return;
    }
    if (typeof v === "object") {
      appendJsonObject(node, v, depth || 0);
      return;
    }
    node.appendChild(el("span", "tok", String(v)));
  }

  function repeat(str, n) {
    var out = "";
    for (var i = 0; i < n; i++) {
      out += str;
    }
    return out;
  }

  function appendJsonObject(node, v, depth) {
    var isArr = Array.isArray(v);
    var keys = isArr ? v.map(function (_, i) { return String(i); }) : Object.keys(v);

    node.appendChild(el("span", "tok tok-punc", isArr ? "[" : "{"));
    if (keys.length === 0) {
      node.appendChild(el("span", "tok tok-punc", isArr ? "]" : "}"));
      return;
    }

    var innerPad = "\n" + repeat("  ", depth + 1);
    var closePad = "\n" + repeat("  ", depth);
    keys.forEach(function (key, i) {
      node.appendChild(el("span", "tok tok-line", innerPad));
      if (!isArr) {
        node.appendChild(el("span", "tok tok-key", JSON.stringify(key)));
        node.appendChild(el("span", "tok tok-punc", ": "));
      }
      appendJson(node, v[key], depth + 1);
      if (i < keys.length - 1) {
        node.appendChild(el("span", "tok tok-punc", ","));
      }
    });
    node.appendChild(el("span", "tok tok-line", closePad));
    node.appendChild(el("span", "tok tok-punc", isArr ? "]" : "}"));
  }

  function jsonCell(v, side) {
    var cell = el("span", "cell cell-" + (side || "a") + " cell-val");
    appendJson(cell, v);
    return cell;
  }

  // ---------- Paths como sub-nodos ----------

  function pathDepth(key) {
    var depth = 0;
    for (var i = 0; i < key.length; i++) {
      if (key.charAt(i) === ".") {
        depth++;
      }
    }
    return depth;
  }

  function pathLabel(key) {
    var segs = key.split(".");
    var wrap = el("span");
    segs.forEach(function (seg, i) {
      if (i > 0) {
        wrap.appendChild(el("span", "tok tok-punc", "."));
      }
      wrap.appendChild(el("span", "tok " + (i < segs.length - 1 ? "path-anc" : "path-leaf"), seg));
    });
    return wrap;
  }

  function labelCell(key) {
    var cell = el("span", "cell cell-label");
    cell.style.paddingLeft = (1 + pathDepth(key) * 1.1) + "rem";
    cell.appendChild(pathLabel(key));
    return cell;
  }

  function renderVerdict(total) {
    var differ = total > 0;
    var stamp = el("div", "stamp " + (differ ? "stamp-difiere" : "stamp-coinciden"));
    stamp.appendChild(el("span", "stamp-eyebrow", "Veredicto"));
    stamp.appendChild(el("span", "stamp-text", differ ? "Difiere · " + total : "Coinciden"));
    els.verdict.textContent = "";
    els.verdict.appendChild(stamp);
  }

  var lastData = null;

  function renderLedger(data) {
    var statusCodes = data.status_codes || [];
    var headers = data.headers || {};
    var body = data.body_differences || [];

    var showVolatile = !els.volatileToggle.checked;
    var headerKeys = Object.keys(headers).filter(function (key) {
      return showVolatile || !isVolatileHeader(key);
    });
    var hiddenVolatile = headerKeys.length !== Object.keys(headers).length;

    var bodyCount = body.length;
    var statusDiff = statusCodes.length === 2;
    var total = (statusDiff ? 1 : 0) + headerKeys.length + bodyCount;

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
      if (hiddenVolatile) {
        gHeaders.appendChild(el("p", "ledger-note", "Headers volátiles ocultos (Date, Age, Cf-Ray, …)"));
      }
      root.appendChild(gHeaders);
    }

    if (body.length > 0) {
      var gBody = group("Body");
      body.forEach(function (item) {
        if (item.tipo === "error") {
          gBody.appendChild(rawBodyRow(item.values));
        } else {
          gBody.appendChild(bodyDiffRow(item.path, item.values));
        }
      });
      root.appendChild(gBody);
    }

    if (total === 0) {
      root.appendChild(el("p", "ledger-note", "No hay diferencias: status, headers y body coinciden."));
    }
  }

  function group(label) {
    return el("h3", "group-label", label);
  }

  function diffRow(label, a, b) {
    var row = el("div", "row row-diff");
    row.appendChild(el("span", "cell cell-label", label));
    row.appendChild(jsonCell(a, "a"));
    row.appendChild(jsonCell(b, "b"));
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
    row.appendChild(labelCell(key));

    var cellA = el("span", "cell cell-a cell-val" + (aMissing ? " cell-missing" : ""));
    if (aMissing) {
      cellA.textContent = "(no existe en A)";
    } else {
      appendJson(cellA, a);
    }

    var cellB = el("span", "cell cell-b cell-val" + (bMissing ? " cell-missing" : ""));
    if (bMissing) {
      cellB.textContent = "(no existe en B)";
    } else {
      appendJson(cellB, b);
    }

    row.appendChild(cellA);
    row.appendChild(cellB);
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
      lastData = data;
      renderLedger(data);
      els.results.hidden = false;
    } catch (e) {
      showError("No se pudo comparar: " + e.message);
    } finally {
      els.compareBtn.disabled = false;
      els.compareBtn.textContent = "Comparar";
    }
  });

  els.volatileToggle.addEventListener("change", function () {
    if (lastData) {
      renderLedger(lastData);
    }
  });
})();
