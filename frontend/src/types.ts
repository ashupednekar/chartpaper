export interface Dependency {
  name: string
  version: string
  repository: string
  condition?: string
}

export interface Chart {
  apiVersion: string
  name: string
  version: string
  description: string
  type: string
  dependencies: Dependency[]
}

export interface ChartInfo {
  chart: Chart
  imageTag: string
  canaryTag: string
}

export interface Node {
  id: string
  name: string
  version: string
  type: 'application' | 'library'
  imageTag: string
  canaryTag: string
  x: number
  y: number
  dependencies: string[]
  expanded: boolean
  isRoot: boolean
}

export interface Edge {
  from: string
  to: string
  version: string
  repository?: string
}