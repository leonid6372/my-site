import type { ApiError as ApiErrorBody, Lang } from './types';

const BASE = '/api/v1';

/** A failed API request: an HTTP error (with the contract's error code) or a network error. */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

function isErrorBody(v: unknown): v is ApiErrorBody {
  if (typeof v !== 'object' || v === null || !('error' in v)) return false;
  const e = v.error;
  return (
    typeof e === 'object' &&
    e !== null &&
    'code' in e &&
    typeof e.code === 'string' &&
    'message' in e &&
    typeof e.message === 'string'
  );
}

async function toApiError(res: Response): Promise<ApiError> {
  try {
    const body: unknown = await res.json();
    if (isErrorBody(body)) return new ApiError(res.status, body.error.code, body.error.message);
  } catch {
    // not JSON (e.g. a proxy error page): fall through
  }
  return new ApiError(res.status, 'http_error', `HTTP ${res.status}`);
}

/**
 * GET an API resource in `lang`. The language travels only in the Accept-Language
 * header (contract §1.1), never in the URL.
 */
export async function apiGet<T>(path: string, lang: Lang, signal?: AbortSignal): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${BASE}${path}`, { headers: { 'Accept-Language': lang }, signal });
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') throw err;
    throw new ApiError(0, 'network_error', 'network error');
  }
  if (!res.ok) throw await toApiError(res);
  return (await res.json()) as T;
}
