import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Provider } from 'react-redux'
import { BrowserRouter } from 'react-router-dom'
import { configureStore } from '@reduxjs/toolkit'
import Register from './Register'
import authReducer from '../store/slices/authSlice'

vi.mock('@/api/client', () => ({
  api: { post: vi.fn() },
}))

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

describe('Register page', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('renders username, email, password, and confirm password inputs', () => {
    const { container } = renderWithProviders(<Register />)
    expect(container.querySelector('input[type="text"]')).toBeInTheDocument()
    expect(container.querySelector('input[type="email"]')).toBeInTheDocument()
    const passwordInputs = container.querySelectorAll('input[type="password"]')
    expect(passwordInputs).toHaveLength(2)
  })

  it('shows Sign Up button', () => {
    renderWithProviders(<Register />)
    expect(screen.getByRole('button', { name: /sign up/i })).toBeInTheDocument()
  })

  it('shows "Already have an account?" link to /login', () => {
    renderWithProviders(<Register />)
    expect(screen.getByText(/already have an account/i)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /sign in/i })).toHaveAttribute('href', '/login')
  })

  it('password mismatch shows "Passwords do not match" error', () => {
    const { container } = renderWithProviders(<Register />)
    fireEvent.change(container.querySelector('input[type="text"]')!, { target: { value: 'user1' } })
    fireEvent.change(container.querySelector('input[type="email"]')!, { target: { value: 'a@b.com' } })
    fireEvent.change(container.querySelectorAll('input[type="password"]')[0], { target: { value: 'password1' } })
    fireEvent.change(container.querySelectorAll('input[type="password"]')[1], { target: { value: 'password2' } })
    fireEvent.click(screen.getByRole('button', { name: /sign up/i }))

    expect(screen.getByText('Passwords do not match')).toBeInTheDocument()
  })

  it('dismiss button clears local error', () => {
    const { container } = renderWithProviders(<Register />)
    fireEvent.change(container.querySelector('input[type="text"]')!, { target: { value: 'user1' } })
    fireEvent.change(container.querySelector('input[type="email"]')!, { target: { value: 'a@b.com' } })
    fireEvent.change(container.querySelectorAll('input[type="password"]')[0], { target: { value: 'password1' } })
    fireEvent.change(container.querySelectorAll('input[type="password"]')[1], { target: { value: 'password2' } })
    fireEvent.click(screen.getByRole('button', { name: /sign up/i }))

    expect(screen.getByText('Passwords do not match')).toBeInTheDocument()
    fireEvent.click(screen.getByText('dismiss'))
    expect(screen.queryByText('Passwords do not match')).not.toBeInTheDocument()
  })

  it('shows error from auth state', () => {
    renderWithProviders(<Register />, {
      auth: { token: null, username: null, email: null, loading: false, error: 'Username taken' },
    })
    expect(screen.getByText('Username taken')).toBeInTheDocument()
  })

  it('shows Create Account heading', () => {
    renderWithProviders(<Register />)
    expect(screen.getByRole('heading', { name: /create account/i })).toBeInTheDocument()
  })

  it('shows "Creating account..." when loading', () => {
    renderWithProviders(<Register />, {
      auth: { token: null, username: null, email: null, loading: true, error: null },
    })
    expect(screen.getByRole('button', { name: /creating account/i })).toBeDisabled()
  })

  it('allows typing in all fields', () => {
    const { container } = renderWithProviders(<Register />)
    fireEvent.change(container.querySelector('input[type="text"]')!, { target: { value: 'user1' } })
    fireEvent.change(container.querySelector('input[type="email"]')!, { target: { value: 'a@b.com' } })
    fireEvent.change(container.querySelectorAll('input[type="password"]')[0], { target: { value: 'pass1234' } })
    fireEvent.change(container.querySelectorAll('input[type="password"]')[1], { target: { value: 'pass1234' } })

    expect(container.querySelector('input[type="text"]')).toHaveValue('user1')
    expect(container.querySelector('input[type="email"]')).toHaveValue('a@b.com')
  })

  it('matching passwords does not show local error', () => {
    const { container } = renderWithProviders(<Register />)
    const pwInputs = container.querySelectorAll('input[type="password"]')
    fireEvent.change(pwInputs[0], { target: { value: 'password123' } })
    fireEvent.change(pwInputs[1], { target: { value: 'password123' } })
    fireEvent.click(screen.getByRole('button', { name: /sign up/i }))

    expect(screen.queryByText('Passwords do not match')).not.toBeInTheDocument()
  })
})
