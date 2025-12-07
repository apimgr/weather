// Service Worker for Weather PWA
const CACHE_NAME = 'weather-v1';
const RUNTIME_CACHE = 'weather-runtime';

// Assets to cache on install
const PRECACHE_URLS = [
  '/',
  '/static/css/dracula.css',
  '/static/manifest.json'
];

// Install event - cache essential assets
self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => cache.addAll(PRECACHE_URLS))
      .then(() => self.skipWaiting())
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys().then(cacheNames => {
      return Promise.all(
        cacheNames
          .filter(name => name !== CACHE_NAME && name !== RUNTIME_CACHE)
          .map(name => caches.delete(name))
      );
    }).then(() => self.clients.claim())
  );
});

// Fetch event - network first, fallback to cache
self.addEventListener('fetch', event => {
  // Skip cross-origin requests
  if (!event.request.url.startsWith(self.location.origin)) {
    return;
  }

  // Network first strategy for API calls
  if (event.request.url.includes('/api/')) {
    event.respondWith(
      fetch(event.request)
        .then(response => {
          // Cache successful API responses
          if (response.ok) {
            const responseClone = response.clone();
            caches.open(RUNTIME_CACHE).then(cache => {
              cache.put(event.request, responseClone);
            });
          }
          return response;
        })
        .catch(() => {
          // Fallback to cache on network error
          return caches.match(event.request);
        })
    );
    return;
  }

  // Cache first strategy for static assets
  event.respondWith(
    caches.match(event.request)
      .then(cachedResponse => {
        if (cachedResponse) {
          return cachedResponse;
        }

        return fetch(event.request).then(response => {
          // Cache successful responses
          if (response.ok && event.request.method === 'GET') {
            const responseClone = response.clone();
            caches.open(RUNTIME_CACHE).then(cache => {
              cache.put(event.request, responseClone);
            });
          }
          return response;
        });
      })
  );
});

// Background sync for offline actions (future enhancement)
self.addEventListener('sync', event => {
  if (event.tag === 'sync-weather') {
    event.waitUntil(syncWeatherData());
  }
});

async function syncWeatherData() {
  // Placeholder for syncing saved locations when back online
  console.log('Syncing weather data...');
}

// Push notifications (future enhancement)
self.addEventListener('push', event => {
  if (!event.data) return;

  const data = event.data.json();
  const options = {
    body: data.body || 'Weather update available',
    icon: '/static/images/icon-192.png',
    badge: '/static/images/icon-72.png',
    data: data.url || '/'
  };

  event.waitUntil(
    self.registration.showNotification(data.title || 'Weather Alert', options)
  );
});

// Notification click handler
self.addEventListener('notificationclick', event => {
  event.notification.close();
  event.waitUntil(
    clients.openWindow(event.notification.data)
  );
});
