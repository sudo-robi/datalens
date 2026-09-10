import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { api } from '@/api/client'

interface Dataset {
  id: number
  name: string
  description: string
  filename: string
  mimeType: string
  rowCount: number
  sizeBytes: number
  createdAt: string
  updatedAt: string
}

interface DatasetsState {
  items: Dataset[]
  current: Dataset | null
  loading: boolean
  error: string | null
}

const initialState: DatasetsState = {
  items: [],
  current: null,
  loading: false,
  error: null,
}

export const fetchDatasets = createAsyncThunk<Dataset[]>(
  'datasets/fetchAll',
  async () => {
    const response = await api.get<Dataset[]>('/api/datasets')
    return response.data
  }
)

export const fetchDataset = createAsyncThunk<Dataset, number>(
  'datasets/fetchOne',
  async (id) => {
    const response = await api.get<Dataset>(`/api/datasets/${id}`)
    return response.data
  }
)

export const uploadDataset = createAsyncThunk<Dataset, { name: string; description: string; file: File }>(
  'datasets/upload',
  async ({ name, description, file }) => {
    const formData = new FormData()
    formData.append('name', name)
    formData.append('description', description)
    formData.append('file', file)
    const response = await api.post<Dataset>('/api/datasets/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return response.data
  }
)

export const deleteDataset = createAsyncThunk<number, number>(
  'datasets/delete',
  async (id) => {
    await api.delete(`/api/datasets/${id}`)
    return id
  }
)

const datasetsSlice = createSlice({
  name: 'datasets',
  initialState,
  reducers: {
    clearCurrent(state) {
      state.current = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchDatasets.pending, (state) => {
        state.loading = true
      })
      .addCase(fetchDatasets.fulfilled, (state, action) => {
        state.loading = false
        state.items = action.payload
      })
      .addCase(fetchDatasets.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Failed to fetch datasets'
      })
      .addCase(fetchDataset.fulfilled, (state, action) => {
        state.current = action.payload
      })
      .addCase(uploadDataset.fulfilled, (state, action) => {
        state.items.unshift(action.payload)
      })
      .addCase(deleteDataset.fulfilled, (state, action) => {
        state.items = state.items.filter((d) => d.id !== action.payload)
      })
  },
})

export const { clearCurrent } = datasetsSlice.actions
export default datasetsSlice.reducer
