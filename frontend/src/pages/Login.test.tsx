import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Provider } from 'react-redux'
import { BrowserRouter } from 'react-router-dom'
import { configureStore } from '@reduxjs/toolkit'
import Login from './Login'
import authReducer from '../store/slices/authSlice'

function renderWithProviders(ui: React.ReactElement, preloadedState = {}) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState,
  })
  return render(
    <Provider store={store}>
      <BrowserRouter>{ui}</BrowserRouter>
    </Provider>
  )
}

describe('Login page', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('renders username and password inputs', () => {
    const { container } = renderWithProviders(<Login />)
    expect(container.querySelector('input[type="text"]')).toBeInTheDocument()
    expect(container.querySelector('input[type="password"]')).toBeInTheDocument()
  })

  it('renders Sign In button', () => {
    renderWithProviders(<Login />)
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('renders "Don\'t have an account?" link', () => {
    renderWithProviders(<Login />)
    expect(screen.getByText(/don't have an account/i)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /sign up/i })).toHaveAttribute('href', '/register')
  })

  it('renders DataLens heading', () => {
    renderWithProviders(<Login />)
    expect(screen.getByRole('heading', { name: /datalens/i })).toBeInTheDocument()
  })

  it('shows error message when auth state has error', () => {
    renderWithProviders(<Login />, {
      auth: { token: null, username: null, email: null, loading: false, error: 'Invalid credentials' },
    })
    expect(screen.getByText('Invalid credentials')).toBeInTheDocument()
  })

  it('dismiss button clears error', () => {
    const store = configureStore({
      reducer: { auth: authReducer },
      preloadedState: {
        auth: { token: null, username: null, email: null, loading: false, error: 'Some error' },
      },
    })

    render(
      <Provider store={store}>
        <BrowserRouter><Login /></BrowserRouter>
      </Provider>
    )

    expect(screen.getByText('Some error')).toBeInTheDocument()
    fireEvent.click(screen.getByText('dismiss'))
    expect(screen.queryByText('Some error')).not.toBeInTheDocument()
  })

  it('shows "Signing in..." text when loading', () => {
    renderWithProviders(<Login />, {
      auth: { token: null, username: null, email: null, loading: true, error: null },
    })
    expect(screen.getByRole('button', { name: /signing in/i })).toBeDisabled()
  })

  it('allows typing in username and password fields', () => {
    const { container } = renderWithProviders(<Login />)
    const usernameInput = container.querySelector('input[type="text"]') as HTMLInputElement
    const passwordInput = container.querySelector('input[type="password"]') as HTMLInputElement

    fireEvent.change(usernameInput, { target: { value: 'testuser' } })
    fireEvent.change(passwordInput, { target: { value: 'testpass' } })

    expect(usernameInput).toHaveValue('testuser')
    expect(passwordInput).toHaveValue('testpass')
  })
})
