import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Provider } from 'react-redux'
import { BrowserRouter } from 'react-router-dom'
import { configureStore } from '@reduxjs/toolkit'
import Upload from './Upload'
import datasetsReducer from '../store/slices/datasetsSlice'

vi.mock('@/api/client', () => ({
  api: { post: vi.fn(), get: vi.fn() },
}))

function renderWithProviders(ui: React.ReactElement, preloadedState = {}) {
  const store = configureStore({
    reducer: { datasets: datasetsReducer },
    preloadedState,
  })
  return render(
    <Provider store={store}>
      <BrowserRouter>{ui}</BrowserRouter>
    </Provider>
  )
}

describe('Upload page', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('renders name, description, and file inputs', () => {
    const { container } = renderWithProviders(<Upload />)
    expect(container.querySelector('input[type="text"]')).toBeInTheDocument()
    expect(container.querySelector('textarea')).toBeInTheDocument()
    expect(container.querySelector('input[type="file"]')).toBeInTheDocument()
  })

  it('shows Upload Dataset button', () => {
    renderWithProviders(<Upload />)
    expect(screen.getByRole('button', { name: /upload dataset/i })).toBeInTheDocument()
  })

  it('shows supported formats text', () => {
    renderWithProviders(<Upload />)
    expect(screen.getByText(/supported formats: csv, json/i)).toBeInTheDocument()
  })

  it('submitting without file shows "Please select a file" error', () => {
    const { container } = renderWithProviders(<Upload />)
    const form = container.querySelector('form')!
    HTMLFormElement.prototype.reportValidity = vi.fn().mockReturnValue(true)
    fireEvent.submit(form)
    expect(screen.getByText('Please select a file')).toBeInTheDocument()
  })

  it('renders Upload Dataset heading', () => {
    renderWithProviders(<Upload />)
    expect(screen.getByRole('heading', { name: /upload dataset/i })).toBeInTheDocument()
  })

  it('allows typing in name and description fields', () => {
    const { container } = renderWithProviders(<Upload />)
    const nameInput = container.querySelector('input[type="text"]') as HTMLInputElement
    const descInput = container.querySelector('textarea') as HTMLTextAreaElement

    fireEvent.change(nameInput, { target: { value: 'My Dataset' } })
    fireEvent.change(descInput, { target: { value: 'A test dataset' } })

    expect(nameInput).toHaveValue('My Dataset')
    expect(descInput).toHaveValue('A test dataset')
  })

  it('renders Dataset Name label', () => {
    renderWithProviders(<Upload />)
    expect(screen.getByText('Dataset Name')).toBeInTheDocument()
  })

  it('renders Description label', () => {
    renderWithProviders(<Upload />)
    expect(screen.getByText('Description')).toBeInTheDocument()
  })

  it('renders File label', () => {
    renderWithProviders(<Upload />)
    expect(screen.getByText('File')).toBeInTheDocument()
  })
})
