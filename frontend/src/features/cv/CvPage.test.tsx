import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { MemoryRouter } from 'react-router';
import { describe, expect, it } from 'vitest';
import type { Cv, Lang } from '../../api/types';
import { renderWithQuery } from '../../test/render';
import { failEndpoint, fixtures, server } from '../../test/server';
import CvPage from './CvPage';

function renderCv(lang: Lang = 'en') {
  return renderWithQuery(
    <MemoryRouter>
      <CvPage />
    </MemoryRouter>,
    { lang },
  );
}

const sectionHeadings = () =>
  screen.getAllByRole('heading', { level: 2 }).map((h) => h.textContent);

describe('CvPage', () => {
  it('renders the sections in spec order', async () => {
    renderCv();
    expect(await screen.findByRole('heading', { level: 1 })).toHaveTextContent(
      fixtures.main.en.fullName,
    );
    expect(sectionHeadings()).toEqual([
      'Professional summary',
      'Experience',
      'Education',
      'Skills',
      'Languages',
    ]);
  });

  it('renders the RU fixture with RU headings and dates', async () => {
    const cv = fixtures.cv.ru;
    renderCv('ru');
    await screen.findByRole('heading', { level: 1 });
    expect(sectionHeadings()).toEqual([
      'Профессиональное резюме',
      'Опыт работы',
      'Образование',
      'Навыки',
      'Языки',
    ]);
    expect(screen.getByText(cv.summary)).toBeInTheDocument();
    for (const e of cv.experience) {
      expect(screen.getByRole('heading', { level: 3, name: e.position })).toBeInTheDocument();
    }
    expect(screen.getByText(/н\. в\./)).toBeInTheDocument();
  });

  it('shows "Present" and a duration for a null endDate', async () => {
    renderCv();
    const current = fixtures.cv.en.experience.find((e) => e.endDate === null);
    expect(current).toBeDefined();
    const item = (await screen.findByRole('heading', { level: 3, name: current?.position }))
      .parentElement;
    expect(item).toHaveTextContent('May 2021 – Present');
    expect(item).toHaveTextContent(/Present\W+\d+ yrs?/);
  });

  it('links the company only when companyUrl is set; hides a null description', async () => {
    const cv: Cv = {
      ...fixtures.cv.en,
      experience: [
        {
          id: 'with-url',
          position: 'With URL',
          company: 'Linked Co',
          companyUrl: 'https://linked.example',
          location: 'Remote',
          startDate: '2022-01-01',
          endDate: null,
          projects: [
            { description: 'Has a description.', highlights: ['One'] },
            { description: 'Second project.', highlights: ['Two', 'Three'] },
          ],
        },
        {
          id: 'no-url',
          position: 'Without URL',
          company: 'Plain Co',
          companyUrl: null,
          location: 'Remote',
          startDate: '2019-01-01',
          endDate: '2021-12-31',
          projects: [{ description: null, highlights: [] }],
        },
      ],
    };
    server.use(http.get('*/api/v1/cv', () => HttpResponse.json(cv)));
    renderCv();

    const linked = await screen.findByRole('link', { name: /Linked Co/ });
    expect(linked).toHaveAttribute('href', 'https://linked.example');
    expect(linked).toHaveAttribute('target', '_blank');
    expect(linked).toHaveAttribute('data-print-url', 'https://linked.example');

    expect(screen.queryByRole('link', { name: /Plain Co/ })).toBeNull();
    const plain = screen.getByRole('heading', { level: 3, name: 'Without URL' }).parentElement;
    if (!plain) throw new Error('no item');
    expect(within(plain).getByText('Plain Co')).toBeInTheDocument();
    expect(within(plain).queryAllByRole('listitem')).toHaveLength(0); // no highlights list
    expect(plain.querySelectorAll('p')).toHaveLength(2); // meta + period, no description
    expect(screen.getByText('Has a description.')).toBeInTheDocument();
  });

  it('shows every project of a job with its own highlights', async () => {
    const cv: Cv = {
      ...fixtures.cv.en,
      experience: [
        {
          id: 'multi',
          position: 'Multi',
          company: 'Multi Co',
          companyUrl: null,
          location: 'Remote',
          startDate: '2022-01-01',
          endDate: null,
          projects: [
            { description: 'First project.', highlights: ['A1'] },
            { description: 'Second project.', highlights: ['B1', 'B2'] },
          ],
        },
      ],
    };
    server.use(http.get('*/api/v1/cv', () => HttpResponse.json(cv)));
    renderCv();

    const item = (await screen.findByRole('heading', { level: 3, name: 'Multi' })).parentElement;
    if (!item) throw new Error('no item');
    const lists = within(item).getAllByRole('list');
    expect(lists).toHaveLength(2);
    const [first, second] = lists;
    if (!first || !second) throw new Error('no lists');
    const texts = (ul: HTMLElement) =>
      within(ul)
        .getAllByRole('listitem')
        .map((li) => li.textContent);
    expect(texts(first)).toEqual(['A1']);
    expect(texts(second)).toEqual(['B1', 'B2']);
    // each description sits right before its own list
    expect(first.previousElementSibling).toHaveTextContent('First project.');
    expect(second.previousElementSibling).toHaveTextContent('Second project.');
  });

  it('prints contact targets: plain email, URLs for the rest', async () => {
    renderCv();
    await screen.findByRole('heading', { level: 1 });
    for (const l of fixtures.main.en.links) {
      const a = screen.getByRole('link', { name: new RegExp(`^${l.label}`) });
      expect(a).toHaveAttribute(
        'data-print-url',
        l.kind === 'email' ? l.url.replace('mailto:', '') : l.url,
      );
    }
  });

  it('shows the error state when either request fails, and Retry recovers', async () => {
    failEndpoint('cv');
    renderCv();
    const alert = await screen.findByRole('alert');
    server.resetHandlers();
    await userEvent.click(within(alert).getByRole('button', { name: 'Retry' }));
    expect(await screen.findByRole('heading', { level: 1 })).toBeInTheDocument();
  });
});
