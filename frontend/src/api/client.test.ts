import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

describe('api client', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('api has correct baseURL', async () => {
    const { api } = await import('./client')
    expect(api.defaults.baseURL).toBeDefined()
  })

  it('ingestionApi has correct baseURL', async () => {
    const { ingestionApi } = await import('./client')
    expect(ingestionApi.defaults.baseURL).toBeDefined()
  })

  it('api request interceptor adds Authorization header when token exists', async () => {
    const { api } = await import('./client')
    localStorage.setItem('token', 'test-token-123')

    const config = { headers: {} as Record<string, string> }
    const interceptors = (api.interceptors.request as any).handlers
    const fulfilled = interceptors[0].fulfilled
    const result = fulfilled(config)

    expect(result.headers.Authorization).toBe('Bearer test-token-123')
  })

  it('api request interceptor does not add Authorization when no token', async () => {
    const { api } = await import('./client')
    localStorage.removeItem('token')

    const config = { headers: {} as Record<string, string> }
    const interceptors = (api.interceptors.request as any).handlers
    const fulfilled = interceptors[0].fulfilled
    const result = fulfilled(config)

    expect(result.headers.Authorization).toBeUndefined()
  })

  it('api response interceptor removes token on 401', async () => {
    const { api } = await import('./client')
    localStorage.setItem('token', 'expired-token')

    const error = { response: { status: 401 }, code: undefined }
    const interceptors = (api.interceptors.response as any).handlers
    const rejected = interceptors[0].rejected

    const originalHref = window.location.href
    Object.defineProperty(window, 'location', {
      value: { ...window.location, href: originalHref, set href(v: string) {} },
      writable: true,
    })

    try {
      await rejected(error)
    } catch {
      // expected to reject
    }

    expect(localStorage.getItem('token')).toBeNull()
  })

  it('api response interceptor rejects with network error message on ERR_NETWORK', async () => {
    const { api } = await import('./client')
    const error = { response: undefined, code: 'ERR_NETWORK' }
    const interceptors = (api.interceptors.response as any).handlers
    const rejected = interceptors[0].rejected

    await expect(rejected(error)).rejects.toThrow('API unavailable. The backend server is not running.')
  })

  it('api response interceptor rejects with network error message on 404', async () => {
    const { api } = await import('./client')
    const error = { response: { status: 404 }, code: undefined }
    const interceptors = (api.interceptors.response as any).handlers
    const rejected = interceptors[0].rejected

    await expect(rejected(error)).rejects.toThrow('API unavailable. The backend server is not running.')
  })

  it('api response interceptor passes through non-401/non-network errors', async () => {
    const { api } = await import('./client')
    const error = { response: { status: 500 }, code: undefined }
    const interceptors = (api.interceptors.response as any).handlers
    const rejected = interceptors[0].rejected

    await expect(rejected(error)).rejects.toEqual(error)
  })

  it('api response interceptor passes successful responses through', async () => {
    const { api } = await import('./client')
    const interceptors = (api.interceptors.response as any).handlers
    const fulfilled = interceptors[0].fulfilled

    const response = { data: { ok: true }, status: 200 }
    const result = fulfilled(response)
    expect(result).toEqual(response)
  })
})
