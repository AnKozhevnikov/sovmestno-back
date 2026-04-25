// Заливка тестовых данных на staging
// Запуск: k6 run seed.js

import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'https://api.sovmestno-test.ru';

const ADMIN_EMAIL = 'admin@sovmestno.ru';
const ADMIN_PASSWORD = 'loadtest_admin123';
const ADMIN_SECRET = 'АДМИНСКИЙ_СЕКРЕТ_СЮДА';

const CATEGORIES = ['Музыка', 'Театр', 'Выставка', 'Фестиваль', 'Спорт', 'Кино', 'Лекция', 'Мастер-класс'];

export const options = {
  vus: 1,
  iterations: 1,
};

function registerCreator(i) {
  const res = http.post(`${BASE_URL}/api/user/auth/register/creator`, JSON.stringify({
    email: `creator${i}@loadtest.ru`,
    password: 'loadtest123',
    name: `Тестовый Организатор ${i}`,
    description: `Организатор мероприятий номер ${i} для нагрузочного теста`,
  }), { headers: { 'Content-Type': 'application/json' } });

  check(res, { 'creator registered': r => r.status === 200 || r.status === 201 });
  return res.status === 200 || res.status === 201 ? res.json() : null;
}

function registerVenue(i) {
  const res = http.post(`${BASE_URL}/api/user/auth/register/venue`, JSON.stringify({
    email: `venue${i}@loadtest.ru`,
    password: 'loadtest123',
    name: `Тестовая Площадка ${i}`,
    description: `Площадка номер ${i} для нагрузочного теста`,
    street_address: `ул. Тестовая, д. ${i}`,
    capacity: 50 + i * 10,
  }), { headers: { 'Content-Type': 'application/json' } });

  check(res, { 'venue registered': r => r.status === 200 || r.status === 201 });
  return res.status === 200 || res.status === 201 ? res.json() : null;
}

function withRetry(fn) {
  for (let attempt = 1; attempt <= 3; attempt++) {
    const res = fn();
    if (res.status !== 503) return res;
    console.log('503 — retry ' + attempt + '/3');
    sleep(2);
  }
  return fn();
}

function createEvent(token, categoryIDs, i) {
  const res = withRetry(() => http.post(`${BASE_URL}/api/event/events`, JSON.stringify({
    title: `Тестовое мероприятие ${i}`,
    description: `Описание мероприятия ${i} для нагрузочного теста. `.repeat(5),
    category_ids: categoryIDs,
  }), { headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` } }));

  const ok = check(res, { 'event created': r => r.status === 200 || r.status === 201 });
  if (!ok) {
    console.log('FAIL createEvent #' + i + ' status=' + res.status);
  }
  return ok ? res.json() : null;
}

function publishEvent(token, eventID) {
  const res = withRetry(() => http.patch(`${BASE_URL}/api/event/events/${eventID}/publish`, null,
    { headers: { 'Authorization': `Bearer ${token}` } }));
  const ok = check(res, { 'event published': r => r.status === 200 || r.status === 204 });
  if (!ok) {
    console.log('FAIL publishEvent id=' + eventID + ' status=' + res.status);
  }
}

function setupAdmin() {
  // Регистрируем admin (если уже существует — логинимся)
  let res = http.post(`${BASE_URL}/api/user/auth/register/admin`, JSON.stringify({
    email: ADMIN_EMAIL,
    password: ADMIN_PASSWORD,
    admin_secret: ADMIN_SECRET,
  }), { headers: { 'Content-Type': 'application/json' } });

  if (res.status !== 200 && res.status !== 201) {
    res = http.post(`${BASE_URL}/api/user/auth/login`, JSON.stringify({
      email: ADMIN_EMAIL,
      password: ADMIN_PASSWORD,
    }), { headers: { 'Content-Type': 'application/json' } });
  }

  check(res, { 'admin ready': r => r.status === 200 || r.status === 201 });
  return res.json().access_token;
}

function seedCategories(adminToken) {
  const existing = http.get(`${BASE_URL}/api/event/categories`).json();
  if (existing && existing.length >= CATEGORIES.length) {
    console.log(`Категории уже есть (${existing.length}), пропускаем`);
    return existing.map(c => c.id);
  }

  const ids = [];
  for (const name of CATEGORIES) {
    const res = http.post(`${BASE_URL}/api/event/categories`, JSON.stringify({ name }),
      { headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${adminToken}` } });
    check(res, { 'category created': r => r.status === 200 || r.status === 201 });
    if (res.status === 200 || res.status === 201) {
      ids.push(res.json().id);
    }
  }
  console.log(`Создано категорий: ${ids.length}`);
  return ids;
}

export default function () {
  console.log('=== Заливка тестовых данных ===');

  // Создаём категории через admin
  const adminToken = setupAdmin();
  const categoryIDs = seedCategories(adminToken);

  // Регистрируем 10 creators
  console.log('Регистрируем creators...');
  const creators = [];
  for (let i = 1; i <= 10; i++) {
    const data = registerCreator(i);
    if (data && data.access_token) {
      creators.push({ token: data.access_token, user: data.user });
    }
  }
  console.log(`Создано creators: ${creators.length}`);

  // Регистрируем 10 venues
  console.log('Регистрируем venues...');
  for (let i = 1; i <= 10; i++) {
    registerVenue(i);
  }
  console.log('Создано venues: 10');

  // Каждый creator создаёт по 20 мероприятий
  console.log('Создаём мероприятия...');
  let eventCount = 0;
  for (const creator of creators) {
    for (let i = 1; i <= 20; i++) {
      const shuffled = categoryIDs.slice().sort(() => Math.random() - 0.5);
      const categoryCount = Math.floor(Math.random() * 3) + 1;
      const ids = shuffled.slice(0, categoryCount);
      const event = createEvent(creator.token, ids, eventCount + 1);
      if (event && event.id) {
        publishEvent(creator.token, event.id);
        eventCount++;
      }
    }
  }

  console.log(`=== Готово: ${eventCount} мероприятий, ${creators.length} creators, 10 venues ===`);
}
