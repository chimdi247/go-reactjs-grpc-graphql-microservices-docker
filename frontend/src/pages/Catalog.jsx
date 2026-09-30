import * as React from 'react'
import { toast } from 'sonner'
import { Search, PlusCircle, PackageSearch } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { api } from '@/lib/api'
import { useCart } from '@/context/CartContext'

export default function Catalog() {
  const [products, setProducts] = React.useState([])
  const [query, setQuery] = React.useState('')
  const [loading, setLoading] = React.useState(true)
  const { addItem } = useCart()

  const load = React.useCallback((q) => {
    setLoading(true)
    api
      .products(q)
      .then(setProducts)
      .catch((err) => toast.error(err.message || 'Failed to load products'))
      .finally(() => setLoading(false))
  }, [])

  React.useEffect(() => {
    load('')
  }, [load])

  const onSearch = (e) => {
    e.preventDefault()
    load(query)
  }

  return (
    <div className="container py-8">
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Catalog</h1>
          <p className="text-muted-foreground">Browse products across the catalog service.</p>
        </div>
        <form onSubmit={onSearch} className="flex w-full gap-2 sm:w-80">
          <Input
            placeholder="Search products..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <Button type="submit" variant="secondary" size="icon">
            <Search className="h-4 w-4" />
          </Button>
        </form>
      </div>

      {loading ? (
        <p className="text-muted-foreground">Loading products...</p>
      ) : products.length === 0 ? (
        <div className="flex flex-col items-center justify-center gap-2 rounded-lg border border-dashed py-16 text-muted-foreground">
          <PackageSearch className="h-8 w-8" />
          <p>No products found.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {products.map((p) => (
            <Card key={p.id} className="flex flex-col">
              <CardHeader>
                <CardTitle className="line-clamp-1">{p.name}</CardTitle>
                <CardDescription className="line-clamp-2">{p.description}</CardDescription>
              </CardHeader>
              <CardContent className="mt-auto">
                <span className="text-xl font-semibold">${p.price.toFixed(2)}</span>
              </CardContent>
              <CardFooter>
                <Button className="w-full" onClick={() => { addItem(p); toast.success(`${p.name} added to cart`) }}>
                  <PlusCircle className="mr-2 h-4 w-4" />
                  Add to cart
                </Button>
              </CardFooter>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
