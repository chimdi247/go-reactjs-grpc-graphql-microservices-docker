import * as React from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { PackageOpen } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/context/AuthContext'
import { api } from '@/lib/api'

export default function Orders() {
  const { account, loading: authLoading } = useAuth()
  const navigate = useNavigate()
  const [orders, setOrders] = React.useState([])
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    if (authLoading) return
    if (!account) {
      navigate('/login')
      return
    }
    api
      .myOrders(account.id)
      .then((o) => setOrders([...o].sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt))))
      .catch((err) => toast.error(err.message || 'Failed to load orders'))
      .finally(() => setLoading(false))
  }, [account, authLoading, navigate])

  if (authLoading || loading) {
    return <div className="container py-8 text-muted-foreground">Loading orders...</div>
  }

  return (
    <div className="container max-w-3xl py-8">
      <h1 className="mb-6 text-2xl font-bold tracking-tight">My orders</h1>

      {orders.length === 0 ? (
        <div className="flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed py-16 text-center text-muted-foreground">
          <PackageOpen className="h-8 w-8" />
          <p>No orders yet.</p>
          <Button variant="outline" onClick={() => navigate('/')}>Browse catalog</Button>
        </div>
      ) : (
        <div className="space-y-4">
          {orders.map((order) => (
            <Card key={order.id}>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle className="text-base">Order #{order.id.slice(0, 8)}</CardTitle>
                  <CardDescription>{new Date(order.createdAt).toLocaleString()}</CardDescription>
                </div>
                <Badge variant="secondary">${order.totalPrice.toFixed(2)}</Badge>
              </CardHeader>
              <CardContent>
                <ul className="space-y-1 text-sm text-muted-foreground">
                  {order.products.map((p) => (
                    <li key={p.id} className="flex justify-between">
                      <span>{p.name} × {p.quantity}</span>
                      <span>${(p.price * p.quantity).toFixed(2)}</span>
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
