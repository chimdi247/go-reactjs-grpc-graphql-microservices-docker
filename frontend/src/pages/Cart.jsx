import * as React from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { Minus, Plus, Trash2, ShoppingBag } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { useCart } from '@/context/CartContext'
import { useAuth } from '@/context/AuthContext'
import { api } from '@/lib/api'

export default function Cart() {
  const { items, removeItem, setQuantity, total, clear } = useCart()
  const { account } = useAuth()
  const navigate = useNavigate()
  const [placing, setPlacing] = React.useState(false)

  const checkout = async () => {
    if (!account) {
      toast.error('Please log in to place an order')
      navigate('/login')
      return
    }
    setPlacing(true)
    try {
      await api.createOrder(account.id, items)
      clear()
      toast.success('Order placed!')
      navigate('/orders')
    } catch (err) {
      toast.error(err.message || 'Failed to place order')
    } finally {
      setPlacing(false)
    }
  }

  if (items.length === 0) {
    return (
      <div className="container flex flex-col items-center justify-center gap-3 py-24 text-center">
        <ShoppingBag className="h-10 w-10 text-muted-foreground" />
        <h2 className="text-xl font-semibold">Your cart is empty</h2>
        <p className="text-muted-foreground">Add some products from the catalog first.</p>
        <Button onClick={() => navigate('/')}>Browse catalog</Button>
      </div>
    )
  }

  return (
    <div className="container max-w-2xl py-8">
      <h1 className="mb-6 text-2xl font-bold tracking-tight">Your cart</h1>
      <Card>
        <CardHeader>
          <CardTitle>Items</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {items.map((item, idx) => (
            <div key={item.id}>
              <div className="flex items-center justify-between gap-4">
                <div>
                  <p className="font-medium">{item.name}</p>
                  <p className="text-sm text-muted-foreground">${item.price.toFixed(2)} each</p>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setQuantity(item.id, item.quantity - 1)}
                  >
                    <Minus className="h-4 w-4" />
                  </Button>
                  <span className="w-6 text-center">{item.quantity}</span>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setQuantity(item.id, item.quantity + 1)}
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon" onClick={() => removeItem(item.id)}>
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </div>
              {idx < items.length - 1 && <Separator className="mt-4" />}
            </div>
          ))}
        </CardContent>
        <CardFooter className="flex flex-col gap-4 border-t pt-6">
          <div className="flex w-full items-center justify-between text-lg font-semibold">
            <span>Total</span>
            <span>${total.toFixed(2)}</span>
          </div>
          <Button className="w-full" size="lg" onClick={checkout} disabled={placing}>
            {placing ? 'Placing order...' : 'Checkout'}
          </Button>
        </CardFooter>
      </Card>
    </div>
  )
}
