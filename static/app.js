/* Minimal “HTMX-style” interactions + SSE live refresh (offline-friendly) */

(function () {
  const HX_REQ_HEADER = { "HX-Request": "true" };

  function qs(sel, root = document) { return root.querySelector(sel); }
  function qsa(sel, root = document) { return Array.from(root.querySelectorAll(sel)); }

  function debounce(fn, ms) {
    let t;
    return (...args) => {
      clearTimeout(t);
      t = setTimeout(() => fn(...args), ms);
    };
  }

  function hxFetch(url, opts = {}) {
    return fetch(url, {
      credentials: "same-origin",
      headers: Object.assign({}, HX_REQ_HEADER, opts.headers || {}),
      method: opts.method || "GET",
      body: opts.body,
    });
  }

  function applySwap(target, html, swap) {
    if (!target) return;
    swap = swap || "innerHTML";
    if (swap === "outerHTML") {
      target.outerHTML = html;
    } else {
      target.innerHTML = html;
    }
  }

  function parseTrigger(str) {
    // supports: "keyup delay:300ms" or "change" or "click"
    if (!str) return { event: "click", delay: 0 };
    const parts = str.split(/\s+/);
    let event = parts[0];
    let delay = 0;
    for (const p of parts) {
      const m = p.match(/^delay:(\d+)ms$/);
      if (m) delay = parseInt(m[1], 10) || 0;
    }
    return { event, delay };
  }

  function buildUrlWithInputs(baseUrl, el) {
    // If the element is an input/select with a name, build query from key filter inputs (simple).
    // For our filters, we read known ids if present.
    const u = new URL(baseUrl, window.location.origin);
    const map = [
      ["alc", "#f_alc"],
      ["tag", "#f_tag"],
      ["include", "#f_include"],
      ["exclude", "#f_exclude"],
      ["q", "#prodSearch"],
    ];
    for (const [key, id] of map) {
      const node = qs(id);
      if (node && node.value != null && String(node.value).trim() !== "") {
        u.searchParams.set(key, node.value);
      } else {
        u.searchParams.delete(key);
      }
    }
    return u.pathname + u.search;
  }

  function wireHX() {
    // data-hx-get / data-hx-post (to avoid requiring external htmx.js)
    qsa("[data-hx-get],[data-hx-post]").forEach((el) => {
      const getUrl = el.getAttribute("data-hx-get");
      const postUrl = el.getAttribute("data-hx-post");
      const targetSel = el.getAttribute("data-hx-target");
      const swap = el.getAttribute("data-hx-swap") || "innerHTML";
      const trig = parseTrigger(el.getAttribute("data-hx-trigger") || "");

      const handler = async (evt) => {
        if (evt) evt.preventDefault();

        const url = buildUrlWithInputs(getUrl || postUrl, el);
        const target = targetSel ? qs(targetSel) : el;

        try {
          let resp;
          if (getUrl) {
            resp = await hxFetch(url, { method: "GET" });
          } else {
            // post: if element is inside form, serialize that form; else empty post
            const form = el.tagName === "FORM" ? el : el.closest("form");
            const body = form ? new FormData(form) : null;
            resp = await hxFetch(postUrl, { method: "POST", body });
          }

          const redir = resp.headers.get("HX-Redirect");
          if (redir) { window.location.assign(redir); return; }

          const html = await resp.text();
          applySwap(target, html, swap);
        } catch (e) {
          console.warn("hx error", e);
        }
      };

      let attach = handler;
      if (trig.delay > 0) attach = debounce(handler, trig.delay);

      // choose event
      el.addEventListener(trig.event, attach);
    });
  }

  /* ---------------- SSE ---------------- */

  function beep() {
    try {
      const ctx = new (window.AudioContext || window.webkitAudioContext)();
      const o = ctx.createOscillator();
      const g = ctx.createGain();
      o.type = "sine";
      o.frequency.value = 880;
      o.connect(g);
      g.connect(ctx.destination);
      g.gain.setValueAtTime(0.0001, ctx.currentTime);
      g.gain.exponentialRampToValueAtTime(0.2, ctx.currentTime + 0.01);
      g.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 0.12);
      o.start();
      o.stop(ctx.currentTime + 0.13);
    } catch {}
  }

  async function refreshPartial(kind) {
    const path = document.body.getAttribute("data-path") || window.location.pathname;

    try {
      if (kind === "inventory") {
        // update user cocktail grid if present
        const grid = qs("#cocktailGrid");
        if (grid) {
          const url = buildUrlWithInputs("/partials/user/cocktails", grid);
          const resp = await hxFetch(url);
          const html = await resp.text();
          grid.innerHTML = html;
        }
        // bartender products
        const pt = qs("#productsTable");
        if (pt && path.startsWith("/bartender/products")) {
          const url = buildUrlWithInputs("/partials/bartender/products", pt);
          const resp = await hxFetch(url);
          const html = await resp.text();
          pt.innerHTML = html;
        }
        // bartender cocktails list
        const ct = qs("#cocktailsTable");
        if (ct && path.startsWith("/bartender/cocktails")) {
          const resp = await hxFetch("/partials/bartender/cocktails");
          const html = await resp.text();
          ct.innerHTML = html;
        }
        return;
      }

      if (kind === "orders") {
        const ordersList = qs("#ordersList");
        if (!ordersList) return;

        if (path === "/orders") {
          const resp = await hxFetch("/partials/user/orders");
          ordersList.innerHTML = await resp.text();
          return;
        }
        if (path.startsWith("/bartender")) {
          const resp = await hxFetch("/partials/bartender/orders");
          ordersList.innerHTML = await resp.text();
          return;
        }
      }
    } catch (e) {
      console.warn("refreshPartial error", e);
    }
  }

  function wireSSE() {
    // Only if logged in (data-user exists)
    const userId = document.body.getAttribute("data-user");
    if (!userId) return;

    const role = document.body.getAttribute("data-role") || "";
    const es = new EventSource("/sse");

    es.addEventListener("order:created", (e) => {
      if (role === "BARTENDER" || role === "ADMIN") {
        beep();
        refreshPartial("orders");
      }
    });

    es.addEventListener("order:updated", () => {
      refreshPartial("orders");
    });

    es.addEventListener("inventory:updated", () => {
      refreshPartial("inventory");
    });

    es.addEventListener("hello", () => {});
    es.onerror = () => {
      // browser will reconnect automatically
    };
  }

  document.addEventListener("DOMContentLoaded", () => {
    wireHX();
    wireSSE();

    // dashboard convenience auto refresh placeholder
    const auto = qs("[data-auto-refresh='orders']");
    const path = document.body.getAttribute("data-path") || "";
    if (auto && path === "/bartender") {
      // load initial queue partial into the dashboard
      const host = qs("#ordersList");
      if (host) {
        refreshPartial("orders");
      }
    }
  });
})();
