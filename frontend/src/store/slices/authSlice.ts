import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { api } from '@/api/client'

interface AuthResponse {
  token: string
  username: string
  email: string
}

interface AuthState {
  token: string | null
  username: string | null
  email: string | null
  loading: boolean
  error: string | null
}

const initialState: AuthState = {
  token: localStorage.getItem('token'),
  username: localStorage.getItem('username'),
  email: localStorage.getItem('email'),
  loading: false,
  error: null,
}

export const login = createAsyncThunk<AuthResponse, { username: string; password: string }>(
  'auth/login',
  async ({ username, password }) => {
    const response = await api.post<AuthResponse>('/api/auth/login', { username, password })
    return response.data
  }
)

export const register = createAsyncThunk<AuthResponse, { username: string; email: string; password: string }>(
  'auth/register',
  async ({ username, email, password }) => {
    const response = await api.post<AuthResponse>('/api/auth/register', { username, email, password })
    return response.data
  }
)

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    logout(state) {
      state.token = null
      state.username = null
      state.email = null
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      localStorage.removeItem('email')
    },
    clearError(state) {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(login.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(login.fulfilled, (state, action) => {
        state.loading = false
        state.token = action.payload.token
        state.username = action.payload.username
        state.email = action.payload.email
        localStorage.setItem('token', action.payload.token)
        localStorage.setItem('username', action.payload.username)
        localStorage.setItem('email', action.payload.email)
      })
      .addCase(login.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Login failed'
      })
      .addCase(register.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(register.fulfilled, (state, action) => {
        state.loading = false
        state.token = action.payload.token
        state.username = action.payload.username
        state.email = action.payload.email
        localStorage.setItem('token', action.payload.token)
        localStorage.setItem('username', action.payload.username)
        localStorage.setItem('email', action.payload.email)
      })
      .addCase(register.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Registration failed'
      })
  },
})

export const { logout, clearError } = authSlice.actions
export default authSlice.reducer
