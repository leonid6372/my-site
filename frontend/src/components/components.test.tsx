import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Photo } from '../api/types';
import { ChipList } from './Chip';
import { ErrorState } from './ErrorState';
import { RemoteImage } from './RemoteImage';

const full: Photo = {
  url: 'https://img.example.com/full.jpg',
  thumbUrl: 'https://img.example.com/thumb.jpg',
  alt: 'Canal at night',
  caption: 'Amsterdam',
  width: 1600,
  height: 1067,
};
const bare: Photo = {
  url: 'https://img.example.com/bare.jpg',
  thumbUrl: null,
  alt: null,
  caption: null,
  width: null,
  height: null,
};

function frameOf(img: HTMLElement) {
  const frame = img.closest('[data-status]');
  if (!(frame instanceof HTMLElement)) throw new Error('no frame');
  return frame;
}

describe('RemoteImage', () => {
  it('uses thumbUrl for thumbs and url for full and portrait', () => {
    const { rerender } = render(<RemoteImage photo={full} variant="thumb" />);
    expect(screen.getByRole('img')).toHaveAttribute('src', full.thumbUrl);
    rerender(<RemoteImage photo={full} variant="full" />);
    expect(screen.getByRole('img')).toHaveAttribute('src', full.url);
    rerender(<RemoteImage photo={full} variant="portrait" />);
    expect(screen.getByRole('img')).toHaveAttribute('src', full.url);
  });

  it('falls back to url when thumbUrl is null', () => {
    render(<RemoteImage photo={bare} variant="thumb" />);
    expect(screen.getByRole('presentation')).toHaveAttribute('src', bare.url);
  });

  it('sets aspect-ratio only when width and height are known', () => {
    const { rerender } = render(<RemoteImage photo={full} variant="full" />);
    expect(frameOf(screen.getByRole('img'))).toHaveStyle({ aspectRatio: '1600 / 1067' });
    rerender(<RemoteImage photo={bare} variant="full" />);
    expect(frameOf(screen.getByRole('presentation')).style.aspectRatio).toBe('');
    rerender(<RemoteImage photo={full} variant="thumb" aspect="square" />);
    expect(frameOf(screen.getByRole('img')).style.aspectRatio).toBe('');
  });

  it('shows a placeholder instead of a broken image on error, keeping the alt', () => {
    render(<RemoteImage photo={full} variant="thumb" />);
    const img = screen.getByRole('img');
    expect(frameOf(img)).toHaveAttribute('data-status', 'loading');
    fireEvent.error(img);
    expect(document.querySelector('img')).toBeNull();
    expect(screen.getByRole('img', { name: 'Canal at night' })).toBeInTheDocument();
  });

  it('labels the error placeholder when there is no alt text', () => {
    render(<RemoteImage photo={bare} variant="thumb" />);
    fireEvent.error(screen.getByRole('presentation'));
    expect(screen.getByRole('img', { name: 'Image unavailable' })).toBeInTheDocument();
  });

  it('marks itself loaded on load', () => {
    render(<RemoteImage photo={full} variant="thumb" />);
    const img = screen.getByRole('img');
    fireEvent.load(img);
    expect(frameOf(img)).toHaveAttribute('data-status', 'loaded');
  });

  it('alt chain: alt → caption → ""', () => {
    const { rerender } = render(<RemoteImage photo={full} variant="thumb" />);
    expect(screen.getByRole('img')).toHaveAttribute('alt', 'Canal at night');
    rerender(<RemoteImage photo={{ ...full, alt: null }} variant="thumb" />);
    expect(screen.getByRole('img')).toHaveAttribute('alt', 'Amsterdam');
    rerender(<RemoteImage photo={bare} variant="thumb" />);
    expect(screen.getByRole('presentation')).toHaveAttribute('alt', '');
  });

  it('sends no referrer, decodes async, lazy except the portrait', () => {
    const { rerender } = render(<RemoteImage photo={full} variant="thumb" />);
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('referrerpolicy', 'no-referrer');
    expect(img).toHaveAttribute('decoding', 'async');
    expect(img).toHaveAttribute('loading', 'lazy');
    expect(img).not.toHaveAttribute('fetchpriority');
    rerender(<RemoteImage photo={full} variant="portrait" />);
    expect(screen.getByRole('img')).toHaveAttribute('loading', 'eager');
    expect(screen.getByRole('img')).toHaveAttribute('fetchpriority', 'high');
  });

  it('starts loading again when the src changes after an error', () => {
    const { rerender } = render(<RemoteImage photo={full} variant="thumb" />);
    fireEvent.error(screen.getByRole('img'));
    rerender(
      <RemoteImage
        photo={{ ...full, thumbUrl: 'https://img.example.com/other.jpg' }}
        variant="thumb"
      />,
    );
    expect(screen.getByRole('img')).toHaveAttribute('src', 'https://img.example.com/other.jpg');
  });
});

describe('ErrorState', () => {
  it('shows the message and Retry calls the handler', async () => {
    const onRetry = vi.fn();
    render(<ErrorState onRetry={onRetry} />);
    expect(screen.getByRole('alert')).toHaveTextContent('Something went wrong');
    await userEvent.click(screen.getByRole('button', { name: 'Retry' }));
    expect(onRetry).toHaveBeenCalledTimes(1);
  });
});

describe('ChipList', () => {
  it('renders one list item per chip', () => {
    render(<ChipList items={['Go', 'SQL']} label="Skills" />);
    expect(screen.getByRole('list', { name: 'Skills' }).children).toHaveLength(2);
  });
});
