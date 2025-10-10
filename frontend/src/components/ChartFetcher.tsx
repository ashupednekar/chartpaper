import { useState } from 'react'
import { ChartInfo } from '../types'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Badge } from './ui/badge'
import { Textarea } from './ui/textarea'
import { Switch } from './ui/switch'
import { Label } from './ui/label'
import { Download, Shield, AlertCircle, CheckCircle, Sparkles, Zap, Database } from 'lucide-react'

interface ChartFetcherProps {
  onChartFetched: (chart: ChartInfo) => void
}

interface FetchResponse {
  chart: ChartInfo
  apps: any[]
}

export function ChartFetcher({ onChartFetched }: ChartFetcherProps) {
  const [chartUrl, setChartUrl] = useState('oci://registry-1.docker.io/bitnamicharts/wordpress')
  const [valuesPath, setValuesPath] = useState('values')
  const [setValues, setSetValues] = useState('')
  const [useHostNetwork, setUseHostNetwork] = useState(false)
  const [loading, setLoading] = useState(false)
  const [authStatus, setAuthStatus] = useState<'unknown' | 'authenticated' | 'failed'>('unknown')
  const [error, setError] = useState<string | null>(null)
  const [lastFetched, setLastFetched] = useState<FetchResponse | null>(null)

  const authenticate = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await fetch('/api/authenticate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      })
      
      if (response.ok) {
        setAuthStatus('authenticated')
      } else {
        const errorData = await response.json()
        setError(errorData.error || 'Authentication failed')
        setAuthStatus('failed')
      }
    } catch (err) {
      setError('Network error during authentication')
      setAuthStatus('failed')
    } finally {
      setLoading(false)
    }
  }

  const fetchChart = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const setValuesArray = setValues
        .split('\n')
        .map(line => line.trim())
        .filter(line => line.length > 0)

      const response = await fetch('/api/fetch-chart', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          chartUrl,
          valuesPath,
          setValues: setValuesArray,
          useHostNetwork,
        }),
      })
      
      if (response.ok) {
        const data: FetchResponse = await response.json()
        setLastFetched(data)
        onChartFetched(data.chart)
      } else {
        const errorData = await response.json()
        setError(errorData.error || 'Failed to fetch chart')
      }
    } catch (err) {
      setError('Network error during chart fetch')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="text-center py-8">
        <Sparkles className="h-12 w-12 mx-auto mb-4 text-primary" />
        <h2 className="text-2xl font-bold mb-2">Fetch Helm Charts from OCI Registries</h2>
        <p className="text-muted-foreground max-w-2xl mx-auto">
          Connect to any OCI-compatible registry and fetch Helm charts with their dependencies, 
          image tags, and metadata. Charts are automatically stored for future reference.
        </p>
      </div>

      <Card className="border-2 border-dashed border-primary/20 bg-primary/5">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Registry Authentication
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="text-sm">Status:</span>
              {authStatus === 'authenticated' && (
                <Badge variant="default" className="flex items-center gap-1">
                  <CheckCircle className="h-3 w-3" />
                  Authenticated
                </Badge>
              )}
              {authStatus === 'failed' && (
                <Badge variant="destructive" className="flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  Failed
                </Badge>
              )}
              {authStatus === 'unknown' && (
                <Badge variant="outline">Not authenticated</Badge>
              )}
            </div>
            <Button onClick={authenticate} disabled={loading} variant="outline">
              {loading ? 'Authenticating...' : 'Authenticate'}
            </Button>
          </div>
          <p className="text-sm text-muted-foreground">
            Uses Docker registry credentials for private registries
          </p>
        </CardContent>
      </Card>

      <Card className="border-primary/20">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Zap className="h-5 w-5 text-primary" />
            Fetch Helm Chart
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="chartUrl">Chart URL</Label>
            <Input
              id="chartUrl"
              value={chartUrl}
              onChange={(e) => setChartUrl(e.target.value)}
              placeholder="oci://registry.k8s.io/ingress-nginx/ingress-nginx"
            />
            <p className="text-xs text-muted-foreground mb-2">
              OCI registry URL for the Helm chart
            </p>
            <div className="flex flex-wrap gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setChartUrl('oci://ghcr.io/nginxinc/charts/nginx-ingress')}
              >
                NGINX Ingress
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setChartUrl('oci://registry-1.docker.io/bitnamicharts/postgresql')}
              >
                PostgreSQL
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setChartUrl('oci://registry-1.docker.io/bitnamicharts/redis')}
              >
                Redis
              </Button>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="valuesPath">Values File Path</Label>
            <Input
              id="valuesPath"
              value={valuesPath}
              onChange={(e) => setValuesPath(e.target.value)}
              placeholder="values.yaml"
            />
            <p className="text-xs text-muted-foreground">
              Path to values file (optional)
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="setValues">Set Values</Label>
            <Textarea
              id="setValues"
              value={setValues}
              onChange={(e) => setSetValues(e.target.value)}
              placeholder="key1=value1&#10;key2=value2"
              rows={3}
            />
            <p className="text-xs text-muted-foreground">
              Override values (one per line, format: key=value)
            </p>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="useHostNetwork"
              checked={useHostNetwork}
              onCheckedChange={setUseHostNetwork}
            />
            <Label htmlFor="useHostNetwork">Use Host Network</Label>
          </div>

          {error && (
            <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}

          <Button 
            onClick={fetchChart} 
            disabled={loading || !chartUrl}
            className="w-full bg-gradient-to-r from-primary to-primary/80 hover:from-primary/90 hover:to-primary/70"
            size="lg"
          >
            {loading ? (
              <>
                <Download className="h-4 w-4 mr-2 animate-pulse" />
                Fetching Chart...
              </>
            ) : (
              <>
                <Download className="h-4 w-4 mr-2" />
                Fetch & Store Chart
              </>
            )}
          </Button>
        </CardContent>
      </Card>

      {lastFetched && (
        <Card className="border-green-200 bg-green-50 dark:bg-green-950 dark:border-green-800">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-green-800 dark:text-green-200">
              <Database className="h-5 w-5" />
              Chart Successfully Fetched & Stored
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label className="text-sm font-medium">Chart Name</Label>
                <p className="text-sm text-muted-foreground">{lastFetched.chart.chart.name}</p>
              </div>
              <div>
                <Label className="text-sm font-medium">Version</Label>
                <p className="text-sm text-muted-foreground">{lastFetched.chart.chart.version}</p>
              </div>
              <div>
                <Label className="text-sm font-medium">Image Tag</Label>
                <Badge variant={lastFetched.chart.imageTag === 'N/A' ? 'destructive' : 'outline'}>
                  {lastFetched.chart.imageTag}
                </Badge>
              </div>
              <div>
                <Label className="text-sm font-medium">Canary Tag</Label>
                <Badge variant={lastFetched.chart.canaryTag === 'N/A' ? 'destructive' : 'outline'}>
                  {lastFetched.chart.canaryTag}
                </Badge>
              </div>
            </div>
            
            {lastFetched.apps && lastFetched.apps.length > 0 && (
              <div>
                <Label className="text-sm font-medium">Applications ({lastFetched.apps.length})</Label>
                <div className="mt-2 space-y-2">
                  {lastFetched.apps.map((app, index) => (
                    <div key={index} className="p-2 border rounded-md">
                      <div className="flex justify-between items-start">
                        <div>
                          <p className="font-medium text-sm">{app.Name}</p>
                          <p className="text-xs text-muted-foreground">{app.Image}</p>
                        </div>
                        <Badge variant="outline" className="text-xs">
                          {app.Type}
                        </Badge>
                      </div>
                      {app.Ports && app.Ports.length > 0 && (
                        <div className="mt-1">
                          <p className="text-xs text-muted-foreground">
                            Ports: {app.Ports.join(', ')}
                          </p>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}