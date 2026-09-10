import { Outlet, Link, useNavigate } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../hooks'
import { logout } from '../store/slices/authSlice'

export default function Layout() {
  const { username } = useAppSelector((state) => state.auth)
  const dispatch = useAppDispatch()
  const navigate = useNavigate()

  const handleLogout = () => {
    dispatch(logout())
    navigate('/login')
  }

  return (
    <div className="min-h-screen flex">
      <aside className="w-64 bg-gray-900 text-white p-4">
        <h1 className="text-2xl font-bold mb-8 text-primary-400">DataLens</h1>
        <nav className="space-y-2">
          <Link to="/" className="block px-4 py-2 rounded hover:bg-gray-800 transition">
            Dashboard
          </Link>
          <Link to="/upload" className="block px-4 py-2 rounded hover:bg-gray-800 transition">
            Upload
          </Link>
          <Link to="/search" className="block px-4 py-2 rounded hover:bg-gray-800 transition">
            Search
          </Link>
        </nav>
        <div className="mt-auto pt-8 border-t border-gray-700">
          <p className="text-sm text-gray-400 mb-2">{username}</p>
          <button
            onClick={handleLogout}
            className="w-full px-4 py-2 bg-red-600 rounded hover:bg-red-700 transition text-sm"
          >
            Logout
          </button>
        </div>
      </aside>
      <main className="flex-1 p-8 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
