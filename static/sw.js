const DEFAULT_URL = "/bartender/orders";
const DEFAULT_ICON = "/static/icons/icon-192.png";
const DEFAULT_BADGE = "/static/icons/badge-72.png";

self.addEventListener("install", (event) => {
  event.waitUntil(self.skipWaiting());
});

self.addEventListener("activate", (event) => {
  event.waitUntil(self.clients.claim());
});

self.addEventListener("push", (event) => {
  let payload = {};

  if (event.data) {
    try {
      payload = event.data.json();
    } catch (err) {
      payload = {
        body: event.data.text(),
      };
    }
  }

  const title = typeof payload.title === "string" && payload.title ? payload.title : "House Bartender";
  const body = typeof payload.body === "string" && payload.body ? payload.body : "A new order is waiting in the bartender queue.";
  const url = typeof payload.url === "string" && payload.url ? payload.url : DEFAULT_URL;
  const tag = typeof payload.tag === "string" && payload.tag ? payload.tag : "bartender-order";
  const timestamp = Number(payload.timestamp) || Date.now();

  event.waitUntil(
    self.registration.showNotification(title, {
      body,
      tag,
      renotify: true,
      icon: DEFAULT_ICON,
      badge: DEFAULT_BADGE,
      timestamp,
      data: {
        url,
        orderId: payload.order_id || null,
        timestamp,
      },
    })
  );
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();

  const targetURL = new URL(
    (event.notification.data && event.notification.data.url) || DEFAULT_URL,
    self.location.origin
  ).href;

  event.waitUntil(
    self.clients.matchAll({ type: "window", includeUncontrolled: true }).then(async (clients) => {
      for (const client of clients) {
        if (!client.url || !client.url.startsWith(self.location.origin)) {
          continue;
        }

        try {
          if ("focus" in client) {
            await client.focus();
          }
          if ("navigate" in client) {
            const navigated = await client.navigate(targetURL);
            if (navigated && "focus" in navigated) {
              await navigated.focus();
            }
          }
          return;
        } catch (err) {
          // Fall back to opening a new window below.
        }
      }

      return self.clients.openWindow(targetURL);
    })
  );
});
