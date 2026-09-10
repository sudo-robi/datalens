import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../hooks'
import { fetchDatasets } from '../store/slices/datasetsSlice'

export default function Dashboard() {
  const dispatch = useAppDispatch()
  const { items, loading } = useAppSelector((state) => state.datasets)

  useEffect(() => {
    dispatch(fetchDatasets())
  }, [dispatch])

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <Link
          to="/upload"
          className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition"
        >
          Upload Dataset
        </Link>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading datasets...</div>
      ) : items.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-gray-500 mb-4">No datasets yet</p>
          <Link
            to="/upload"
            className="text-primary-600 hover:underline"
          >
            Upload your first dataset
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {items.map((dataset) => (
            <Link
              key={dataset.id}
              to={`/datasets/${dataset.id}`}
              className="block p-6 bg-white rounded-lg shadow hover:shadow-md transition"
            >
              <h3 className="font-semibold text-lg mb-2">{dataset.name}</h3>
              {dataset.description && (
                <p className="text-gray-600 text-sm mb-3 line-clamp-2">{dataset.description}</p>
              )}
              <div className="flex justify-between text-sm text-gray-500">
                <span>{dataset.rowCount?.toLocaleString() ?? 0} rows</span>
                <span>{formatBytes(dataset.sizeBytes ?? 0)}</span>
              </div>
              <div className="mt-2 text-xs text-gray-400">
                {new Date(dataset.createdAt).toLocaleDateString()}
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
