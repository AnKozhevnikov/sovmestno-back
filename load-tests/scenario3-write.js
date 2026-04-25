// Сценарий 3: write-операции — регистрация, создание мероприятия, заявка
// Запуск: k6 run scenario3-write.js
// Очистка после: DELETE FROM users WHERE email LIKE '%@loadtest-write.ru';

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const BASE_URL = 'https://api.sovmestno-test.ru';

const registerDuration    = new Trend('register_duration', true);
const createEventDuration = new Trend('create_event_duration', true);
const applyDuration       = new Trend('apply_duration', true);
const errorRate           = new Rate('errors');

export const options = {
  vus: 3,
  duration: '2m',
  thresholds: {
    http_req_duration:     ['p(95)<2000'],
    http_req_failed:       ['rate<0.05'],
    errors:                ['rate<0.05'],
    register_duration:     ['p(95)<2000'],
    create_event_duration: ['p(95)<1500'],
    apply_duration:        ['p(95)<1500'],
  },
};

export default function () {
  const uid = `${__VU}_${__ITER}_${Date.now()}`;

  // Регистрируем creator
  const creatorRes = http.post(`${BASE_URL}/api/user/auth/register/creator`, JSON.stringify({
    email:    `creator_${uid}@loadtest-write.ru`,
    password: 'loadtest123',
    name:     `Creator ${uid}`,
  }), { headers: { 'Content-Type': 'application/json' } });

  registerDuration.add(creatorRes.timings.duration);
  const creatorOk = check(creatorRes, { 'creator registered': r => r.status === 200 || r.status === 201 });
  errorRate.add(!creatorOk);
  if (!creatorOk) { sleep(2); return; }

  const creatorToken    = creatorRes.json().access_token;
  const creatorUserID   = creatorRes.json().user.id;
  const creatorHeaders  = { 'Authorization': `Bearer ${creatorToken}`, 'Content-Type': 'application/json' };

  sleep(0.3);

  // Регистрируем venue
  const venueRes = http.post(`${BASE_URL}/api/user/auth/register/venue`, JSON.stringify({
    email:    `venue_${uid}@loadtest-write.ru`,
    password: 'loadtest123',
    name:     `Venue ${uid}`,
    capacity: 100,
  }), { headers: { 'Content-Type': 'application/json' } });

  registerDuration.add(venueRes.timings.duration);
  const venueOk = check(venueRes, { 'venue registered': r => r.status === 200 || r.status === 201 });
  errorRate.add(!venueOk);
  if (!venueOk) { sleep(2); return; }

  const venueToken   = venueRes.json().access_token;
  const venueHeaders = { 'Authorization': `Bearer ${venueToken}`, 'Content-Type': 'application/json' };

  sleep(0.3);

  // Creator создаёт мероприятие
  const eventRes = http.post(`${BASE_URL}/api/event/events`, JSON.stringify({
    title:        `Мероприятие от ${uid}`,
    description:  `Тестовое мероприятие, итерация ${uid}`,
  }), { headers: creatorHeaders });

  createEventDuration.add(eventRes.timings.duration);
  const eventOk = check(eventRes, { 'event created': r => r.status === 200 || r.status === 201 });
  errorRate.add(!eventOk);
  if (!eventOk) {
    console.log('FAIL event: status=' + eventRes.status + ' body=' + eventRes.body);
    sleep(2); return;
  }

  const eventID = eventRes.json().id;

  sleep(0.3);

  // Venue подаёт заявку
  const appRes = http.post(`${BASE_URL}/api/application/applications`, JSON.stringify({
    receiver_id:   creatorUserID,
    receiver_type: 'creator',
    event_id:      eventID,
    message:       'Хотим провести ваше мероприятие у нас',
  }), { headers: venueHeaders });

  applyDuration.add(appRes.timings.duration);
  const appOk = check(appRes, { 'application created': r => r.status === 200 || r.status === 201 });
  errorRate.add(!appOk);
  if (!appOk) { sleep(2); return; }

  sleep(0.3);

  // Creator принимает заявку
  const acceptRes = http.patch(
    `${BASE_URL}/api/application/applications/${appRes.json().id}/accept`,
    null,
    { headers: creatorHeaders }
  );
  check(acceptRes, { 'application accepted': r => r.status === 200 });
  errorRate.add(acceptRes.status !== 200);

  sleep(1);
}
