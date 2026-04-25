// Сценарий 2: авторизованный пользователь
// Запуск: k6 run scenario2-auth.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const BASE_URL = 'https://api.sovmestno-test.ru';

const loginDuration  = new Trend('login_duration', true);
const eventsDuration = new Trend('auth_events_duration', true);
const errorRate      = new Rate('errors');

export const options = {
  vus: 5,
  duration: '2m',
  thresholds: {
    http_req_duration:    ['p(95)<1000'],
    http_req_failed:      ['rate<0.05'],
    errors:               ['rate<0.05'],
    login_duration:       ['p(95)<1500'],
    auth_events_duration: ['p(95)<1000'],
  },
};

export default function () {
  const userIndex = (__VU % 10) + 1;

  // Логин
  const loginRes = http.post(`${BASE_URL}/api/user/auth/login`, JSON.stringify({
    email:    `creator${userIndex}@loadtest.ru`,
    password: 'loadtest123',
  }), { headers: { 'Content-Type': 'application/json' } });

  loginDuration.add(loginRes.timings.duration);
  const loginOk = check(loginRes, {
    'login 200':        r => r.status === 200,
    'has access_token': r => { try { return !!r.json().access_token; } catch { return false; } },
  });
  errorRate.add(!loginOk);

  if (!loginOk) { sleep(2); return; }

  const headers = { 'Authorization': `Bearer ${loginRes.json().access_token}`, 'Content-Type': 'application/json' };

  sleep(0.3);

  // Листинг мероприятий
  const eventsRes = http.get(`${BASE_URL}/api/event/events?limit=20`, { headers });
  eventsDuration.add(eventsRes.timings.duration);
  check(eventsRes, { 'auth events 200': r => r.status === 200 });
  errorRate.add(eventsRes.status !== 200);

  sleep(0.5);

  // Одно мероприятие
  let eventID = 1;
  try {
    const events = eventsRes.json().events;
    if (events && events.length > 0) {
      eventID = events[Math.floor(Math.random() * events.length)].id;
    }
  } catch {}

  const eventRes = http.get(`${BASE_URL}/api/event/events/${eventID}`, { headers });
  check(eventRes, { 'single event 200': r => r.status === 200 });
  errorRate.add(eventRes.status !== 200);

  sleep(0.5);

  // Список площадок
  const venuesRes = http.get(`${BASE_URL}/api/user/users/venues?limit=20`, { headers });
  check(venuesRes, { 'venues 200': r => r.status === 200 });
  errorRate.add(venuesRes.status !== 200);

  sleep(1);
}
