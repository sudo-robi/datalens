import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { Provider } from 'react-redux'
import { BrowserRouter } from 'react-router-dom'
import { configureStore } from '@reduxjs/toolkit'
import DatasetDetail from './DatasetDetail'
import datasetsReducer from '../store/slices/datasetsSlice'

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useParams: () => ({ id: '42' }) }
})

vi.mock('@/api/client', () => ({
  api: {
    get: vi.fn().mockResolvedValue({
      data: {
        id: 42,
        name: 'Sales Data',
        description: 'Q4 2024 sales',
        filename: 'sales.csv',
        mimeType: 'text/csv',
        rowCount: 1500,
        sizeBytes: 1048576,
        createdAt: '2024-01-15T00:00:00Z',
        updatedAt: '2024-01-20T00:00:00Z',
      },
    }),
    post: vi.fn(),
    delete: vi.fn().mockResolvedValue({}),
  },
}))

const mockDataset = {
  id: 42,
  name: 'Sales Data',
  description: 'Q4 2024 sales',
  filename: 'sales.csv',
  mimeType: 'text/csv',
  rowCount: 1500,
  sizeBytes: 1048576,
  createdAt: '2024-01-15T00:00:00Z',
  updatedAt: '2024-01-20T00:00:00Z',
}

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

describe('DatasetDetail page', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('shows "Loading..." when current is null', async () => {
    const { api } = await import('@/api/client') as { api: { get: ReturnType<typeof vi.fn> } }
    api.get.mockImplementation(() => new Promise(() => {}))

    renderWithProviders(<DatasetDetail />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('shows dataset name and description after loading', async () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: { items: [], current: mockDataset, loading: false, error: null },
    })
    expect(screen.getByText('Sales Data')).toBeInTheDocument()
    expect(screen.getByText('Q4 2024 sales')).toBeInTheDocument()
  })

  it('shows row count', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: { items: [], current: mockDataset, loading: false, error: null },
    })
    expect(screen.getByText('1,500')).toBeInTheDocument()
  })

  it('shows size formatted', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: { items: [], current: mockDataset, loading: false, error: null },
    })
    expect(screen.getByText('1 MB')).toBeInTheDocument()
  })

  it('shows format', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: { items: [], current: mockDataset, loading: false, error: null },
    })
    expect(screen.getByText('text/csv')).toBeInTheDocument()
  })

  it('shows created date', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: { items: [], current: mockDataset, loading: false, error: null },
    })
    expect(screen.getByText(new Date('2024-01-15T00:00:00Z').toLocaleDateString())).toBeInTheDocument()
  })

  it('shows filename', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: { items: [], current: mockDataset, loading: false, error: null },
    })
    expect(screen.getByText('sales.csv')).toBeInTheDocument()
  })

  it('shows last updated date', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: { items: [], current: mockDataset, loading: false, error: null },
    })
    expect(screen.getByText(new Date('2024-01-20T00:00:00Z').toLocaleString())).toBeInTheDocument()
  })

  it('shows delete button', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: { items: [], current: mockDataset, loading: false, error: null },
    })
    expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument()
  })

  it('shows "Never" when updatedAt is null', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: {
        items: [],
        current: { ...mockDataset, updatedAt: null as unknown as string },
        loading: false,
        error: null,
      },
    })
    expect(screen.getByText('Never')).toBeInTheDocument()
  })

  it('shows "0 B" when sizeBytes is 0', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: {
        items: [],
        current: { ...mockDataset, sizeBytes: 0 },
        loading: false,
        error: null,
      },
    })
    expect(screen.getByText('0 B')).toBeInTheDocument()
  })

  it('shows "Unknown" when mimeType is null', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: {
        items: [],
        current: { ...mockDataset, mimeType: null as unknown as string },
        loading: false,
        error: null,
      },
    })
    expect(screen.getByText('Unknown')).toBeInTheDocument()
  })

  it('shows "0" when rowCount is null', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: {
        items: [],
        current: { ...mockDataset, rowCount: null as unknown as number },
        loading: false,
        error: null,
      },
    })
    expect(screen.getByText('0')).toBeInTheDocument()
  })

  it('hides description when description is empty', () => {
    renderWithProviders(<DatasetDetail />, {
      datasets: {
        items: [],
        current: { ...mockDataset, description: '' },
        loading: false,
        error: null,
      },
    })
    expect(screen.queryByText('Q4 2024 sales')).not.toBeInTheDocument()
  })

  it('dispatches fetchDataset on mount', async () => {
    const { api } = await import('@/api/client') as { api: { get: ReturnType<typeof vi.fn> } }
    api.get.mockResolvedValueOnce({ data: mockDataset })

    renderWithProviders(<DatasetDetail />)
    await waitFor(() => {
      expect(api.get).toHaveBeenCalledWith('/api/datasets/42')
    })
  })
})
