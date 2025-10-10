import { ChartInfo } from '../types'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { Package, Tag, ExternalLink, Trash2, RefreshCw } from 'lucide-react'

interface ChartListProps {
  charts: ChartInfo[]
  selectedChart: ChartInfo | null
  onChartSelect: (chart: ChartInfo) => void
  onRefresh: () => void
}

const deleteChart = async (chartName: string) => {
  try {
    await fetch(`/api/charts/${chartName}`, { method: 'DELETE' })
    window.location.reload() // Simple refresh for now
  } catch (error) {
    console.error('Failed to delete chart:', error)
  }
}

export function ChartList({ charts, selectedChart, onChartSelect, onRefresh }: ChartListProps) {
  if (charts.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between text-lg">
            <span>Charts</span>
            <Button variant="outline" size="sm" onClick={onRefresh}>
              <RefreshCw className="h-4 w-4" />
            </Button>
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center justify-center py-8">
          <Package className="h-8 w-8 text-muted-foreground mb-2" />
          <p className="text-sm text-muted-foreground text-center">
            No charts found. Use "Fetch from Registry" to add charts.
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between text-lg">
          <span>Charts ({charts.length})</span>
          <Button variant="outline" size="sm" onClick={onRefresh}>
            <RefreshCw className="h-4 w-4" />
          </Button>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-2 max-h-96 overflow-y-auto">
        {charts.map((chartInfo) => (
          <div
            key={chartInfo.chart.name}
            className={`p-3 border rounded-lg cursor-pointer transition-colors hover:bg-accent ${
              selectedChart?.chart.name === chartInfo.chart.name ? 'ring-2 ring-primary bg-accent' : ''
            }`}
            onClick={() => onChartSelect(chartInfo)}
          >
            <div className="flex items-center gap-2 mb-2">
              <Package className="h-4 w-4" />
              <span className="font-medium text-sm">{chartInfo.chart.name}</span>
            </div>
            
            <div className="flex gap-1 mb-2">
              <Badge variant="outline" className="text-xs">v{chartInfo.chart.version}</Badge>
              <Badge variant={chartInfo.chart.type === 'library' ? 'secondary' : 'default'} className="text-xs">
                {chartInfo.chart.type || 'app'}
              </Badge>
            </div>

            <div className="space-y-1">
              <div className="flex items-center justify-between text-xs">
                <span className="text-muted-foreground">Image:</span>
                <Badge variant={chartInfo.imageTag === 'N/A' ? 'destructive' : 'outline'} className="text-xs">
                  {chartInfo.imageTag}
                </Badge>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-muted-foreground">Canary:</span>
                <Badge variant={chartInfo.canaryTag === 'N/A' ? 'destructive' : 'outline'} className="text-xs">
                  {chartInfo.canaryTag}
                </Badge>
              </div>
            </div>

            {chartInfo.chart.dependencies && chartInfo.chart.dependencies.length > 0 && (
              <div className="mt-2">
                <p className="text-xs text-muted-foreground">
                  {chartInfo.chart.dependencies.length} dependencies
                </p>
              </div>
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  )
}