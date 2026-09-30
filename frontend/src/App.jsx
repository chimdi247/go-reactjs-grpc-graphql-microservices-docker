import * as React from 'react'
import { Routes, Route } from 'react-router-dom'
import { Toaster } from 'sonner'
import Navbar from '@/components/Navbar'
import Catalog from '@/pages/Catalog'
import Cart from '@/pages/Cart'
import Orders from '@/pages/Orders'
import Login from '@/pages/Login'
import Signup from '@/pages/Signup'
import Admin from '@/pages/Admin'
import { AuthProvider } from '@/context/AuthContext'
import { CartProvider } from '@/context/CartContext'

export default function App() {
  return (
    <AuthProvider>
      <CartProvider>
        <Toaster richColors position="top-right" />
        <Navbar />
        <main>
          <Routes>
            <Route path="/" element={<Catalog />} />
            <Route path="/cart" element={<Cart />} />
            <Route path="/orders" element={<Orders />} />
            <Route path="/admin" element={<Admin />} />
            <Route path="/login" element={<Login />} />
            <Route path="/signup" element={<Signup />} />
          </Routes>
        </main>
      </CartProvider>
    </AuthProvider>
  )
}
