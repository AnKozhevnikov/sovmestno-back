// Сценарий 1: публичный каталог без авторизации
// Запуск: k6 run scenario1-public.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const BASE_URL = 'https://api.sovmestno-test.ru';

const eventDuration = new Trend('event_list_duration', true);
const venueDuration = new Trend('venue_list_duration', true);
const errorRate = new Rate('errors');

export const options = {
  vus: 5,
  duration: '2m',
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed:   ['rate<0.05'],
    errors:            ['rate<0.05'],
  },
};

export default function () {
  // Листинг мероприятий
  const eventsRes = http.get(`${BASE_URL}/api/event/public/events?limit=20`);
  eventDuration.add(eventsRes.timings.duration);
  const eventsOk = check(eventsRes, {
    'events list 200':      r => r.status === 200,
    'events list has data': r => {
      try { return r.json().events.length > 0; } catch { return false; }
    },
  });
  errorRate.add(!eventsOk);

  sleep(0.5);

  let eventID = 1;
  try {
    const events = eventsRes.json().events;
    if (events && events.length > 0) {
      eventID = events[Math.floor(Math.random() * events.length)].id;
    }
  } catch {}

  // Одно мероприятие
  const eventRes = http.get(`${BASE_URL}/api/event/public/events/${eventID}`);
  check(eventRes, { 'single event 200': r => r.status === 200 });
  errorRate.add(eventRes.status !== 200);

  sleep(0.5);

  // Листинг площадок
  const venuesRes = http.get(`${BASE_URL}/api/user/public/venues?limit=20`);
  venueDuration.add(venuesRes.timings.duration);
  check(venuesRes, { 'venues list 200': r => r.status === 200 });
  errorRate.add(venuesRes.status !== 200);

  sleep(1);
}
