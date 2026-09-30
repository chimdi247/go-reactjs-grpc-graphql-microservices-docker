import * as React from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ShoppingBag, ShoppingCart, LogOut, User, LayoutDashboard } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useAuth } from '@/context/AuthContext'
import { useCart } from '@/context/CartContext'

export default function Navbar() {
  const { account, logout } = useAuth()
  const { count } = useCart()
  const navigate = useNavigate()

  return (
    <header className="sticky top-0 z-40 border-b bg-background/95 backdrop-blur">
      <div className="container flex h-14 items-center justify-between">
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <ShoppingBag className="h-5 w-5" />
          ShopSphere
        </Link>

        <nav className="flex items-center gap-1">
          <Button variant="ghost" size="sm" asChild>
            <Link to="/">Catalog</Link>
          </Button>

          {account && (
            <Button variant="ghost" size="sm" asChild>
              <Link to="/orders">My Orders</Link>
            </Button>
          )}

          <Button variant="ghost" size="sm" asChild>
            <Link to="/admin">
              <LayoutDashboard className="mr-1 h-4 w-4" />
              Admin
            </Link>
          </Button>

          <Button variant="ghost" size="icon" asChild className="relative">
            <Link to="/cart">
              <ShoppingCart className="h-5 w-5" />
              {count > 0 && (
                <Badge className="absolute -right-1 -top-1 h-5 w-5 justify-center rounded-full p-0 text-[10px]">
                  {count}
                </Badge>
              )}
            </Link>
          </Button>

          {account ? (
            <div className="ml-2 flex items-center gap-2 border-l pl-3">
              <span className="hidden items-center gap-1 text-sm text-muted-foreground sm:flex">
                <User className="h-4 w-4" />
                {account.name}
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  logout()
                  navigate('/')
                }}
              >
                <LogOut className="mr-1 h-4 w-4" />
                Logout
              </Button>
            </div>
          ) : (
            <div className="ml-2 flex items-center gap-2 border-l pl-3">
              <Button variant="ghost" size="sm" asChild>
                <Link to="/login">Login</Link>
              </Button>
              <Button size="sm" asChild>
                <Link to="/signup">Sign up</Link>
              </Button>
            </div>
          )}
        </nav>
      </div>
    </header>
  )
}
