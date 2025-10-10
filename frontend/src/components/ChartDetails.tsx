import { ChartInfo } from '../types'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import { Badge } from './ui/badge'
import { Package, Tag, GitBranch } from 'lucide-react'

interface ChartDetailsProps {
  chart: ChartInfo
}

export function ChartDetails({ chart }: ChartDetailsProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Package className="h-5 w-5" />
          {chart.chart.name}
        </CardTitle>
        <div className="flex gap-2">
          <Badge variant="outline">v{chart.chart.version}</Badge>
          <Badge variant={chart.chart.type === 'library' ? 'secondary' : 'default'}>
            {chart.chart.type || 'application'}
          </Badge>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {chart.chart.description && (
          <div>
            <h4 className="font-medium mb-1">Description</h4>
            <p className="text-sm text-muted-foreground">{chart.chart.description}</p>
          </div>
        )}

        <div>
          <h4 className="font-medium mb-2 flex items-center gap-2">
            <Tag className="h-4 w-4" />
            Image Tags
          </h4>
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <span className="text-sm">Image Tag:</span>
              <Badge variant={chart.imageTag === 'N/A' ? 'destructive' : 'outline'}>
                {chart.imageTag}
              </Badge>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm">Canary Tag:</span>
              <Badge variant={chart.canaryTag === 'N/A' ? 'destructive' : 'outline'}>
                {chart.canaryTag}
              </Badge>
            </div>
          </div>
        </div>

        {chart.chart.dependencies && chart.chart.dependencies.length > 0 && (
          <div>
            <h4 className="font-medium mb-2 flex items-center gap-2">
              <GitBranch className="h-4 w-4" />
              Dependencies ({chart.chart.dependencies.length})
            </h4>
            <div className="space-y-2">
              {chart.chart.dependencies.map((dep, index) => (
                <div key={index} className="border rounded-lg p-3">
                  <div className="flex justify-between items-start mb-1">
                    <span className="font-medium text-sm">{dep.name}</span>
                    <Badge variant="outline" className="text-xs">
                      {dep.version}
                    </Badge>
                  </div>
                  {dep.repository && (
                    <p className="text-xs text-muted-foreground mb-1">
                      {dep.repository}
                    </p>
                  )}
                  {dep.condition && (
                    <p className="text-xs text-muted-foreground">
                      Condition: {dep.condition}
                    </p>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {(!chart.chart.dependencies || chart.chart.dependencies.length === 0) && (
          <div className="text-center py-4 text-muted-foreground">
            <Package className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">No dependencies</p>
          </div>
        )}
      </CardContent>
    </Card>
  )
}