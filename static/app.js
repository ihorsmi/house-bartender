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
      return value === "light" ? value : null;
    } catch (err) {
      return null;
    }
  }

  function getSystemTheme() {
    return "light";
  }

  function getActiveTheme() {
    return "light";
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
      return value ? normalizeCocktailView(value) : "cards";
    } catch (err) {
      return "cards";
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

  function jsonFetch(url, opts = {}) {
    const headers = Object.assign(
      {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      opts.headers || {}
    );

    return fetch(url, {
      credentials: "same-origin",
      method: opts.method || "GET",
      headers,
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
      ["q", "[data-shell-search-input]"],
    ];

    for (const [key, selector] of map) {
      const node = qs(selector);
      if (!node || node.value == null) {
        continue;
      }
      if (String(node.value).trim() !== "") {
        url.searchParams.set(key, node.value);
      } else {
        url.searchParams.delete(key);
      }
    }

    return url.pathname + url.search;
  }

  function currentPath() {
    return document.body.getAttribute("data-path") || window.location.pathname;
  }

  function isLibrarySearchPath(path) {
    return path === "/" || path.startsWith("/bartender/cocktails");
  }

  function buildLibraryUrl(basePath, overrides = {}) {
    const current = new URL(window.location.href);
    const next = new URL(basePath, window.location.origin);
    current.searchParams.forEach((value, key) => {
      next.searchParams.set(key, value);
    });
    next.searchParams.delete("page_size");

    Object.keys(overrides).forEach((key) => {
      const value = overrides[key];
      if (value == null || String(value).trim() === "") {
        next.searchParams.delete(key);
        return;
      }
      next.searchParams.set(key, String(value));
    });

    return next.pathname + next.search;
  }

  async function refreshLibraryResults(overrides = {}) {
    const path = currentPath();
    const host = qs("#libraryResults");
    if (!host || !isLibrarySearchPath(path)) {
      return;
    }

    const partialPath = path.startsWith("/bartender/cocktails") ? "/partials/bartender/cocktails" : "/partials/user/cocktails";
    const pagePath = path.startsWith("/bartender/cocktails") ? "/bartender/cocktails" : "/";
    const partialUrl = buildLibraryUrl(partialPath, overrides);
    const pageUrl = buildLibraryUrl(pagePath, overrides);

    const resp = await hxFetch(partialUrl);
    applySwap(host, await resp.text(), "innerHTML");
    window.history.replaceState({}, "", pageUrl);
  }

  function getInventoryControls(root = document) {
    return qs("[data-inventory-controls]", root) || qs("[data-inventory-controls]");
  }

  function buildInventoryUrl(basePath, controls = null) {
    const url = new URL(basePath, window.location.origin);
    const scope = controls || getInventoryControls();
    const searchInput = scope ? qs("[data-inventory-search]", scope) : qs("#prodSearch");

    if (searchInput && String(searchInput.value || "").trim() !== "") {
      url.searchParams.set("q", String(searchInput.value).trim());
    }

    return url.pathname + url.search;
  }

  async function refreshInventoryResults() {
    const controls = getInventoryControls();
    const host = qs("#productsTable");
    if (!host || !controls) {
      return;
    }

    const partialPath = controls.getAttribute("data-inventory-partial-url") || "/partials/bartender/products";
    const pagePath = controls.getAttribute("data-inventory-page-url") || controls.getAttribute("action") || currentPath();
    const partialUrl = buildInventoryUrl(partialPath, controls);
    const pageUrl = buildInventoryUrl(pagePath, controls);
    const resp = await hxFetch(partialUrl);
    applySwap(host, await resp.text(), "innerHTML");
    window.history.replaceState({}, "", pageUrl);
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
      const pushUrl = el.getAttribute("data-hx-push-url");
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
          if (getUrl && pushUrl) {
            window.history.replaceState({}, "", buildUrlWithInputs(pushUrl));
          }
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
        const libraryResults = qs("#libraryResults");
        if (libraryResults && isLibrarySearchPath(path)) {
          const partialPath = path.startsWith("/bartender/cocktails") ? "/partials/bartender/cocktails" : "/partials/user/cocktails";
          const resp = await hxFetch(buildLibraryUrl(partialPath));
          applySwap(libraryResults, await resp.text(), "innerHTML");
        }

        const productsTable = qs("#productsTable");
        if (productsTable && getInventoryControls()) {
          await refreshInventoryResults();
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
          applyHighlightedOrder(ordersList, false);
          return;
        }

        if (path.startsWith("/bartender")) {
          const url = path === "/bartender" ? "/partials/bartender/orders?view=dashboard" : "/partials/bartender/orders";
          const resp = await hxFetch(url);
          ordersList.innerHTML = await resp.text();
          initUI(ordersList);
          applyHighlightedOrder(ordersList, false);
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

  let highlightScrollDone = false;

  function getHighlightedOrderId() {
    try {
      const value = new URLSearchParams(window.location.search).get("highlight");
      return value && /^\d+$/.test(value) ? value : "";
    } catch (err) {
      return "";
    }
  }

  function applyHighlightedOrder(root = document, shouldScroll = false) {
    const highlightedOrderId = getHighlightedOrderId();
    if (!highlightedOrderId) {
      return;
    }

    qsa("[data-order-id]", root).forEach((node) => {
      node.classList.toggle("order-card--highlight", node.getAttribute("data-order-id") === highlightedOrderId);
    });

    if (shouldScroll && !highlightScrollDone) {
      const target = qs("[data-order-id='" + highlightedOrderId + "']", root);
      if (target) {
        target.scrollIntoView({ block: "center", behavior: "smooth" });
        highlightScrollDone = true;
      }
    }
  }

  function isSecureContextLike() {
    return window.isSecureContext || window.location.hostname === "localhost" || window.location.hostname === "127.0.0.1";
  }

  function isBartenderPushContext() {
    const role = document.body.getAttribute("data-role") || "";
    const path = document.body.getAttribute("data-path") || window.location.pathname;
    return role === "BARTENDER" && path.startsWith("/bartender");
  }

  function urlBase64ToUint8Array(base64String) {
    const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
    const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
    const raw = window.atob(base64);
    const bytes = new Uint8Array(raw.length);
    for (let i = 0; i < raw.length; i += 1) {
      bytes[i] = raw.charCodeAt(i);
    }
    return bytes;
  }

  async function ensurePushServiceWorker() {
    if (!("serviceWorker" in navigator)) {
      throw new Error("Service workers are not supported in this browser.");
    }
    return navigator.serviceWorker.register("/sw.js", { scope: "/" });
  }

  function parseSavedDeviceCount(card) {
    const count = parseInt(card.dataset.pushEnabledDevices || "0", 10);
    return Number.isFinite(count) && count > 0 ? count : 0;
  }

  function savedDeviceMessage(card) {
    const count = parseSavedDeviceCount(card);
    if (count === 1) {
      return "Notifications are enabled on 1 saved device for your account.";
    }
    if (count > 1) {
      return "Notifications are enabled on " + count + " saved devices for your account.";
    }
    return "Notifications are not enabled on any saved devices yet.";
  }

  function renderPushCard(card, state) {
    const label = qs("[data-push-state]", card);
    const message = qs("[data-push-message]", card);
    const enableButton = qs("[data-push-enable]", card);
    const disableButton = qs("[data-push-disable]", card);
    const configured = card.dataset.pushConfigured === "1";

    if (label) {
      label.textContent = state.label;
      label.classList.remove(
        "notification-chip--ok",
        "notification-chip--warn",
        "notification-chip--blocked",
        "notification-chip--muted"
      );
      if (state.tone) {
        label.classList.add("notification-chip--" + state.tone);
      }
    }
    if (message) {
      message.textContent = state.message;
    }
    if (enableButton) {
      enableButton.disabled = !!state.busy || !configured || !state.canEnable;
    }
    if (disableButton) {
      disableButton.disabled = !!state.busy || !state.canDisable;
    }
  }

  async function getPushState(card) {
    if (card.dataset.pushConfigured !== "1") {
      return {
        label: "Unavailable",
        message: "Push notifications are not configured on this server yet.",
        tone: "muted",
        canEnable: false,
        canDisable: false,
      };
    }

    if (!isSecureContextLike()) {
      return {
        label: "HTTPS required",
        message: "Notifications need HTTPS (or localhost during development) before this browser can subscribe.",
        tone: "warn",
        canEnable: false,
        canDisable: false,
      };
    }

    if (!("Notification" in window) || !("PushManager" in window) || !("serviceWorker" in navigator)) {
      return {
        label: "Unsupported",
        message: "This browser does not support the Web Push features required for bartender alerts.",
        tone: "warn",
        canEnable: false,
        canDisable: false,
      };
    }

    const registration = await ensurePushServiceWorker();
    const subscription = await registration.pushManager.getSubscription();
    if (subscription) {
      if (Notification.permission === "denied") {
        return {
          label: "Blocked in browser",
          message: "Browser permission is blocked, but this saved device can still be removed here.",
          tone: "blocked",
          canEnable: false,
          canDisable: true,
        };
      }
      return {
        label: "Enabled on this device",
        message: savedDeviceMessage(card),
        tone: "ok",
        canEnable: false,
        canDisable: true,
      };
    }

    if (Notification.permission === "denied") {
      return {
        label: "Blocked",
        message: "Notifications are blocked for this site. Re-enable them in browser settings and then try again.",
        tone: "blocked",
        canEnable: false,
        canDisable: false,
      };
    }

    return {
      label: Notification.permission === "granted" ? "Ready to enable" : "Disabled on this device",
      message: savedDeviceMessage(card),
      tone: "muted",
      canEnable: true,
      canDisable: false,
    };
  }

  async function syncPushCard(card, fallbackMessage) {
    try {
      const state = await getPushState(card);
      if (fallbackMessage) {
        state.message = fallbackMessage;
      }
      renderPushCard(card, state);
    } catch (err) {
      renderPushCard(card, {
        label: "Unavailable",
        message: "Could not read this browser's notification state right now.",
        tone: "warn",
        canEnable: false,
        canDisable: false,
      });
    }
  }

  async function enablePushNotifications(card) {
    renderPushCard(card, {
      label: "Enabling...",
      message: "Requesting permission and saving this device.",
      tone: "muted",
      canEnable: false,
      canDisable: false,
      busy: true,
    });

    try {
      const registration = await ensurePushServiceWorker();
      const hadSubscription = !!(await registration.pushManager.getSubscription());

      let permission = Notification.permission;
      if (permission !== "granted") {
        permission = await Notification.requestPermission();
      }
      if (permission !== "granted") {
        await syncPushCard(card);
        return;
      }

      let subscription = await registration.pushManager.getSubscription();
      if (!subscription) {
        subscription = await registration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: urlBase64ToUint8Array(card.dataset.pushPublicKey || ""),
        });
      }

      const response = await jsonFetch("/bartender/notifications/subscribe", {
        method: "POST",
        body: JSON.stringify(subscription.toJSON()),
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(data.error || "Could not enable notifications on this device.");
      }

      if (!hadSubscription) {
        const nextCount = Math.max(parseSavedDeviceCount(card), 0) + 1;
        card.dataset.pushEnabledDevices = String(nextCount);
      }

      await syncPushCard(card, "Notifications are enabled on this device.");
    } catch (err) {
      renderPushCard(card, {
        label: "Could not enable",
        message: err && err.message ? err.message : "Could not enable notifications on this device.",
        tone: "warn",
        canEnable: true,
        canDisable: false,
      });
    }
  }

  async function disablePushNotifications(card) {
    renderPushCard(card, {
      label: "Disabling...",
      message: "Removing this device from bartender alerts.",
      tone: "muted",
      canEnable: false,
      canDisable: false,
      busy: true,
    });

    try {
      const registration = await ensurePushServiceWorker();
      const subscription = await registration.pushManager.getSubscription();
      let serverWarning = "";

      if (subscription) {
        const response = await jsonFetch("/bartender/notifications/unsubscribe", {
          method: "POST",
          body: JSON.stringify({ endpoint: subscription.endpoint }),
        });
        const data = await response.json().catch(() => ({}));
        if (!response.ok) {
          serverWarning = data.error || "This browser unsubscribed, but the server could not confirm removal yet.";
        }

        try {
          await subscription.unsubscribe();
        } catch (err) {
          if (!serverWarning) {
            throw err;
          }
        }

        const nextCount = Math.max(parseSavedDeviceCount(card) - 1, 0);
        card.dataset.pushEnabledDevices = String(nextCount);
      }

      await syncPushCard(card, serverWarning || "Notifications are disabled on this device.");
    } catch (err) {
      renderPushCard(card, {
        label: "Could not disable",
        message: err && err.message ? err.message : "Could not disable notifications on this device.",
        tone: "warn",
        canEnable: true,
        canDisable: true,
      });
    }
  }

  function wirePushCard(root = document) {
    collect("[data-push-card]", root).forEach((card) => {
      if (card.dataset.bound === "1") {
        return;
      }
      card.dataset.bound = "1";

      const enableButton = qs("[data-push-enable]", card);
      const disableButton = qs("[data-push-disable]", card);

      if (enableButton) {
        enableButton.addEventListener("click", () => {
          enablePushNotifications(card);
        });
      }
      if (disableButton) {
        disableButton.addEventListener("click", () => {
          disablePushNotifications(card);
        });
      }

      syncPushCard(card);
    });
  }

  function wirePushSupport() {
    if (!isBartenderPushContext() || !isSecureContextLike() || !("serviceWorker" in navigator)) {
      return;
    }
    ensurePushServiceWorker().catch((err) => {
      console.warn("push service worker registration failed", err);
    });
  }

  function dismissFlash(item) {
    if (!item) {
      return;
    }
    item.remove();
    const container = qs("[data-flash-container]");
    if (container && container.children.length === 0) {
      const wrapper = container.parentElement;
      if (wrapper) {
        wrapper.remove();
      }
    }
  }

  function wireFlashes(root = document) {
    collect("[data-flash-item]", root).forEach((item) => {
      if (item.dataset.bound === "1") {
        return;
      }
      item.dataset.bound = "1";

      const closeButton = qs("[data-flash-dismiss]", item);
      if (closeButton) {
        closeButton.addEventListener("click", () => dismissFlash(item));
      }

      const timeoutMs = item.classList.contains("text-red-800") ? 8000 : 5000;
      window.setTimeout(() => dismissFlash(item), timeoutMs);
    });
  }

  function wireInventoryFilters(root = document) {
    const controls = getInventoryControls(root);
    const host = qs("#productsTable");
    if (!controls || !host) {
      return;
    }

    if (controls.dataset.inventoryBound !== "1") {
      controls.dataset.inventoryBound = "1";
      controls.addEventListener("submit", (evt) => {
        evt.preventDefault();
        refreshInventoryResults().catch((err) => {
          console.warn("inventory submit error", err);
        });
      });
    }

    const searchInput = qs("[data-inventory-search]", controls);

    if (searchInput && searchInput.dataset.inventorySearchBound !== "1") {
      searchInput.dataset.inventorySearchBound = "1";
      const handleSearch = debounce(() => {
        refreshInventoryResults().catch((err) => {
          console.warn("inventory search error", err);
        });
      }, 250);
      searchInput.addEventListener("input", handleSearch);
      searchInput.addEventListener("search", handleSearch);
    }
  }

  function applyShellSearch() {
    const input = qs("[data-shell-search-input]");
    const query = input ? String(input.value || "").trim().toLowerCase() : "";
    const items = qsa("[data-shell-search-item]");
    let visibleCount = 0;

    items.forEach((node) => {
      const haystack = String(node.getAttribute("data-shell-search-item") || node.textContent || "").toLowerCase();
      const match = query === "" || haystack.includes(query);
      node.hidden = !match;
      if (match) {
        visibleCount += 1;
      }
    });

    qsa("[data-shell-search-empty-state]").forEach((node) => {
      node.hidden = query === "" || visibleCount > 0;
    });
  }

  function wireShellSearch() {
    const input = qs("[data-shell-search-input]");
    if (!input) {
      return;
    }
    if (input.dataset.bound === "1") {
      if (!isLibrarySearchPath(currentPath())) {
        applyShellSearch();
      }
      return;
    }

    input.dataset.bound = "1";
    if (isLibrarySearchPath(currentPath()) && qs("#libraryResults")) {
      const syncLibrarySearch = debounce(() => {
        refreshLibraryResults({
          q: input.value,
          page: "",
        }).catch((err) => {
          console.warn("library search error", err);
        });
      }, 250);

      input.addEventListener("input", syncLibrarySearch);
      input.addEventListener("search", syncLibrarySearch);
      return;
    }

    input.addEventListener("input", applyShellSearch);
    applyShellSearch();
  }

  function syncSelectProxyButtons() {
    qsa("[data-select-proxy]").forEach((button) => {
      const selector = button.getAttribute("data-select-proxy");
      const target = selector ? qs(selector) : null;
      const active = !!target && target.value === button.getAttribute("data-select-proxy-value");
      button.classList.toggle("is-active", active);
      button.setAttribute("aria-pressed", String(active));
    });
  }

  function wireSelectProxy(root = document) {
    collect("[data-select-proxy]", root).forEach((button) => {
      if (button.dataset.selectProxyBound === "1") {
        return;
      }
      button.dataset.selectProxyBound = "1";

      button.addEventListener("click", () => {
        const selector = button.getAttribute("data-select-proxy");
        const target = selector ? qs(selector) : null;
        if (!target) {
          return;
        }

        target.value = button.getAttribute("data-select-proxy-value") || "";
        target.dispatchEvent(new Event("change", { bubbles: true }));
        syncSelectProxyButtons();
      });
    });

    syncSelectProxyButtons();
  }

  function initUI(root = document) {
    wireHX(root);
    wireCocktailViewToggle(root);
    wireIngredientEditors(root);
    wirePushCard(root);
    wireFlashes(root);
    wireInventoryFilters(root);
    wireShellSearch();
    wireSelectProxy(root);
  }

  document.addEventListener("DOMContentLoaded", () => {
    applyTheme(getActiveTheme());
    wireThemeToggle();
    initUI(document);
    wireSSE();
    applyHighlightedOrder(document, true);

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


