/* Small UI behaviors for theme, HTMX-style requests, ingredient editing, and SSE refreshes. */

(function () {
  const THEME_KEY = "hb-theme";
  const COCKTAIL_VIEW_KEY = "hb-cocktail-view-v2";
  const HX_REQ_HEADER = { "HX-Request": "true" };
  const systemThemeQuery = window.matchMedia ? window.matchMedia("(prefers-color-scheme: dark)") : null;

  function qs(sel, root = document) {
    return root.querySelector(sel);
  }

  function qsa(sel, root = document) {
    return Array.from(root.querySelectorAll(sel));
  }

  function collect(sel, root = document) {
    const nodes = [];
    if (root.matches && root.matches(sel)) {
      nodes.push(root);
    }
    return nodes.concat(qsa(sel, root));
  }

  function debounce(fn, ms) {
    let t;
    return (...args) => {
      clearTimeout(t);
      t = setTimeout(() => fn(...args), ms);
    };
  }

  function getStoredTheme() {
    try {
      const value = localStorage.getItem(THEME_KEY);
      return value === "light" || value === "dark" ? value : null;
    } catch (err) {
      return null;
    }
  }

  function getSystemTheme() {
    return systemThemeQuery && systemThemeQuery.matches ? "dark" : "light";
  }

  function getActiveTheme() {
    return document.documentElement.getAttribute("data-theme") || getStoredTheme() || getSystemTheme();
  }

  function updateThemeToggle(theme) {
    const button = qs("[data-theme-toggle]");
    const label = qs("[data-theme-toggle-label]");
    if (!button || !label) {
      return;
    }

    const nextTheme = theme === "dark" ? "light" : "dark";
    label.textContent = nextTheme === "dark" ? "Dark mode" : "Light mode";
    button.setAttribute("data-theme", theme);
    button.setAttribute("aria-label", "Switch to " + nextTheme + " theme");
    button.setAttribute("aria-pressed", String(theme === "dark"));
  }

  function applyTheme(theme) {
    document.documentElement.setAttribute("data-theme", theme);
    updateThemeToggle(theme);
  }

  function setTheme(theme) {
    try {
      localStorage.setItem(THEME_KEY, theme);
    } catch (err) {}
    applyTheme(theme);
  }

  function wireThemeToggle() {
    const button = qs("[data-theme-toggle]");
    if (!button || button.dataset.bound === "1") {
      updateThemeToggle(getActiveTheme());
      return;
    }

    button.dataset.bound = "1";
    updateThemeToggle(getActiveTheme());
    button.addEventListener("click", () => {
      const nextTheme = getActiveTheme() === "dark" ? "light" : "dark";
      setTheme(nextTheme);
    });
  }

  function normalizeCocktailView(view) {
    return view === "list" ? "list" : "cards";
  }

  function getStoredCocktailView() {
    try {
      const value = localStorage.getItem(COCKTAIL_VIEW_KEY);
      return value ? normalizeCocktailView(value) : "list";
    } catch (err) {
      return "list";
    }
  }

  function applyCocktailView(view) {
    const mode = normalizeCocktailView(view);

    qsa("[data-cocktail-view-target]").forEach((node) => {
      node.setAttribute("data-cocktail-view", mode);
    });

    qsa("[data-cocktail-view-set]").forEach((button) => {
      const targetMode = normalizeCocktailView(button.getAttribute("data-cocktail-view-set"));
      const isActive = targetMode === mode;
      button.setAttribute("aria-pressed", String(isActive));
      button.classList.toggle("btn--primary", isActive);
      button.classList.toggle("btn--ghost", !isActive);
    });
  }

  function setCocktailView(view) {
    const mode = normalizeCocktailView(view);
    try {
      localStorage.setItem(COCKTAIL_VIEW_KEY, mode);
    } catch (err) {}
    applyCocktailView(mode);
  }

  function wireCocktailViewToggle(root = document) {
    collect("[data-cocktail-view-set]", root).forEach((button) => {
      if (button.dataset.bound === "1") {
        return;
      }
      button.dataset.bound = "1";

      button.addEventListener("click", () => {
        setCocktailView(button.getAttribute("data-cocktail-view-set"));
      });
    });

    if (qsa("[data-cocktail-view-target]").length > 0) {
      applyCocktailView(getStoredCocktailView());
    }
  }

  function hxFetch(url, opts = {}) {
    const headers = Object.assign({}, HX_REQ_HEADER, opts.headers || {});
    if (opts.body instanceof URLSearchParams && !headers["Content-Type"] && !headers["content-type"]) {
      headers["Content-Type"] = "application/x-www-form-urlencoded;charset=UTF-8";
    }

    return fetch(url, {
      credentials: "same-origin",
      headers,
      method: opts.method || "GET",
      body: opts.body,
    });
  }

  function applySwap(target, html, swap) {
    if (!target) {
      return;
    }

    const mode = swap || "innerHTML";
    if (mode === "outerHTML") {
      const wrapper = document.createElement("div");
      wrapper.innerHTML = html.trim();
      const next = wrapper.firstElementChild;
      if (!next) {
        return;
      }
      target.replaceWith(next);
      initUI(next);
      return;
    }

    target.innerHTML = html;
    initUI(target);
  }

  function parseTrigger(str) {
    if (!str) {
      return { event: "click", delay: 0 };
    }

    const parts = str.split(/\s+/);
    let event = parts[0];
    let delay = 0;
    for (const part of parts) {
      const match = part.match(/^delay:(\d+)ms$/);
      if (match) {
        delay = parseInt(match[1], 10) || 0;
      }
    }
    return { event, delay };
  }

  function defaultHXTrigger(el, getUrl, postUrl) {
    if (postUrl && el.tagName === "FORM") {
      return { event: "submit", delay: 0 };
    }

    return { event: "click", delay: 0 };
  }

  function buildPostBody(form) {
    if (!form) {
      return null;
    }

    const enctype = (form.getAttribute("enctype") || "").toLowerCase();
    if (enctype === "multipart/form-data") {
      return new FormData(form);
    }

    const params = new URLSearchParams();
    new FormData(form).forEach((value, key) => {
      if (typeof value === "string") {
        params.append(key, value);
      }
    });
    return params;
  }

  function buildUrlWithInputs(baseUrl) {
    const url = new URL(baseUrl, window.location.origin);
    const map = [
      ["alc", "#f_alc"],
      ["tag", "#f_tag"],
      ["include", "#f_include"],
      ["exclude", "#f_exclude"],
      ["q", "#prodSearch"],
    ];

    for (const [key, selector] of map) {
      const node = qs(selector);
      if (node && node.value != null && String(node.value).trim() !== "") {
        url.searchParams.set(key, node.value);
      } else {
        url.searchParams.delete(key);
      }
    }

    return url.pathname + url.search;
  }

  function wireHX(root = document) {
    collect("[data-hx-get],[data-hx-post]", root).forEach((el) => {
      if (el.dataset.hxBound === "1") {
        return;
      }
      el.dataset.hxBound = "1";

      const getUrl = el.getAttribute("data-hx-get");
      const postUrl = el.getAttribute("data-hx-post");
      const targetSel = el.getAttribute("data-hx-target");
      const swap = el.getAttribute("data-hx-swap") || "innerHTML";
      const triggerAttr = el.getAttribute("data-hx-trigger");
      const trig = triggerAttr ? parseTrigger(triggerAttr) : defaultHXTrigger(el, getUrl, postUrl);

      const handler = async (evt) => {
        if (evt) {
          evt.preventDefault();
        }

        const url = buildUrlWithInputs(getUrl || postUrl);
        const target = targetSel ? qs(targetSel) : el;

        try {
          let resp;
          if (getUrl) {
            resp = await hxFetch(url, { method: "GET" });
          } else {
            const form = el.tagName === "FORM" ? el : el.closest("form");
            const body = buildPostBody(form);
            resp = await hxFetch(postUrl, { method: "POST", body });
          }

          const redirect = resp.headers.get("HX-Redirect");
          if (redirect) {
            window.location.assign(redirect);
            return;
          }

          const html = await resp.text();
          applySwap(target, html, swap);
        } catch (err) {
          console.warn("hx error", err);
        }
      };

      const attach = trig.delay > 0 ? debounce(handler, trig.delay) : handler;
      el.addEventListener(trig.event, attach);
    });
  }

  function clearIngredientRow(row) {
    qsa("input, select, textarea", row).forEach((field) => {
      if (field.tagName === "SELECT") {
        field.selectedIndex = 0;
      } else if (field.type === "checkbox" || field.type === "radio") {
        field.checked = false;
      } else {
        field.value = "";
      }
    });

    const required = qs("select[name='ingredient_required']", row);
    if (required) {
      required.value = "1";
    }
  }

  function syncIngredientRow(row, index) {
    const required = qs("select[name='ingredient_required']", row);
    const isRequired = !required || required.value !== "0";
    const title = qs("[data-ingredient-title]", row);
    const badge = qs("[data-ingredient-rule-badge]", row);
    const hint = qs("[data-ingredient-rule-help]", row);

    if (title) {
      title.textContent = "Ingredient " + index;
    }

    row.classList.toggle("ingredient-row--required", isRequired);
    row.classList.toggle("ingredient-row--optional", !isRequired);

    if (badge) {
      badge.textContent = isRequired ? "Required" : "Optional";
      badge.classList.toggle("ingredient-row__rule--required", isRequired);
      badge.classList.toggle("ingredient-row__rule--optional", !isRequired);
    }

    if (hint) {
      hint.textContent = isRequired
        ? "If this ingredient is unavailable, the cocktail is hidden from ordering."
        : "This ingredient stays on the recipe, but it does not block ordering.";
    }
  }

  function wireIngredientEditors(root = document) {
    collect("[data-ingredient-editor]", root).forEach((editor) => {
      if (editor.dataset.editorBound === "1") {
        return;
      }
      editor.dataset.editorBound = "1";

      const list = qs("[data-ingredient-list]", editor);
      const template = qs("template[data-ingredient-template]", editor);
      const addButton = qs("[data-ingredient-add]", editor);
      if (!list || !template) {
        return;
      }

      function syncRows() {
        qsa("[data-ingredient-row]", list).forEach((row, index) => {
          syncIngredientRow(row, index + 1);
        });
      }

      function wireRuleFields(scope) {
        collect("select[name='ingredient_required']", scope).forEach((field) => {
          if (field.dataset.bound === "1") {
            return;
          }
          field.dataset.bound = "1";
          field.addEventListener("change", syncRows);
        });
      }

      function wireRemoveButtons(scope) {
        collect("[data-ingredient-remove]", scope).forEach((button) => {
          if (button.dataset.bound === "1") {
            return;
          }
          button.dataset.bound = "1";

          button.addEventListener("click", () => {
            const row = button.closest("[data-ingredient-row]");
            if (!row) {
              return;
            }

            const rows = qsa("[data-ingredient-row]", list);
            if (rows.length === 1) {
              clearIngredientRow(row);
              syncRows();
              const productField = qs("select[name='ingredient_product_id']", row);
              if (productField) {
                productField.focus();
              }
              return;
            }

            row.remove();
            syncRows();
          });
        });
      }

      wireRemoveButtons(editor);
      wireRuleFields(editor);
      syncRows();

      if (addButton) {
        addButton.addEventListener("click", () => {
          const fragment = template.content.cloneNode(true);
          list.appendChild(fragment);
          const row = list.lastElementChild;
          if (row) {
            wireRemoveButtons(row);
            wireRuleFields(row);
            syncRows();
            const productField = qs("select[name='ingredient_product_id']", row);
            if (productField) {
              productField.focus();
            }
          }
        });
      }
    });
  }

  async function refreshPartial(kind) {
    const path = document.body.getAttribute("data-path") || window.location.pathname;

    try {
      if (kind === "inventory") {
        const grid = qs("#cocktailGrid");
        if (grid) {
          const url = buildUrlWithInputs("/partials/user/cocktails");
          const resp = await hxFetch(url);
          grid.innerHTML = await resp.text();
          initUI(grid);
        }

        const productsTable = qs("#productsTable");
        if (productsTable && path.startsWith("/bartender/products")) {
          const url = buildUrlWithInputs("/partials/bartender/products");
          const resp = await hxFetch(url);
          productsTable.innerHTML = await resp.text();
          initUI(productsTable);
        }

        const cocktailsTable = qs("#cocktailsTable");
        if (cocktailsTable && path.startsWith("/bartender/cocktails")) {
          const resp = await hxFetch("/partials/bartender/cocktails");
          cocktailsTable.innerHTML = await resp.text();
          initUI(cocktailsTable);
        }
        return;
      }

      if (kind === "orders") {
        const ordersList = qs("#ordersList");
        if (!ordersList) {
          return;
        }

        if (path === "/orders") {
          const resp = await hxFetch("/partials/user/orders");
          ordersList.innerHTML = await resp.text();
          initUI(ordersList);
          return;
        }

        if (path.startsWith("/bartender")) {
          const resp = await hxFetch("/partials/bartender/orders");
          ordersList.innerHTML = await resp.text();
          initUI(ordersList);
        }
      }
    } catch (err) {
      console.warn("refreshPartial error", err);
    }
  }

  function beep() {
    try {
      const ctx = new (window.AudioContext || window.webkitAudioContext)();
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = "sine";
      osc.frequency.value = 880;
      osc.connect(gain);
      gain.connect(ctx.destination);
      gain.gain.setValueAtTime(0.0001, ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.2, ctx.currentTime + 0.01);
      gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 0.12);
      osc.start();
      osc.stop(ctx.currentTime + 0.13);
    } catch (err) {}
  }

  function wireSSE() {
    const userId = document.body.getAttribute("data-user");
    if (!userId) {
      return;
    }

    const role = document.body.getAttribute("data-role") || "";
    const es = new EventSource("/sse");

    es.addEventListener("order:created", () => {
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
    es.onerror = () => {};
  }

  function initUI(root = document) {
    wireHX(root);
    wireCocktailViewToggle(root);
    wireIngredientEditors(root);
  }

  document.addEventListener("DOMContentLoaded", () => {
    applyTheme(getActiveTheme());
    wireThemeToggle();
    initUI(document);
    wireSSE();

    if (systemThemeQuery) {
      const syncTheme = () => {
        if (!getStoredTheme()) {
          applyTheme(getSystemTheme());
        }
      };

      if (systemThemeQuery.addEventListener) {
        systemThemeQuery.addEventListener("change", syncTheme);
      } else if (systemThemeQuery.addListener) {
        systemThemeQuery.addListener(syncTheme);
      }
    }

    const auto = qs("[data-auto-refresh='orders']");
    const path = document.body.getAttribute("data-path") || "";
    if (auto && path === "/bartender") {
      const host = qs("#ordersList");
      if (host) {
        refreshPartial("orders");
      }
    }
  });
})();
