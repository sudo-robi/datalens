import { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../hooks'
import { fetchDataset, deleteDataset, clearCurrent } from '../store/slices/datasetsSlice'

export default function DatasetDetail() {
  const { id } = useParams<{ id: string }>()
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const { current } = useAppSelector((state) => state.datasets)

  useEffect(() => {
    if (id) {
      dispatch(fetchDataset(Number(id)))
    }
    return () => {
      dispatch(clearCurrent())
    }
  }, [id, dispatch])

  const handleDelete = async () => {
    if (id && window.confirm('Are you sure you want to delete this dataset?')) {
      await dispatch(deleteDataset(Number(id)))
      navigate('/')
    }
  }

  const formatBytes = (bytes: number) => {
    if (!bytes) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
  }

  if (!current) {
    return <div className="text-center py-12 text-gray-500">Loading...</div>
  }

  return (
    <div className="max-w-4xl">
      <div className="flex justify-between items-start mb-8">
        <div>
          <h1 className="text-3xl font-bold">{current.name}</h1>
          {current.description && (
            <p className="text-gray-600 mt-2">{current.description}</p>
          )}
        </div>
        <button
          onClick={handleDelete}
          className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
        >
          Delete
        </button>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <div className="p-4 bg-white rounded-lg shadow">
          <p className="text-sm text-gray-500">Rows</p>
          <p className="text-2xl font-bold">{current.rowCount?.toLocaleString() ?? 0}</p>
        </div>
        <div className="p-4 bg-white rounded-lg shadow">
          <p className="text-sm text-gray-500">Size</p>
          <p className="text-2xl font-bold">{formatBytes(current.sizeBytes ?? 0)}</p>
        </div>
        <div className="p-4 bg-white rounded-lg shadow">
          <p className="text-sm text-gray-500">Format</p>
          <p className="text-2xl font-bold">{current.mimeType ?? 'Unknown'}</p>
        </div>
        <div className="p-4 bg-white rounded-lg shadow">
          <p className="text-sm text-gray-500">Created</p>
          <p className="text-2xl font-bold">
            {new Date(current.createdAt).toLocaleDateString()}
          </p>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4">Dataset Info</h2>
        <dl className="grid grid-cols-2 gap-4">
          <div>
            <dt className="text-sm text-gray-500">Filename</dt>
            <dd className="font-medium">{current.filename}</dd>
          </div>
          <div>
            <dt className="text-sm text-gray-500">Last Updated</dt>
            <dd className="font-medium">
              {current.updatedAt
                ? new Date(current.updatedAt).toLocaleString()
                : 'Never'}
            </dd>
          </div>
        </dl>
      </div>
    </div>
  )
}
