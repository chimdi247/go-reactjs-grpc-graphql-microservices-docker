import * as React from 'react'
import { toast } from 'sonner'
import { Users, Activity } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { api } from '@/lib/api'

export default function Admin() {
  const [totalAccounts, setTotalAccounts] = React.useState(null)
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    api
      .totalAccounts()
      .then(setTotalAccounts)
      .catch((err) => toast.error(err.message || 'Failed to load stats'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="container max-w-3xl py-8">
      <h1 className="mb-2 text-2xl font-bold tracking-tight">Platform stats</h1>
      <p className="mb-6 text-muted-foreground">
        The same figures also feed the Grafana dashboard's business-KPI panel.
      </p>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{loading ? '...' : totalAccounts}</div>
            <CardDescription>Registered accounts (account service)</CardDescription>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Observability</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <a
              href="http://localhost:3001"
              target="_blank"
              rel="noreferrer"
              className="text-sm font-medium text-primary underline underline-offset-4"
            >
              Open Grafana dashboard →
            </a>
            <CardDescription className="mt-1">
              Uptime, latency, error rates, CPU/memory, alerts
            </CardDescription>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
