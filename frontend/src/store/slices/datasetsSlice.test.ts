import { describe, it, expect, beforeEach, vi } from 'vitest'
import datasetsReducer, { clearCurrent, fetchDatasets, fetchDataset, uploadDataset, deleteDataset } from './datasetsSlice'

vi.mock('@/api/client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}))

const { api } = await import('@/api/client') as { api: { get: ReturnType<typeof vi.fn>; post: ReturnType<typeof vi.fn>; delete: ReturnType<typeof vi.fn> } }

describe('datasetsSlice', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('initial state', () => {
    it('has empty items, null current, no loading/error', () => {
      const state = datasetsReducer(undefined, { type: 'unknown' })
      expect(state.items).toEqual([])
      expect(state.current).toBeNull()
      expect(state.loading).toBe(false)
      expect(state.error).toBeNull()
    })
  })

  describe('clearCurrent', () => {
    it('sets current to null', () => {
      const preState = { items: [], current: { id: 1 } as any, loading: false, error: null }
      const state = datasetsReducer(preState, clearCurrent())
      expect(state.current).toBeNull()
    })
  })

  describe('fetchDatasets.pending', () => {
    it('sets loading true', () => {
      const preState = { items: [], current: null, loading: false, error: null }
      const state = datasetsReducer(preState, { type: fetchDatasets.pending.type })
      expect(state.loading).toBe(true)
    })
  })

  describe('fetchDatasets.fulfilled', () => {
    it('sets items from payload and clears loading', () => {
      const preState = { items: [], current: null, loading: true, error: null }
      const datasets = [
        { id: 1, name: 'ds1', description: '', filename: 'a.csv', mimeType: 'text/csv', rowCount: 100, sizeBytes: 1024, createdAt: '2024-01-01', updatedAt: '2024-01-01' },
        { id: 2, name: 'ds2', description: 'desc', filename: 'b.json', mimeType: 'application/json', rowCount: 200, sizeBytes: 2048, createdAt: '2024-01-02', updatedAt: '2024-01-02' },
      ]
      const state = datasetsReducer(preState, { type: fetchDatasets.fulfilled.type, payload: datasets })
      expect(state.loading).toBe(false)
      expect(state.items).toHaveLength(2)
      expect(state.items[0].name).toBe('ds1')
      expect(state.items[1].name).toBe('ds2')
    })
  })

  describe('fetchDatasets.rejected', () => {
    it('sets error message from payload', () => {
      const preState = { items: [], current: null, loading: true, error: null }
      const state = datasetsReducer(preState, {
        type: fetchDatasets.rejected.type,
        error: { message: 'Network error' },
      })
      expect(state.loading).toBe(false)
      expect(state.error).toBe('Network error')
    })

    it('uses default error message when none provided', () => {
      const preState = { items: [], current: null, loading: true, error: null }
      const state = datasetsReducer(preState, {
        type: fetchDatasets.rejected.type,
        error: {},
      })
      expect(state.error).toBe('Failed to fetch datasets')
    })
  })

  describe('fetchDataset.fulfilled', () => {
    it('sets current from payload', () => {
      const preState = { items: [], current: null, loading: false, error: null }
      const dataset = { id: 5, name: 'detail', description: '', filename: 'c.csv', mimeType: 'text/csv', rowCount: 50, sizeBytes: 512, createdAt: '2024-03-01', updatedAt: '2024-03-01' }
      const state = datasetsReducer(preState, { type: fetchDataset.fulfilled.type, payload: dataset })
      expect(state.current).toEqual(dataset)
    })
  })

  describe('uploadDataset.fulfilled', () => {
    it('prepends uploaded dataset to items', () => {
      const preState = {
        items: [{ id: 1, name: 'existing', description: '', filename: 'a.csv', mimeType: 'text/csv', rowCount: 10, sizeBytes: 100, createdAt: '2024-01-01', updatedAt: '2024-01-01' }],
        current: null,
        loading: false,
        error: null,
      }
      const uploaded = { id: 2, name: 'new', description: '', filename: 'b.csv', mimeType: 'text/csv', rowCount: 20, sizeBytes: 200, createdAt: '2024-01-02', updatedAt: '2024-01-02' }
      const state = datasetsReducer(preState, { type: uploadDataset.fulfilled.type, payload: uploaded })
      expect(state.items).toHaveLength(2)
      expect(state.items[0].name).toBe('new')
      expect(state.items[1].name).toBe('existing')
    })
  })

  describe('deleteDataset.fulfilled', () => {
    it('removes dataset by id from items', () => {
      const preState = {
        items: [
          { id: 1, name: 'keep', description: '', filename: 'a.csv', mimeType: 'text/csv', rowCount: 10, sizeBytes: 100, createdAt: '2024-01-01', updatedAt: '2024-01-01' },
          { id: 2, name: 'remove', description: '', filename: 'b.csv', mimeType: 'text/csv', rowCount: 20, sizeBytes: 200, createdAt: '2024-01-02', updatedAt: '2024-01-02' },
        ],
        current: null,
        loading: false,
        error: null,
      }
      const state = datasetsReducer(preState, { type: deleteDataset.fulfilled.type, payload: 2 })
      expect(state.items).toHaveLength(1)
      expect(state.items[0].id).toBe(1)
    })
  })

  describe('fetchDatasets thunk', () => {
    it('dispatches fulfilled with dataset list', async () => {
      const datasets = [{ id: 1, name: 'a' }]
      api.get.mockResolvedValueOnce({ data: datasets })

      const thunk = fetchDatasets()
      const dispatch = vi.fn()
      const getState = vi.fn()

      await thunk(dispatch, getState, undefined)

      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: fetchDatasets.pending.type }))
      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: fetchDatasets.fulfilled.type, payload: datasets }))
    })
  })

  describe('fetchDataset thunk', () => {
    it('dispatches fulfilled with single dataset', async () => {
      const dataset = { id: 3, name: 'detail' }
      api.get.mockResolvedValueOnce({ data: dataset })

      const thunk = fetchDataset(3)
      const dispatch = vi.fn()
      const getState = vi.fn()

      await thunk(dispatch, getState, undefined)

      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: fetchDataset.fulfilled.type, payload: dataset }))
    })
  })

  describe('deleteDataset thunk', () => {
    it('dispatches fulfilled with deleted id', async () => {
      api.delete.mockResolvedValueOnce({})

      const thunk = deleteDataset(5)
      const dispatch = vi.fn()
      const getState = vi.fn()

      await thunk(dispatch, getState, undefined)

      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: deleteDataset.fulfilled.type, payload: 5 }))
    })
  })
})
