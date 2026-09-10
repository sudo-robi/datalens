import { useState, FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAppDispatch } from '../hooks'
import { uploadDataset } from '../store/slices/datasetsSlice'

export default function Upload() {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')
  const dispatch = useAppDispatch()
  const navigate = useNavigate()

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!file) {
      setError('Please select a file')
      return
    }

    setUploading(true)
    setError('')

    try {
      await dispatch(uploadDataset({ name, description, file })).unwrap()
      navigate('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed')
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="max-w-2xl">
      <h1 className="text-3xl font-bold mb-8">Upload Dataset</h1>

      <form onSubmit={handleSubmit} className="space-y-6 bg-white p-6 rounded-lg shadow">
        {error && (
          <div className="p-3 bg-red-100 text-red-700 rounded text-sm">{error}</div>
        )}

        <div>
          <label className="block text-sm font-medium mb-1">Dataset Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
            placeholder="e.g., Sales Data Q4 2024"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Description</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
            rows={3}
            placeholder="Optional description of the dataset"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">File</label>
          <input
            type="file"
            accept=".csv,.json"
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            className="w-full px-3 py-2 border rounded-lg"
            required
          />
          <p className="text-xs text-gray-500 mt-1">Supported formats: CSV, JSON</p>
        </div>

        <button
          type="submit"
          disabled={uploading}
          className="w-full py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50"
        >
          {uploading ? 'Uploading...' : 'Upload Dataset'}
        </button>
      </form>
    </div>
  )
}
