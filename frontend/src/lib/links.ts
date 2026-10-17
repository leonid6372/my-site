import type { Link } from '../api/types';

/** What a printed CV shows after the link text: the address for email, else the URL. */
export function printTarget(link: Link): string {
  return link.kind === 'email' ? link.url.replace(/^mailto:/i, '') : link.url;
}
