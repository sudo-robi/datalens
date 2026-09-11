import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { Provider } from 'react-redux'
import { BrowserRouter } from 'react-router-dom'
import { configureStore } from '@reduxjs/toolkit'
import Dashboard from './Dashboard'
import datasetsReducer from '../store/slices/datasetsSlice'

vi.mock('@/api/client', () => ({
  api: { get: vi.fn(), post: vi.fn(), delete: vi.fn() },
}))

const mockDatasets = [
  {
    id: 1, name: 'Sales Data', description: 'Q4 sales figures',
    filename: 'sales.csv', mimeType: 'text/csv', rowCount: 1500,
    sizeBytes: 1048576, createdAt: '2024-01-15T00:00:00Z', updatedAt: '2024-01-15T00:00:00Z',
  },
  {
    id: 2, name: 'User Logs', description: '',
    filename: 'logs.json', mimeType: 'application/json', rowCount: 50000,
    sizeBytes: 5242880, createdAt: '2024-02-20T00:00:00Z', updatedAt: '2024-02-20T00:00:00Z',
  },
]

function renderWithProviders(ui: React.ReactElement, preloadedState = {}) {
  const store = configureStore({
    reducer: { datasets: datasetsReducer },
    preloadedState,
  })
  return render(
    <Provider store={store}>
      <BrowserRouter>{ui}</BrowserRouter>
    </Provider>
  )
}

describe('Dashboard page', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('shows "Loading datasets..." on initial mount', async () => {
    const { api } = await import('@/api/client') as { api: { get: ReturnType<typeof vi.fn> } }
    api.get.mockImplementation(() => new Promise(() => {}))

    renderWithProviders(<Dashboard />)
    expect(screen.getByText('Loading datasets...')).toBeInTheDocument()
  })

  it('shows "No datasets yet" when API returns empty list', async () => {
    const { api } = await import('@/api/client') as { api: { get: ReturnType<typeof vi.fn> } }
    api.get.mockResolvedValueOnce({ data: [] })

    renderWithProviders(<Dashboard />)
    await waitFor(() => {
      expect(screen.getByText('No datasets yet')).toBeInTheDocument()
    })
  })

  it('renders "Upload your first dataset" link when empty', async () => {
    const { api } = await import('@/api/client') as { api: { get: ReturnType<typeof vi.fn> } }
    api.get.mockResolvedValueOnce({ data: [] })

    renderWithProviders(<Dashboard />)
    await waitFor(() => {
      expect(screen.getByRole('link', { name: /upload your first dataset/i })).toHaveAttribute('href', '/upload')
    })
  })

  it('renders dataset cards with name, description, and row count', async () => {
    const { api } = await import('@/api/client') as { api: { get: ReturnType<typeof vi.fn> } }
    api.get.mockResolvedValueOnce({ data: mockDatasets })

    renderWithProviders(<Dashboard />)
    await waitFor(() => {
      expect(screen.getByText('Sales Data')).toBeInTheDocument()
      expect(screen.getByText('Q4 sales figures')).toBeInTheDocument()
      expect(screen.getByText('1,500 rows')).toBeInTheDocument()
      expect(screen.getByText('User Logs')).toBeInTheDocument()
      expect(screen.getByText('50,000 rows')).toBeInTheDocument()
    })
  })

  it('renders Dashboard heading', async () => {
    const { api } = await import('@/api/client') as { api: { get: ReturnType<typeof vi.fn> } }
    api.get.mockResolvedValueOnce({ data: [] })

    renderWithProviders(<Dashboard />)
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /dashboard/i })).toBeInTheDocument()
    })
  })

  it('renders Upload Dataset button', () => {
    renderWithProviders(<Dashboard />)
    expect(screen.getByRole('link', { name: /upload dataset/i })).toHaveAttribute('href', '/upload')
  })

  it('dataset cards link to detail page', async () => {
    const { api } = await import('@/api/client') as { api: { get: ReturnType<typeof vi.fn> } }
    api.get.mockResolvedValueOnce({ data: [{ id: 42, name: 'Test Dataset', description: 'desc', filename: 'test.csv', mimeType: 'text/csv', rowCount: 100, sizeBytes: 1024, createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' }] })

    renderWithProviders(<Dashboard />)
    await waitFor(() => {
      const card = screen.getByText('Test Dataset').closest('a')
      expect(card).toHaveAttribute('href', '/datasets/42')
    })
  })
})

describe('formatBytes utility', () => {
  function formatBytes(bytes: number) {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
  }

  it('returns "0 B" for 0 bytes', () => {
    expect(formatBytes(0)).toBe('0 B')
  })

  it('formats bytes correctly', () => {
    expect(formatBytes(500)).toBe('500 B')
  })

  it('formats kilobytes correctly', () => {
    expect(formatBytes(1024)).toBe('1 KB')
    expect(formatBytes(1536)).toBe('1.5 KB')
    expect(formatBytes(10240)).toBe('10 KB')
  })

  it('formats megabytes correctly', () => {
    expect(formatBytes(1048576)).toBe('1 MB')
    expect(formatBytes(1572864)).toBe('1.5 MB')
  })

  it('formats gigabytes correctly', () => {
    expect(formatBytes(1073741824)).toBe('1 GB')
    expect(formatBytes(2147483648)).toBe('2 GB')
  })
})
