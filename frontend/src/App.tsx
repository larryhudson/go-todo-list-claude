import { useState, useEffect } from 'react'
import { TodosApi } from './generated/api/src/apis'
import type { ModelsTodo } from './generated/api/src/models'
import './App.css'

const api = new TodosApi()

function App() {
  // State for managing todos
  const [todos, setTodos] = useState<ModelsTodo[]>([])
  const [newTodoTitle, setNewTodoTitle] = useState('')
  const [newTodoDescription, setNewTodoDescription] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // State for filtering
  const [searchTerm, setSearchTerm] = useState('')
  const [filterCompleted, setFilterCompleted] = useState<'all' | 'completed' | 'incomplete'>('all')
  const [sortBy, setSortBy] = useState<'createdAt' | 'updatedAt' | 'title'>('createdAt')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')

  const fetchTodos = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await api.apiTodosGet({
        search: searchTerm || undefined,
        completed: filterCompleted === 'all' ? undefined : filterCompleted === 'completed',
        sortBy,
        sortOrder,
      })
      setTodos(response || [])
    } catch (err) {
      setError('Failed to fetch todos')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTodos()
  }, [searchTerm, filterCompleted, sortBy, sortOrder])

  const handleAddTodo = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTodoTitle.trim()) return

    try {
      setLoading(true)
      setError(null)
      await api.apiTodosPost({
        todo: {
          title: newTodoTitle,
          description: newTodoDescription,
        },
      })
      setNewTodoTitle('')
      setNewTodoDescription('')
      await fetchTodos()
    } catch (err) {
      setError('Failed to create todo')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleToggleComplete = async (todo: ModelsTodo) => {
    try {
      setLoading(true)
      setError(null)
      await api.apiTodosIdPatch({
        id: todo.id!,
        todo: {
          completed: !todo.completed,
        },
      })
      await fetchTodos()
    } catch (err) {
      setError('Failed to update todo')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteTodo = async (id: number) => {
    try {
      setLoading(true)
      setError(null)
      await api.apiTodosIdDelete({ id })
      await fetchTodos()
    } catch (err) {
      setError('Failed to delete todo')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="app">
      <h1>Todo List</h1>

      {error && <div className="error">{error}</div>}

      <form onSubmit={handleAddTodo} className="add-todo-form">
        <input
          type="text"
          placeholder="Title"
          value={newTodoTitle}
          onChange={(e) => setNewTodoTitle(e.target.value)}
          disabled={loading}
        />
        <input
          type="text"
          placeholder="Description (optional)"
          value={newTodoDescription}
          onChange={(e) => setNewTodoDescription(e.target.value)}
          disabled={loading}
        />
        <button type="submit" disabled={loading || !newTodoTitle.trim()}>
          Add Todo
        </button>
      </form>

      <div className="filters">
        <input
          type="text"
          placeholder="Search todos..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="search-input"
          disabled={loading}
        />

        <select
          value={filterCompleted}
          onChange={(e) => setFilterCompleted(e.target.value as 'all' | 'completed' | 'incomplete')}
          className="filter-select"
          disabled={loading}
        >
          <option value="all">All Todos</option>
          <option value="completed">Completed</option>
          <option value="incomplete">Incomplete</option>
        </select>

        <select
          value={sortBy}
          onChange={(e) => setSortBy(e.target.value as 'createdAt' | 'updatedAt' | 'title')}
          className="filter-select"
          disabled={loading}
        >
          <option value="createdAt">Sort by Created</option>
          <option value="updatedAt">Sort by Updated</option>
          <option value="title">Sort by Title</option>
        </select>

        <button
          onClick={() => setSortOrder(sortOrder === 'desc' ? 'asc' : 'desc')}
          className="sort-order-btn"
          disabled={loading}
          title={sortOrder === 'desc' ? 'Descending' : 'Ascending'}
        >
          {sortOrder === 'desc' ? '↓' : '↑'}
        </button>
      </div>

      <div className="todos-list">
        {loading && todos.length === 0 ? (
          <p>Loading...</p>
        ) : todos.length === 0 ? (
          <p>No todos yet. Add one above!</p>
        ) : (
          todos.map((todo) => (
            <div key={todo.id} className={`todo-item ${todo.completed ? 'completed' : ''}`}>
              <div className="todo-content">
                <input
                  type="checkbox"
                  checked={todo.completed || false}
                  onChange={() => handleToggleComplete(todo)}
                  disabled={loading}
                />
                <div className="todo-text">
                  <h3>{todo.title}</h3>
                  {todo.description && <p>{todo.description}</p>}
                  <small>Created: {new Date(todo.createdAt!).toLocaleString()}</small>
                </div>
              </div>
              <button
                onClick={() => handleDeleteTodo(todo.id!)}
                disabled={loading}
                className="delete-btn"
              >
                Delete
              </button>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

export default App
