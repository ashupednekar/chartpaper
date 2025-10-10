import { useEffect, useRef, useState } from 'react'
import { ChartInfo, Node, Edge } from '../types'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { Package, Trash2, Download, History, ChevronDown } from 'lucide-react'
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
  const [nodes, setNodes] = useState<Node[]>([])
  const [edges, setEdges] = useState<Edge[]>([])
  const [expandedCharts, setExpandedCharts] = useState<Set<string>>(new Set())
  const [chartVersions, setChartVersions] = useState<Record<string, any[]>>({})
  const [fetchingDeps, setFetchingDeps] = useState<Set<string>>(new Set())
  
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

    // Create root nodes for each chart with better spacing
    const cardWidth = 300 // Approximate card width
    const padding = 50 // Padding between cards
    
    charts.forEach((chartInfo, index) => {
      let x, y
      
      if (charts.length === 1) {
        // Single chart - center it
        x = 600
        y = 350
      } else if (charts.length === 2) {
        // Two charts - side by side
        x = index === 0 ? 400 : 800
        y = 350
      } else {
        // Multiple charts - circular layout with better spacing
        const angle = (index / charts.length) * 2 * Math.PI
        const radius = Math.max(300, (cardWidth + padding) * charts.length / (2 * Math.PI))
        x = 600 + Math.cos(angle) * radius
        y = 350 + Math.sin(angle) * radius
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
        const parentNode = newNodes[newNodes.length - 1] // Get the parent node we just added
        const centerX = parentNode.x
        const centerY = parentNode.y
        const depCount = chartInfo.chart.dependencies.length
        
        chartInfo.chart.dependencies.forEach((dep, depIndex) => {
          // Better circular positioning - start from top and go clockwise
          const depAngle = (depIndex / depCount) * 2 * Math.PI - Math.PI / 2 // Start from top (-π/2)
          const depRadius = Math.max(120, 80 + depCount * 8) // Dynamic radius based on dependency count
          
          newNodes.push({
            id: `${chartInfo.chart.name}-${dep.name}`,
            name: dep.name,
            version: dep.version,
            type: 'library',
            imageTag: 'latest', // We'll fetch this dynamically later
            canaryTag: 'N/A',   // We'll fetch this dynamically later
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
    })

    setNodes(newNodes)
    setEdges(newEdges)
  }, [charts, expandedCharts])

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
    if (node.isRoot && node.dependencies.length > 0) {
      // Double-click to toggle expansion and fetch dependencies if needed
      const wasExpanded = expandedCharts.has(node.id)
      toggleExpansion(node.id)
      
      // If expanding for the first time, try to fetch dependency details
      if (!wasExpanded) {
        const chart = charts.find(c => c.chart.name === node.id)
        if (chart?.chart.dependencies) {
          // Update dependency nodes with fetched tags
          const updatedNodes = await Promise.all(
            chart.chart.dependencies.map(async (dep) => {
              if (dep.repository) {
                const tags = await fetchDependencyTags(dep.name, dep.repository)
                return { ...dep, ...tags }
              }
              return dep
            })
          )
          
          // Update the nodes state with the new tag information
          setNodes(prevNodes => 
            prevNodes.map(n => {
              if (n.id.startsWith(`${node.id}-`)) {
                const depName = n.id.split('-').slice(1).join('-')
                const updatedDep = updatedNodes.find(d => d.name === depName)
                if (updatedDep) {
                  return { ...n, imageTag: updatedDep.imageTag, canaryTag: updatedDep.canaryTag }
                }
              }
              return n
            })
          )
        }
      }
    }
  }

  const toggleExpansion = (nodeId: string) => {
    const newExpanded = new Set(expandedCharts)
    if (newExpanded.has(nodeId)) {
      newExpanded.delete(nodeId)
    } else {
      newExpanded.add(nodeId)
    }
    setExpandedCharts(newExpanded)
  }

  const deleteChart = async (chartName: string) => {
    try {
      const response = await fetch(`/api/charts/${chartName}`, { 
        method: 'DELETE' 
      })
      
      if (response.ok) {
        // Refresh the charts list
        onFetchDependencies()
        // Clear selection if deleted chart was selected
        if (selectedChart?.chart.name === chartName) {
          onChartSelect(null)
        }
      } else {
        console.error('Failed to delete chart')
      }
    } catch (error) {
      console.error('Error deleting chart:', error)
    }
  }

  const fetchChartDependencies = async (chartName: string) => {
    setFetchingDeps(prev => new Set([...prev, chartName]))
    
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
    } finally {
      setFetchingDeps(prev => {
        const newSet = new Set(prev)
        newSet.delete(chartName)
        return newSet
      })
    }
  }

  const fetchDependencyTags = async (depName: string, repository: string) => {
    try {
      // Try to fetch the dependency chart to get image/canary tags
      const chartUrl = repository.endsWith(`/${depName}`) ? repository : `${repository}/${depName}`
      const response = await fetch('/api/fetch-chart', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          chartUrl,
          valuesPath: 'values',
          setValues: [],
          useHostNetwork: false,
        }),
      })
      
      if (response.ok) {
        const result = await response.json()
        return {
          imageTag: result.chart.imageTag || 'N/A',
          canaryTag: result.chart.canaryTag || 'N/A'
        }
      }
    } catch (error) {
      console.error('Error fetching dependency tags:', error)
    }
    
    return { imageTag: 'N/A', canaryTag: 'N/A' }
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
    <div className="w-full h-full">
      <div 
        ref={containerRef}
        className="relative w-full h-[700px] bg-gray-100 dark:bg-gray-800 border border-border rounded-lg overflow-hidden cursor-grab active:cursor-grabbing"
        onMouseDown={(e) => handleMouseDown(e)}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        onWheel={handleWheel}
      >
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
              className={`absolute transition-all duration-200 cursor-move ${
                node.isRoot ? 'w-72' : 'w-48'
              } ${dragState.draggedNode?.id === node.id ? 'cursor-grabbing' : 'cursor-grab'}`}
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
              bg-gray-900/95 backdrop-blur-sm border border-gray-700 rounded-lg shadow-lg hover:shadow-xl transition-all duration-200
              ${selectedChart?.chart.name === node.id ? 'ring-2 ring-primary border-primary' : 'hover:border-primary/50'}
              ${node.isRoot ? 'p-4' : 'p-3'}
            `}>
              {node.isRoot ? (
                // Root chart node - full card
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
                            className="h-5 px-2 text-xs flex-shrink-0"
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
                    <button
                      className="w-6 h-6 rounded-full flex items-center justify-center text-destructive hover:bg-destructive hover:text-destructive-foreground transition-colors ml-2 flex-shrink-0"
                      onClick={(e) => {
                        e.stopPropagation()
                        deleteChart(node.name)
                      }}
                      title="Delete chart"
                    >
                      <Trash2 className="h-3 w-3" />
                    </button>
                  </div>
                  
                  {charts.find(c => c.chart.name === node.name)?.chart.description && (
                    <p className="text-xs text-muted-foreground line-clamp-2">
                      {charts.find(c => c.chart.name === node.name)?.chart.description}
                    </p>
                  )}

                  <div className="grid grid-cols-2 gap-2 text-xs">
                    <div>
                      <span className="text-muted-foreground">Image:</span>
                      <Badge variant={node.imageTag === 'N/A' ? 'destructive' : 'outline'} className="ml-1 text-xs">
                        {node.imageTag}
                      </Badge>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Canary:</span>
                      <Badge variant={node.canaryTag === 'N/A' ? 'destructive' : 'outline'} className="ml-1 text-xs">
                        {node.canaryTag}
                      </Badge>
                    </div>
                  </div>

                  {node.dependencies.length > 0 ? (
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
                        disabled={fetchingDeps.has(node.name)}
                      >
                        {fetchingDeps.has(node.name) ? (
                          <>
                            <Download className="h-3 w-3 mr-1 animate-spin" />
                            Fetching...
                          </>
                        ) : (
                          <>
                            <Download className="h-3 w-3 mr-1" />
                            Fetch Deps
                          </>
                        )}
                      </Button>
                    </div>
                  ) : (
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <Package className="h-3 w-3" />
                      <span>No dependencies</span>
                    </div>
                  )}
                </div>
              ) : (
                // Dependency node - compact card with tags
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
        <div className="absolute top-4 left-4 bg-card/90 backdrop-blur-sm border border-border rounded-lg p-3 shadow-sm">
          <div className="space-y-2">
            <div className="flex gap-4 text-xs text-muted-foreground">
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded bg-blue-500"></div>
                <span>App</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded bg-purple-500"></div>
                <span>Lib</span>
              </div>
            </div>
            <div className="text-xs text-muted-foreground space-y-1">
              <div>• Drag cards to move</div>
              <div>• Drag background to pan</div>
              <div>• Scroll to zoom</div>
              <div>• Double-click to expand</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}