import { useState, useEffect, useCallback } from 'react'
import { ItemsService } from './api/services/ItemsService'
import type { ItemResponse } from './api/models/ItemResponse'

type Filter = 'all' | 'active' | 'done'

function App() {
  const [items, setItems] = useState<ItemResponse[]>([])
  const [newName, setNewName] = useState('')
  const [filter, setFilter] = useState<Filter>('all')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchItems = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const done = filter === 'all' ? undefined : filter === 'done'
      const result = await ItemsService.getItems(100, 1, undefined, undefined, done)
      setItems(result ?? [])
    } catch {
      setError('Failed to load items')
    } finally {
      setLoading(false)
    }
  }, [filter])

  useEffect(() => {
    fetchItems()
  }, [fetchItems])

  const addItem = async (e: React.FormEvent) => {
    e.preventDefault()
    const name = newName.trim()
    if (!name) return
    try {
      await ItemsService.postItems({ name })
      setNewName('')
      fetchItems()
    } catch {
      setError('Failed to create item')
    }
  }

  const toggleDone = async (item: ItemResponse) => {
    try {
      await ItemsService.patchItemsItemid(item.id!, { done: !item.done })
      fetchItems()
    } catch {
      setError('Failed to update item')
    }
  }

  const activeCount = items.filter(i => !i.done).length

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 py-8 px-4 sm:py-16">
      <div className="mx-auto max-w-lg">
        {/* Header */}
        <div className="mb-8 text-center">
          <h1 className="text-4xl font-extrabold tracking-tight bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
            Todo
          </h1>
          <p className="mt-1 text-sm text-slate-400">Stay organized, get things done</p>
        </div>

        {/* Card */}
        <div className="rounded-2xl bg-white shadow-xl shadow-slate-200/50 ring-1 ring-slate-100 overflow-hidden">
          {/* Add form */}
          <form onSubmit={addItem} className="flex gap-2 p-4 border-b border-slate-100">
            <input
              type="text"
              placeholder="What needs to be done?"
              value={newName}
              onChange={e => setNewName(e.target.value)}
              autoFocus
              className="flex-1 rounded-lg border border-slate-200 px-4 py-2.5 text-sm text-slate-700 placeholder:text-slate-300 outline-none transition-all focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
            />
            <button
              type="submit"
              className="rounded-lg bg-indigo-600 px-5 py-2.5 text-sm font-medium text-white transition-all hover:bg-indigo-700 active:scale-95 disabled:opacity-50"
              disabled={!newName.trim()}
            >
              Add
            </button>
          </form>

          {/* Error */}
          {error && (
            <div className="mx-4 mt-3 flex items-center gap-2 rounded-lg bg-red-50 px-4 py-2.5 text-sm text-red-600">
              <svg className="h-4 w-4 shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
              {error}
              <button onClick={() => setError(null)} className="ml-auto text-red-400 hover:text-red-600">
                <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
          )}

          {/* Filters */}
          <div className="flex gap-1 px-4 pt-3">
            {(['all', 'active', 'done'] as Filter[]).map(f => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`rounded-full px-3.5 py-1 text-xs font-medium transition-all ${
                  filter === f
                    ? 'bg-indigo-100 text-indigo-700'
                    : 'text-slate-400 hover:bg-slate-50 hover:text-slate-600'
                }`}
              >
                {f.charAt(0).toUpperCase() + f.slice(1)}
              </button>
            ))}
          </div>

          {/* List */}
          <div className="p-4">
            {loading ? (
              <div className="flex items-center justify-center py-12">
                <div className="h-6 w-6 animate-spin rounded-full border-2 border-slate-200 border-t-indigo-600" />
              </div>
            ) : items.length === 0 ? (
              <div className="py-12 text-center">
                <div className="text-3xl mb-2">
                  {filter === 'done' ? '\uD83C\uDFAF' : filter === 'active' ? '\u2705' : '\uD83D\uDCDD'}
                </div>
                <p className="text-sm text-slate-400">
                  {filter === 'all' ? 'No tasks yet. Add one above!' : `No ${filter} tasks`}
                </p>
              </div>
            ) : (
              <ul className="space-y-1.5">
                {items.map(item => (
                  <li
                    key={item.id}
                    className={`group flex items-center gap-3 rounded-xl px-3 py-2.5 transition-all hover:bg-slate-50 ${
                      item.done ? 'opacity-60' : ''
                    }`}
                  >
                    <button
                      onClick={() => toggleDone(item)}
                      className={`flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 transition-all ${
                        item.done
                          ? 'border-indigo-500 bg-indigo-500 text-white'
                          : 'border-slate-300 hover:border-indigo-400'
                      }`}
                    >
                      {item.done && (
                        <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                        </svg>
                      )}
                    </button>
                    <span className={`flex-1 text-sm ${item.done ? 'text-slate-400 line-through' : 'text-slate-700'}`}>
                      {item.name}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between border-t border-slate-100 px-4 py-3">
            <span className="text-xs text-slate-400">
              {activeCount} item{activeCount !== 1 ? 's' : ''} left
            </span>
            {items.some(i => i.done) && (
              <span className="text-xs text-slate-400">
                {items.filter(i => i.done).length} completed
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
