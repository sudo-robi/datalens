import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Provider } from 'react-redux'
import { BrowserRouter } from 'react-router-dom'
import { configureStore } from '@reduxjs/toolkit'
import Search from './Search'
import searchReducer from '../store/slices/searchSlice'

vi.mock('@/api/client', () => ({
  api: { post: vi.fn() },
}))

function renderWithProviders(ui: React.ReactElement, preloadedState = {}) {
  const store = configureStore({
    reducer: { search: searchReducer },
    preloadedState,
  })
  return render(
    <Provider store={store}>
      <BrowserRouter>{ui}</BrowserRouter>
    </Provider>
  )
}

describe('Search page', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('renders search input', () => {
    renderWithProviders(<Search />)
    expect(screen.getByPlaceholderText(/search across all datasets/i)).toBeInTheDocument()
  })

  it('shows Search button', () => {
    renderWithProviders(<Search />)
    expect(screen.getByRole('button', { name: /search/i })).toBeInTheDocument()
  })

  it('shows results count when results exist', () => {
    renderWithProviders(<Search />, {
      search: {
        query: 'test',
        results: { total: 42, hits: [{ id: 1 }], page: 1, per_page: 20 },
        loading: false,
        error: null,
      },
    })
    expect(screen.getByText('Found 42 results')).toBeInTheDocument()
  })

  it('shows "No results found" when empty results', () => {
    renderWithProviders(<Search />, {
      search: {
        query: 'test',
        results: { total: 0, hits: [], page: 1, per_page: 20 },
        loading: false,
        error: null,
      },
    })
    expect(screen.getByText('No results found')).toBeInTheDocument()
  })

  it('shows error message when error exists', () => {
    renderWithProviders(<Search />, {
      search: {
        query: 'test',
        results: null,
        loading: false,
        error: 'Something went wrong',
      },
    })
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
  })

  it('Clear button resets query', () => {
    const store = configureStore({
      reducer: { search: searchReducer },
      preloadedState: {
        search: {
          query: 'old query',
          results: { total: 1, hits: [{ a: 1 }], page: 1, per_page: 20 },
          loading: false,
          error: null,
        },
      },
    })

    render(
      <Provider store={store}>
        <BrowserRouter><Search /></BrowserRouter>
      </Provider>
    )

    expect(screen.getByText('Clear')).toBeInTheDocument()
    fireEvent.click(screen.getByText('Clear'))
    expect(store.getState().search.query).toBe('')
    expect(store.getState().search.results).toBeNull()
  })

  it('Search Datasets heading is rendered', () => {
    renderWithProviders(<Search />)
    expect(screen.getByRole('heading', { name: /search datasets/i })).toBeInTheDocument()
  })

  it('shows Searching... when loading', () => {
    renderWithProviders(<Search />, {
      search: { query: 'q', results: null, loading: true, error: null },
    })
    expect(screen.getByRole('button', { name: /searching/i })).toBeDisabled()
  })

  it('renders hit items when results have hits', () => {
    renderWithProviders(<Search />, {
      search: {
        query: 'test',
        results: { total: 1, hits: [{ name: 'Sales Data' }], page: 1, per_page: 20 },
        loading: false,
        error: null,
      },
    })
    expect(screen.getByText(/"name": "Sales Data"/)).toBeInTheDocument()
  })

  it('shows Clear button only when results exist', () => {
    const store = configureStore({
      reducer: { search: searchReducer },
      preloadedState: {
        search: { query: '', results: null, loading: false, error: null },
      },
    })

    const { unmount } = render(
      <Provider store={store}>
        <BrowserRouter><Search /></BrowserRouter>
      </Provider>
    )
    expect(screen.queryByText('Clear')).not.toBeInTheDocument()
    unmount()
  })

  it('allows typing in search input', () => {
    renderWithProviders(<Search />)
    const input = screen.getByPlaceholderText(/search across all datasets/i)
    fireEvent.change(input, { target: { value: 'my query' } })
    expect(input).toHaveValue('my query')
  })
})
