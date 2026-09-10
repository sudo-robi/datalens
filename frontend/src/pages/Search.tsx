import { useState, FormEvent } from 'react'
import { useAppDispatch, useAppSelector } from '../hooks'
import { searchDatasets, clearResults } from '../store/slices/searchSlice'

export default function Search() {
  const [query, setQuery] = useState('')
  const dispatch = useAppDispatch()
  const { results, loading, error } = useAppSelector((state) => state.search)

  const handleSearch = (e: FormEvent) => {
    e.preventDefault()
    if (query.trim()) {
      dispatch(searchDatasets({ query: query.trim() }))
    }
  }

  const handleClear = () => {
    setQuery('')
    dispatch(clearResults())
  }

  return (
    <div>
      <h1 className="text-3xl font-bold mb-8">Search Datasets</h1>

      <form onSubmit={handleSearch} className="flex gap-4 mb-8">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="flex-1 px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
          placeholder="Search across all datasets..."
        />
        <button
          type="submit"
          disabled={loading}
          className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50"
        >
          {loading ? 'Searching...' : 'Search'}
        </button>
        {results && (
          <button
            type="button"
            onClick={handleClear}
            className="px-4 py-2 border rounded-lg hover:bg-gray-100"
          >
            Clear
          </button>
        )}
      </form>

      {error && (
        <div className="p-4 bg-red-100 text-red-700 rounded mb-6">{error}</div>
      )}

      {results && (
        <div>
          <p className="text-sm text-gray-500 mb-4">
            Found {results.total.toLocaleString()} results
          </p>

          {results.hits.length === 0 ? (
            <div className="text-center py-12 text-gray-500">No results found</div>
          ) : (
            <div className="space-y-4">
              {results.hits.map((hit, index) => (
                <div
                  key={index}
                  className="p-4 bg-white rounded-lg shadow border-l-4 border-primary-500"
                >
                  <pre className="text-sm text-gray-700 overflow-x-auto">
                    {JSON.stringify(hit, null, 2)}
                  </pre>
                </div>
              ))}
            </div>
          )}

          {results.total > results.per_page && (
            <div className="mt-6 flex justify-center gap-2">
              {Array.from(
                { length: Math.min(Math.ceil(results.total / results.per_page), 10) },
                (_, i) => (
                  <button
                    key={i}
                    onClick={() =>
                      dispatch(searchDatasets({ query, page: i + 1 }))
                    }
                    className={`px-3 py-1 rounded ${
                      results.page === i + 1
                        ? 'bg-primary-600 text-white'
                        : 'bg-gray-200 hover:bg-gray-300'
                    }`}
                  >
                    {i + 1}
                  </button>
                )
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
