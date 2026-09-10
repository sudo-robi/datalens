import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { api } from '@/api/client'

interface SearchResult {
  total: number
  hits: Record<string, unknown>[]
  page: number
  per_page: number
}

interface SearchState {
  query: string
  results: SearchResult | null
  loading: boolean
  error: string | null
}

const initialState: SearchState = {
  query: '',
  results: null,
  loading: false,
  error: null,
}

export const searchDatasets = createAsyncThunk<SearchResult, { query: string; index?: string; page?: number }>(
  'search/datasets',
  async ({ query, index, page = 1 }) => {
    const response = await api.post<SearchResult>('/search', { query, index, page, per_page: 20 })
    return response.data
  }
)

const searchSlice = createSlice({
  name: 'search',
  initialState,
  reducers: {
    setQuery(state, action) {
      state.query = action.payload
    },
    clearResults(state) {
      state.results = null
      state.query = ''
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(searchDatasets.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(searchDatasets.fulfilled, (state, action) => {
        state.loading = false
        state.results = action.payload
      })
      .addCase(searchDatasets.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Search failed'
      })
  },
})

export const { setQuery, clearResults } = searchSlice.actions
export default searchSlice.reducer
