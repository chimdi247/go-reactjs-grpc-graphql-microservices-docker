import * as React from 'react'
import { api } from '@/lib/api'

const AuthContext = React.createContext(null)

export function AuthProvider({ children }) {
  const [account, setAccount] = React.useState(null)
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    const token = localStorage.getItem('shopsphere_token')
    if (!token) {
      setLoading(false)
      return
    }
    api
      .me()
      .then((a) => setAccount(a))
      .catch(() => {
        localStorage.removeItem('shopsphere_token')
      })
      .finally(() => setLoading(false))
  }, [])

  const login = async (email, password) => {
    const { token, account } = await api.login(email, password)
    localStorage.setItem('shopsphere_token', token)
    setAccount(account)
    return account
  }

  const signup = async (name, email, password) => {
    await api.signup(name, email, password)
    return login(email, password)
  }

  const logout = () => {
    localStorage.removeItem('shopsphere_token')
    setAccount(null)
  }

  return (
    <AuthContext.Provider value={{ account, loading, login, signup, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = React.useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
