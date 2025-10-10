import { useEffect, useRef, useState } from 'react'
import { ChartInfo, Node, Edge } from '../types'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { Package, Download, Plus, X, ChevronDown, History } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'

interface ChartVisualizerProps {
  charts: ChartInfo[]
  selectedChart: ChartInfo | null
  onChartSelect: (chart: ChartInfo | null) => void
  onFetchDependencies: () => void
}

export function ChartVisualizer({ charts, selectedChart, onChartSelect, onFetchDependencies }: ChartVisualizerProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  
  // Canvas state - charts that are placed on the canvas
  const [canvasCharts, setCanvasCharts] = useState<Set<string>>(new Set())
  const [expandedCharts, setExpandedCharts] = useState<Set<string>>(new Set())
  const [chartVersions, setChartVersions] = useState<Record<string, any[]>>({})
  const [nodes, setNodes] = useState<Node[]>([])
  const [edges, setEdges] = useState<Edge[]>([])
  
  // Pan and zoom state
  const [viewState, setViewState] = useState({
    offsetX: 0,
    offsetY: 0,
    scale: 1
  })
  
  // Drag state
  const [dragState, setDragState] = useState({
    isDragging: false,
    draggedNode: null as Node | null,
    startX: 0,
    startY: 0,
    isPanning: false
  })

  // Sidebar visibility state
  const [showAvailableCharts, setShowAvailableCharts] = useState(true)

  useEffect(() => {
    const newNodes: Node[] = []
    const newEdges: Edge[] = []

    // Ensure charts is an array
    if (!Array.isArray(charts)) {
      console.warn('Charts is not an array:', charts)
      setNodes([])
      setEdges([])
      return
    }

    // Only show charts that are added to the canvas
    const canvasChartsArray = charts.filter(chart => canvasCharts.has(chart.chart.name))
    
    // Create nodes for charts on canvas with smart positioning
    canvasChartsArray.forEach((chartInfo, index) => {
      let x, y
      
      if (canvasChartsArray.length === 1) {
        // Single chart - center it
        x = 600
        y = 350
      } else {
        // Multiple charts - grid layout
        const cols = Math.ceil(Math.sqrt(canvasChartsArray.length))
        const row = Math.floor(index / cols)
        const col = index % cols
        x = 200 + col * 400
        y = 200 + row * 300
      }
      
      newNodes.push({
        id: chartInfo.chart.name,
        name: chartInfo.chart.name,
        version: chartInfo.chart.version,
        type: chartInfo.chart.type === 'library' ? 'library' : 'application',
        imageTag: chartInfo.imageTag,
        canaryTag: chartInfo.canaryTag,
        x: x,
        y: y,
        dependencies: chartInfo.chart.dependencies?.map(d => d.name) || [],
        expanded: expandedCharts.has(chartInfo.chart.name),
        isRoot: true
      })

      // If chart is expanded, add dependency nodes
      if (expandedCharts.has(chartInfo.chart.name) && chartInfo.chart.dependencies) {
        const parentNode = newNodes[newNodes.length - 1]
        const centerX = parentNode.x
        const centerY = parentNode.y
        const depCount = chartInfo.chart.dependencies.length
        
        chartInfo.chart.dependencies.forEach((dep, depIndex) => {
          // Circular positioning around parent - moved further away
          const depAngle = (depIndex / depCount) * 2 * Math.PI - Math.PI / 2
          const depRadius = Math.max(250, 200 + depCount * 15)
          
          newNodes.push({
            id: `${chartInfo.chart.name}-${dep.name}`,
            name: dep.name,
            version: dep.version,
            type: 'library',
            imageTag: 'N/A',
            canaryTag: 'N/A',
            x: centerX + Math.cos(depAngle) * depRadius,
            y: centerY + Math.sin(depAngle) * depRadius,
            dependencies: [],
            expanded: false,
            isRoot: false
          })

          // Create edge from parent to dependency
          newEdges.push({
            from: chartInfo.chart.name,
            to: `${chartInfo.chart.name}-${dep.name}`,
            version: dep.version,
            repository: dep.repository
          })
        })
      }

      // Add dependency connections between root charts on canvas
      if (chartInfo.chart.dependencies) {
        chartInfo.chart.dependencies.forEach((dep) => {
          // Check if the dependency is also on the canvas as a root chart
          if (canvasCharts.has(dep.name)) {
            newEdges.push({
              from: chartInfo.chart.name,
              to: dep.name,
              version: dep.version,
              repository: dep.repository
            })
          }
        })
      }
    })

    setNodes(newNodes)
    setEdges(newEdges)
  }, [charts, canvasCharts, expandedCharts])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    // Clear canvas
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    
    // Draw edges only
    edges.forEach(edge => {
      const fromNode = nodes.find(n => n.id === edge.from)
      const toNode = nodes.find(n => n.id === edge.to)
      
      if (fromNode && toNode) {
        ctx.beginPath()
        ctx.moveTo(fromNode.x, fromNode.y)
        ctx.lineTo(toNode.x, toNode.y)
        ctx.strokeStyle = '#64748b'
        ctx.lineWidth = 2
        ctx.stroke()

        // Draw arrow
        const angle = Math.atan2(toNode.y - fromNode.y, toNode.x - fromNode.x)
        const arrowLength = 10
        ctx.beginPath()
        ctx.moveTo(toNode.x, toNode.y)
        ctx.lineTo(
          toNode.x - arrowLength * Math.cos(angle - Math.PI / 6),
          toNode.y - arrowLength * Math.sin(angle - Math.PI / 6)
        )
        ctx.moveTo(toNode.x, toNode.y)
        ctx.lineTo(
          toNode.x - arrowLength * Math.cos(angle + Math.PI / 6),
          toNode.y - arrowLength * Math.sin(angle + Math.PI / 6)
        )
        ctx.stroke()
      }
    })
  }, [nodes, edges])

  const addChartToCanvas = (chartName: string) => {
    if (!canvasCharts.has(chartName)) {
      setCanvasCharts(prev => new Set([...prev, chartName]))
    }
  }

  const removeChartFromCanvas = (chartName: string) => {
    setCanvasCharts(prev => {
      const newSet = new Set(prev)
      newSet.delete(chartName)
      return newSet
    })
    // Also remove from expanded when removed from canvas
    setExpandedCharts(prev => {
      const newSet = new Set(prev)
      newSet.delete(chartName)
      return newSet
    })
  }

  const toggleExpansion = (chartName: string) => {
    setExpandedCharts(prev => {
      const newSet = new Set(prev)
      if (newSet.has(chartName)) {
        newSet.delete(chartName)
      } else {
        newSet.add(chartName)
      }
      return newSet
    })
  }

  const handleNodeClick = (node: Node) => {
    if (node.isRoot) {
      const chart = charts.find(c => c.chart.name === node.id)
      onChartSelect(chart || null)
    } else {
      // Handle dependency node click - show parent chart details
      const parentChartName = node.id.split('-')[0]
      const chart = charts.find(c => c.chart.name === parentChartName)
      onChartSelect(chart || null)
    }
  }

  const handleMouseDown = (e: React.MouseEvent, node?: Node) => {
    const rect = containerRef.current?.getBoundingClientRect()
    if (!rect) return

    const x = e.clientX - rect.left
    const y = e.clientY - rect.top

    if (node) {
      // Start dragging a node
      setDragState({
        isDragging: true,
        draggedNode: node,
        startX: x,
        startY: y,
        isPanning: false
      })
    } else {
      // Start panning the canvas
      setDragState({
        isDragging: false,
        draggedNode: null,
        startX: x,
        startY: y,
        isPanning: true
      })
    }
  }

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!dragState.isDragging && !dragState.isPanning) return

    const rect = containerRef.current?.getBoundingClientRect()
    if (!rect) return

    const x = e.clientX - rect.left
    const y = e.clientY - rect.top
    const deltaX = x - dragState.startX
    const deltaY = y - dragState.startY

    if (dragState.isDragging && dragState.draggedNode) {
      // Update node position
      setNodes(prevNodes => 
        prevNodes.map(node => 
          node.id === dragState.draggedNode?.id
            ? { ...node, x: node.x + deltaX, y: node.y + deltaY }
            : node
        )
      )
      
      setDragState(prev => ({
        ...prev,
        startX: x,
        startY: y
      }))
    } else if (dragState.isPanning) {
      // Update view offset for panning
      setViewState(prev => ({
        ...prev,
        offsetX: prev.offsetX + deltaX,
        offsetY: prev.offsetY + deltaY
      }))
      
      setDragState(prev => ({
        ...prev,
        startX: x,
        startY: y
      }))
    }
  }

  const handleMouseUp = () => {
    setDragState({
      isDragging: false,
      draggedNode: null,
      startX: 0,
      startY: 0,
      isPanning: false
    })
  }

  const handleWheel = (e: React.WheelEvent) => {
    e.preventDefault()
    const scaleFactor = e.deltaY > 0 ? 0.9 : 1.1
    setViewState(prev => ({
      ...prev,
      scale: Math.max(0.5, Math.min(2, prev.scale * scaleFactor))
    }))
  }

  const handleNodeDoubleClick = async (node: Node) => {
    // Double-click to fetch dependencies
    if (node.isRoot && node.dependencies.length > 0) {
      await fetchChartDependencies(node.id)
    }
  }

  const fetchChartDependencies = async (chartName: string) => {
    try {
      const response = await fetch(`/api/charts/${chartName}/fetch-dependencies`, {
        method: 'POST'
      })
      
      if (response.ok) {
        const result = await response.json()
        console.log('Dependencies fetched:', result)
        // Refresh the charts to show newly fetched dependencies
        onFetchDependencies()
      } else {
        console.error('Failed to fetch dependencies')
      }
    } catch (error) {
      console.error('Error fetching dependencies:', error)
    }
  }

  const loadChartVersions = async (chartName: string) => {
    try {
      const response = await fetch(`/api/charts/${chartName}/versions`)
      if (response.ok) {
        const result = await response.json()
        setChartVersions(prev => ({
          ...prev,
          [chartName]: result.versions
        }))
      }
    } catch (error) {
      console.error('Error loading chart versions:', error)
    }
  }

  const switchChartVersion = async (chartName: string, version: string) => {
    try {
      const response = await fetch(`/api/charts/${chartName}/switch-version`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ version })
      })
      
      if (response.ok) {
        console.log(`Switched ${chartName} to version ${version}`)
        // Refresh the charts to show the new version
        onFetchDependencies()
      } else {
        console.error('Failed to switch version')
      }
    } catch (error) {
      console.error('Error switching version:', error)
    }
  }



  return (
    <div className="w-full h-full flex gap-4">
      {/* Chart List Sidebar */}
      <div className={`${showAvailableCharts ? 'w-80' : 'w-auto'} bg-card border border-border rounded-lg p-4 overflow-y-auto transition-all duration-200`}>
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-semibold flex items-center gap-2">
            <Package className="h-4 w-4" />
            Available Charts
          </h3>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => setShowAvailableCharts(!showAvailableCharts)}
            className="h-6 w-6 p-0"
            title={showAvailableCharts ? "Hide available charts" : "Show available charts"}
          >
            <ChevronDown className={`h-4 w-4 transition-transform ${showAvailableCharts ? '' : 'rotate-180'}`} />
          </Button>
        </div>
        
        {showAvailableCharts && (
          <div className="space-y-2">
            {charts.map((chart) => (
            <div
              key={chart.chart.name}
              className={`p-3 border rounded-lg transition-all ${
                canvasCharts.has(chart.chart.name)
                  ? 'border-primary bg-primary/5'
                  : 'border-border hover:border-primary/50'
              }`}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2 flex-1 min-w-0">
                  <Package className="h-3 w-3 text-primary flex-shrink-0" />
                  <span className="font-medium text-sm truncate">{chart.chart.name}</span>
                </div>
                {canvasCharts.has(chart.chart.name) ? (
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => removeChartFromCanvas(chart.chart.name)}
                    className="h-6 w-6 p-0 flex-shrink-0"
                  >
                    <X className="h-3 w-3" />
                  </Button>
                ) : (
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => addChartToCanvas(chart.chart.name)}
                    className="h-6 w-6 p-0 flex-shrink-0"
                  >
                    <Plus className="h-3 w-3" />
                  </Button>
                )}
              </div>
              
              <div className="text-xs text-muted-foreground space-y-1">
                <div>v{chart.chart.version}</div>
                {chart.chart.dependencies && chart.chart.dependencies.length > 0 && (
                  <div>{chart.chart.dependencies.length} dependencies</div>
                )}
                {chart.manifestMetadata?.containerImages && chart.manifestMetadata.containerImages.length > 0 && (
                  <div>{chart.manifestMetadata.containerImages.length} container images</div>
                )}
                {chart.manifestMetadata?.ingressPaths && chart.manifestMetadata.ingressPaths.length > 0 && (
                  <div>{chart.manifestMetadata.ingressPaths.length} ingress paths</div>
                )}
              </div>
              
              <div className="flex gap-1 mt-2">
                <Badge variant="outline" className="text-xs px-1 py-0">
                  {chart.imageTag}
                </Badge>
                {chart.canaryTag !== 'N/A' && (
                  <Badge variant="secondary" className="text-xs px-1 py-0">
                    {chart.canaryTag}
                  </Badge>
                )}
              </div>
            </div>
          ))}
          
            {charts.length === 0 && (
              <div className="text-center py-8 text-muted-foreground">
                <Package className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p className="text-sm">No charts available</p>
                <p className="text-xs">Fetch some charts to get started</p>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Canvas Area */}
      <div className="flex-1">
        <div 
          ref={containerRef}
          className="relative w-full h-[700px] bg-gray-50 dark:bg-gray-900 border border-border rounded-lg overflow-hidden cursor-grab active:cursor-grabbing"
          onMouseDown={(e) => handleMouseDown(e)}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          onWheel={handleWheel}
        >
          {/* Empty state */}
          {canvasCharts.size === 0 && (
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="text-center text-muted-foreground">
                <Package className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <h3 className="text-lg font-medium mb-2">Canvas is empty</h3>
                <p className="text-sm">Click the + button next to charts to add them to the canvas</p>
                <p className="text-xs mt-2">You can then drag them around and see their connections</p>
              </div>
            </div>
          )}

          {/* Background canvas for connections */}
          <canvas
            ref={canvasRef}
            width={1200}
            height={700}
            className="absolute inset-0 w-full h-full pointer-events-none"
            style={{ maxWidth: '100%', height: '100%' }}
          />
          
          {/* Chart nodes as beautiful cards */}
          <div 
            style={{
              transform: `translate(${viewState.offsetX}px, ${viewState.offsetY}px) scale(${viewState.scale})`,
              transformOrigin: 'center center',
              width: '100%',
              height: '100%'
            }}
          >
            {nodes.map((node) => (
              <div
                key={node.id}
                className={`absolute transition-all duration-200 cursor-move w-64 ${
                  dragState.draggedNode?.id === node.id ? 'cursor-grabbing' : 'cursor-grab'
                }`}
                style={{
                  left: `${(node.x / 1200) * 100}%`,
                  top: `${(node.y / 700) * 100}%`,
                  transform: 'translate(-50%, -50%)',
                  zIndex: selectedChart?.chart.name === node.id ? 20 : 10,
                }}
                onClick={() => handleNodeClick(node)}
                onDoubleClick={() => handleNodeDoubleClick(node)}
                onMouseDown={(e) => {
                  e.stopPropagation()
                  handleMouseDown(e, node)
                }}
              >
                <div className={`
                  bg-card border border-border rounded-lg shadow-lg hover:shadow-xl transition-all duration-200 p-4
                  ${selectedChart?.chart.name === node.id ? 'ring-2 ring-primary border-primary' : 'hover:border-primary/50'}
                `}>
                  {node.isRoot ? (
                    // Root chart node
                    <div className="space-y-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2 flex-1 min-w-0">
                          <Package className="h-4 w-4 text-primary flex-shrink-0" />
                          <h3 className="font-semibold text-sm truncate">{node.name}</h3>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button 
                                variant="outline" 
                                size="sm" 
                                className="h-6 px-2 text-xs flex-shrink-0"
                                onClick={(e) => {
                                  e.stopPropagation()
                                  loadChartVersions(node.name)
                                }}
                              >
                                v{node.version}
                                <ChevronDown className="h-3 w-3 ml-1" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent>
                              {chartVersions[node.name]?.map((version: any) => (
                                <DropdownMenuItem
                                  key={version.version}
                                  onClick={(e: React.MouseEvent) => {
                                    e.stopPropagation()
                                    switchChartVersion(node.name, version.version)
                                  }}
                                  className={version.is_latest ? 'bg-primary/10' : ''}
                                >
                                  <div className="flex items-center justify-between w-full">
                                    <span>v{version.version}</span>
                                    {version.is_latest && (
                                      <Badge variant="secondary" className="text-xs ml-2">current</Badge>
                                    )}
                                  </div>
                                </DropdownMenuItem>
                              ))}
                              {(!chartVersions[node.name] || chartVersions[node.name].length === 0) && (
                                <DropdownMenuItem disabled>
                                  <History className="h-3 w-3 mr-2" />
                                  Loading versions...
                                </DropdownMenuItem>
                              )}
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={(e) => {
                            e.stopPropagation()
                            removeChartFromCanvas(node.name)
                          }}
                          className="h-6 w-6 p-0 flex-shrink-0"
                        >
                          <X className="h-3 w-3" />
                        </Button>
                      </div>
                      
                      {charts.find(c => c.chart.name === node.name)?.chart.description && (
                        <p className="text-xs text-muted-foreground line-clamp-2">
                          {charts.find(c => c.chart.name === node.name)?.chart.description}
                        </p>
                      )}

                      <div className="space-y-2">
                        {/* Main image tags */}
                        <div className="flex gap-2 text-xs">
                          <Badge variant={node.imageTag === 'N/A' ? 'destructive' : 'outline'} className="text-xs">
                            {node.imageTag}
                          </Badge>
                          {node.canaryTag !== 'N/A' && (
                            <Badge variant="secondary" className="text-xs">
                              {node.canaryTag}
                            </Badge>
                          )}
                        </div>
                        
                        {/* Container images from manifest metadata */}
                        {(() => {
                          const chart = charts.find(c => c.chart.name === node.name)
                          const metadata = chart?.manifestMetadata
                          if (!metadata?.containerImages || metadata.containerImages.length === 0) return null
                          
                          return (
                            <div className="space-y-1">
                              <div className="text-xs text-muted-foreground font-medium">Container Images:</div>
                              <div className="flex flex-wrap gap-1">
                                {metadata.containerImages.map((image, idx) => {
                                  // Extract just the image name and tag for display
                                  const displayName = image.includes('/') 
                                    ? image.split('/').pop() || image 
                                    : image
                                  
                                  return (
                                    <Badge 
                                      key={idx} 
                                      variant="outline" 
                                      className="text-xs px-1 py-0 max-w-32 truncate" 
                                      title={image}
                                    >
                                      {displayName}
                                    </Badge>
                                  )
                                })}
                              </div>
                            </div>
                          )
                        })()}
                        
                        {/* Ingress paths */}
                        {(() => {
                          const chart = charts.find(c => c.chart.name === node.name)
                          const metadata = chart?.manifestMetadata
                          if (!metadata?.ingressPaths || metadata.ingressPaths.length === 0) return null
                          
                          return (
                            <div className="space-y-1">
                              <div className="text-xs text-muted-foreground font-medium">Ingress Paths:</div>
                              <div className="flex flex-wrap gap-1">
                                {metadata.ingressPaths.slice(0, 3).map((path, idx) => (
                                  <Badge key={idx} variant="secondary" className="text-xs px-1 py-0" title={path}>
                                    {path}
                                  </Badge>
                                ))}
                                {metadata.ingressPaths.length > 3 && (
                                  <Badge variant="secondary" className="text-xs px-1 py-0">
                                    +{metadata.ingressPaths.length - 3}
                                  </Badge>
                                )}
                              </div>
                            </div>
                          )
                        })()}
                        
                        {/* Service ports */}
                        {(() => {
                          const chart = charts.find(c => c.chart.name === node.name)
                          const metadata = chart?.manifestMetadata
                          if (!metadata?.servicePorts || metadata.servicePorts.length === 0) return null
                          
                          return (
                            <div className="space-y-1">
                              <div className="text-xs text-muted-foreground font-medium">Service Ports:</div>
                              <div className="flex flex-wrap gap-1">
                                {metadata.servicePorts.slice(0, 4).map((port, idx) => (
                                  <Badge key={idx} variant="outline" className="text-xs px-1 py-0">
                                    {port}
                                  </Badge>
                                ))}
                                {metadata.servicePorts.length > 4 && (
                                  <Badge variant="outline" className="text-xs px-1 py-0">
                                    +{metadata.servicePorts.length - 4}
                                  </Badge>
                                )}
                              </div>
                            </div>
                          )
                        })()}
                      </div>

                      {node.dependencies.length > 0 && (
                        <div className="space-y-2">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2 text-xs text-muted-foreground">
                              <Package className="h-3 w-3" />
                              <span>{node.dependencies.length} dependencies</span>
                            </div>
                            <button
                              className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold transition-colors ${
                                node.expanded 
                                  ? 'bg-primary text-primary-foreground' 
                                  : 'bg-muted text-muted-foreground hover:bg-primary hover:text-primary-foreground'
                              }`}
                              onClick={(e) => {
                                e.stopPropagation()
                                toggleExpansion(node.id)
                              }}
                              title={node.expanded ? 'Collapse dependencies' : 'Expand dependencies'}
                            >
                              {node.expanded ? '−' : '+'}
                            </button>
                          </div>
                          <Button
                            size="sm"
                            variant="outline"
                            className="w-full h-6 text-xs"
                            onClick={(e) => {
                              e.stopPropagation()
                              fetchChartDependencies(node.name)
                            }}
                          >
                            <Download className="h-3 w-3 mr-1" />
                            Fetch Deps
                          </Button>
                        </div>
                      )}
                    </div>
                  ) : (
                    // Dependency node - compact card
                    <div className="space-y-2">
                      <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-purple-500"></div>
                        <h4 className="font-medium text-xs truncate">{node.name}</h4>
                      </div>
                      <Badge variant="secondary" className="text-xs">v{node.version}</Badge>
                      
                      <div className="space-y-1 text-xs">
                        <div className="flex items-center gap-1">
                          <span className="text-muted-foreground">Repo:</span>
                          <Badge variant="outline" className="text-xs px-1 py-0 truncate max-w-24" title={
                            edges.find(e => e.to === node.id)?.repository || 'Unknown'
                          }>
                            {edges.find(e => e.to === node.id)?.repository?.split('/').pop() || 'Unknown'}
                          </Badge>
                        </div>
                        <Button
                          size="sm"
                          variant="outline"
                          className="w-full h-5 text-xs px-2"
                          onClick={async (e: React.MouseEvent) => {
                            e.stopPropagation()
                            const edge = edges.find(e => e.to === node.id)
                            if (edge?.repository) {
                              const chartUrl = edge.repository.endsWith(`/${node.name}`) 
                                ? edge.repository 
                                : `${edge.repository}/${node.name}`
                              
                              try {
                                const response = await fetch('/api/fetch-chart', {
                                  method: 'POST',
                                  headers: { 'Content-Type': 'application/json' },
                                  body: JSON.stringify({
                                    chartUrl,
                                    valuesPath: 'values',
                                    setValues: [],
                                    useHostNetwork: false,
                                  }),
                                })
                                
                                if (response.ok) {
                                  console.log(`Fetched dependency: ${node.name}`)
                                  onFetchDependencies()
                                }
                              } catch (error) {
                                console.error('Error fetching dependency:', error)
                              }
                            }
                          }}
                        >
                          <Download className="h-2 w-2 mr-1" />
                          Fetch
                        </Button>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
          
          {/* Legend and Controls */}
          {canvasCharts.size > 0 && (
            <div className="absolute top-4 left-4 bg-card/90 backdrop-blur-sm border border-border rounded-lg p-3 shadow-sm">
              <div className="space-y-2">
                <div className="text-xs font-medium">Canvas Controls</div>
                <div className="text-xs text-muted-foreground space-y-1">
                  <div>• Drag cards to move</div>
                  <div>• Drag background to pan</div>
                  <div>• Scroll to zoom</div>
                  <div>• Double-click to fetch deps</div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}