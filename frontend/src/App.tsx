import { useState, useEffect } from 'react'
import { ChartVisualizer } from './components/ChartVisualizer'
import { ChartList } from './components/ChartList'
import { ChartFetcher } from './components/ChartFetcher'
import { Card, CardContent, CardHeader, CardTitle } from './components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from './components/ui/tabs'
import { Button } from './components/ui/button'
import { RefreshCw } from 'lucide-react'
import { ChartInfo } from './types'

function App() {
  const [charts, setCharts] = useState<ChartInfo[]>([])
  const [selectedChart, setSelectedChart] = useState<ChartInfo | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchStoredCharts = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/charts')
      const data = await response.json()
      console.log('Fetched stored charts:', data)
      // Ensure data is an array
      setCharts(Array.isArray(data) ? data : [])
    } catch (error) {
      console.error('Failed to fetch stored charts:', error)
      setCharts([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    // On initial load, fetch stored charts from database
    fetchStoredCharts()
  }, [])

  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto p-6">
        <div className="mb-6">
          <h1 className="text-3xl font-bold mb-2">Chart Paper</h1>
          <p className="text-muted-foreground">
            Visualize your Helm chart dependencies, versions, and image tags
          </p>
        </div>



        <Tabs defaultValue="charts" className="mt-6">
          <TabsList>
            <TabsTrigger value="charts">Charts</TabsTrigger>
            <TabsTrigger value="fetch">Fetch from Registry</TabsTrigger>
          </TabsList>
          
          <TabsContent value="charts">
            <Card className="border-0 shadow-none">
              <CardContent className="p-0">
                <ChartVisualizer 
                  charts={charts} 
                  selectedChart={selectedChart}
                  onChartSelect={setSelectedChart}
                  onFetchDependencies={fetchStoredCharts}
                />
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value="fetch">
            <ChartFetcher 
              onChartFetched={(chart) => {
                // Refresh the stored charts list after fetching
                fetchStoredCharts()
                setSelectedChart(chart)
              }}
            />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}

export default App