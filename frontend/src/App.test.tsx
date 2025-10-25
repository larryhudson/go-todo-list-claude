import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from './App'

describe('App', () => {
  it('renders the todo list heading', () => {
    render(<App />)
    expect(screen.getByText('Todo List')).toBeInTheDocument()
  })

  it('renders the add todo form', () => {
    render(<App />)
    expect(screen.getByPlaceholderText('Title')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Description (optional)')).toBeInTheDocument()
    expect(screen.getByText('Add Todo')).toBeInTheDocument()
  })
})
