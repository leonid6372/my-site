import { describe, expect, it } from 'vitest';
import { fixtures } from './server';

describe('test harness', () => {
  it('serves the fixtures through MSW with the requested language', async () => {
    const res = await fetch('/api/v1/main', { headers: { 'Accept-Language': 'ru' } });
    expect(res.headers.get('content-language')).toBe('ru');
    expect(await res.json()).toEqual(fixtures.main.ru);
  });

  it('has a 30-photo place and a 0-photo place in the travel fixture', () => {
    const counts = fixtures.travel.en.places.map((p) => p.photos.length);
    expect(counts).toContain(30);
    expect(counts).toContain(0);
  });
});
