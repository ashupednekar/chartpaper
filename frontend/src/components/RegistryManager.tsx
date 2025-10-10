import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Badge } from './ui/badge'
import { Switch } from './ui/switch'
import { Textarea } from './ui/textarea'
import { 
  Plus, 
  Edit, 
  Trash2, 
  Star, 
  Upload, 
  Database, 
  Settings,
  FileText,
  Container,
  Boxes,
  User,
  Key,
  Globe,
  X
} from 'lucide-react'

interface RegistryConfig {
  id: number
  name: string
  registry_url: string
  username: string
  password: string
  is_default: boolean
  created_at: string
  updated_at: string
}

type ConfigFormat = 'docker' | 'podman' | 'compose'

interface ConfigFormatOption {
  value: ConfigFormat
  label: string
  icon: React.ComponentType<{ className?: string }>
  description: string
}

export function RegistryManager() {
  const [configs, setConfigs] = useState<RegistryConfig[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [editingConfig, setEditingConfig] = useState<RegistryConfig | null>(null)
  const [showAddForm, setShowAddForm] = useState(false)
  const [dragActive, setDragActive] = useState(false)
  const [selectedFormat, setSelectedFormat] = useState<ConfigFormat>('docker')
  const [configFileContent, setConfigFileContent] = useState('')



  const configFormats: ConfigFormatOption[] = [
    {
      value: 'docker',
      label: 'Docker Config',
      icon: Container,
      description: 'Standard Docker registry configuration (config.json)'
    },
    {
      value: 'podman',
      label: 'Podman Config',
      icon: Boxes,
      description: 'Podman registry configuration with base64 auth'
    },
    {
      value: 'compose',
      label: 'Compose Config',
      icon: User,
      description: 'Docker Compose registry authentication format'
    }
  ]

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    registry_url: '',
    username: '',
    password: '',
    is_default: false
  })

  useEffect(() => {
    fetchConfigs()
  }, [])

  // Config parsing functions
  const parseDockerConfig = (configContent: string) => {
    try {
      const config = JSON.parse(configContent)
      const registries: Array<{ name: string; registry_url: string; username: string; password: string }> = []
      
      if (config.auths) {
        Object.entries(config.auths).forEach(([registry, auth]: [string, any]) => {
          if (auth.auth) {
            const decoded = atob(auth.auth)
            const [username, password] = decoded.split(':')
            registries.push({
              name: registry.replace('https://', '').replace('http://', ''),
              registry_url: registry,
              username: username || '',
              password: password || ''
            })
          }
        })
      }
      return registries
    } catch (err) {
      throw new Error('Invalid Docker config format')
    }
  }

  const parsePodmanConfig = (configContent: string) => {
    try {
      const config = JSON.parse(configContent)
      const registries: Array<{ name: string; registry_url: string; username: string; password: string }> = []
      
      if (config.registries) {
        Object.entries(config.registries).forEach(([registry, data]: [string, any]) => {
          if (data.auth) {
            const decoded = atob(data.auth)
            const [username, password] = decoded.split(':')
            registries.push({
              name: registry,
              registry_url: `https://${registry}`,
              username: username || '',
              password: password || ''
            })
          }
        })
      }
      return registries
    } catch (err) {
      throw new Error('Invalid Podman config format')
    }
  }

  const parseComposeConfig = (configContent: string) => {
    try {
      const config = JSON.parse(configContent)
      const registries: Array<{ name: string; registry_url: string; username: string; password: string }> = []
      
      // Docker Compose format
      if (config.registries && Array.isArray(config.registries)) {
        config.registries.forEach((reg: any) => {
          registries.push({
            name: reg.name || reg.registry || 'unknown',
            registry_url: reg.url || reg.registry_url || reg.endpoint,
            username: reg.username || reg.user || '',
            password: reg.password || reg.token || reg.auth || ''
          })
        })
      } else if (config.registry) {
        // Single registry format
        registries.push({
          name: config.registry.name || 'custom-registry',
          registry_url: config.registry.url || config.registry.endpoint,
          username: config.registry.username || config.registry.user || '',
          password: config.registry.password || config.registry.token || ''
        })
      }
      return registries
    } catch (err) {
      throw new Error('Invalid compose config format')
    }
  }

  const parseConfigFile = (content: string, format: ConfigFormat) => {
    switch (format) {
      case 'docker':
        return parseDockerConfig(content)
      case 'podman':
        return parsePodmanConfig(content)
      case 'compose':
        return parseComposeConfig(content)
      default:
        throw new Error('Unsupported config format')
    }
  }

  const fetchConfigs = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await fetch('/api/registry-configs')
      if (response.ok) {
        const data = await response.json()
        setConfigs(data)
      } else {
        setError('Failed to fetch registry configurations')
      }
    } catch (err) {
      setError('Network error while fetching configurations')
    } finally {
      setLoading(false)
    }
  }

  const createConfig = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await fetch('/api/registry-configs', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      })
      
      if (response.ok) {
        await fetchConfigs()
        setShowAddForm(false)
        resetForm()
      } else {
        const errorData = await response.json()
        setError(errorData.error || 'Failed to create configuration')
      }
    } catch (err) {
      setError('Network error while creating configuration')
    } finally {
      setLoading(false)
    }
  }

  const updateConfig = async () => {
    if (!editingConfig) return
    
    setLoading(true)
    setError(null)
    
    try {
      const response = await fetch(`/api/registry-configs/${editingConfig.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      })
      
      if (response.ok) {
        await fetchConfigs()
        setEditingConfig(null)
        resetForm()
      } else {
        const errorData = await response.json()
        setError(errorData.error || 'Failed to update configuration')
      }
    } catch (err) {
      setError('Network error while updating configuration')
    } finally {
      setLoading(false)
    }
  }

  const deleteConfig = async (id: number) => {
    if (!confirm('Are you sure you want to delete this registry configuration?')) return
    
    setLoading(true)
    setError(null)
    
    try {
      const response = await fetch(`/api/registry-configs/${id}`, {
        method: 'DELETE'
      })
      
      if (response.ok) {
        await fetchConfigs()
      } else {
        setError('Failed to delete configuration')
      }
    } catch (err) {
      setError('Network error while deleting configuration')
    } finally {
      setLoading(false)
    }
  }

  const setDefault = async (id: number) => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await fetch(`/api/registry-configs/${id}/set-default`, {
        method: 'POST'
      })
      
      if (response.ok) {
        await fetchConfigs()
      } else {
        setError('Failed to set default registry')
      }
    } catch (err) {
      setError('Network error while setting default')
    } finally {
      setLoading(false)
    }
  }

  const resetForm = () => {
    setFormData({
      name: '',
      registry_url: '',
      username: '',
      password: '',
      is_default: false
    })
    setConfigFileContent('')
  }

  const startEdit = (config: RegistryConfig) => {
    setEditingConfig(config)
    setFormData({
      name: config.name,
      registry_url: config.registry_url,
      username: config.username,
      password: config.password,
      is_default: config.is_default
    })
    setShowAddForm(true)
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDragActive(false)
    
    const files = Array.from(e.dataTransfer.files)
    files.forEach(file => {
      if (file.type === 'application/json' || file.name.endsWith('.json')) {
        const reader = new FileReader()
        reader.onload = (event) => {
          const content = event.target?.result as string
          setConfigFileContent(content)
          try {
            const parsedConfigs = parseConfigFile(content, selectedFormat)
            if (parsedConfigs.length > 0) {
              const firstConfig = parsedConfigs[0]
              setFormData({
                name: firstConfig.name,
                registry_url: firstConfig.registry_url,
                username: firstConfig.username,
                password: firstConfig.password,
                is_default: false
              })
            }
          } catch (err) {
            setError((err as Error).message)
          }
        }
        reader.readAsText(file)
      }
    })
  }

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (event) => {
        const content = event.target?.result as string
        setConfigFileContent(content)
        try {
          const parsedConfigs = parseConfigFile(content, selectedFormat)
          if (parsedConfigs.length > 0) {
            const firstConfig = parsedConfigs[0]
            setFormData({
              name: firstConfig.name,
              registry_url: firstConfig.registry_url,
              username: firstConfig.username,
              password: firstConfig.password,
              is_default: false
            })
          }
        } catch (err) {
          setError((err as Error).message)
        }
      }
      reader.readAsText(file)
    }
  }

  return (
    <div className="space-y-6">
      <div className="text-center py-8">
        <Settings className="h-12 w-12 mx-auto mb-4 text-primary" />
        <h2 className="text-2xl font-bold mb-2">Registry Configuration</h2>
        <p className="text-muted-foreground max-w-2xl mx-auto">
          Manage your container registry configurations for accessing private Helm charts.
          Support for Docker, Podman, and Docker Compose configuration formats.
        </p>
      </div>

      {error && (
        <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md flex items-center justify-between">
          <p className="text-sm text-destructive">{error}</p>
          <Button variant="ghost" size="sm" onClick={() => setError(null)}>
            <X className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Add Registry Form */}
      {!showAddForm ? (
        <Card className="border-2 border-dashed border-primary/20 bg-primary/5">
          <CardContent className="p-6 text-center">
            <Plus className="h-8 w-8 mx-auto mb-4 text-primary" />
            <h3 className="font-semibold mb-2">Add Registry Configuration</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Create a new registry configuration manually or upload a config file
            </p>
            <Button onClick={() => setShowAddForm(true)}>
              <Plus className="h-4 w-4 mr-2" />
              Add Registry
            </Button>
          </CardContent>
        </Card>
      ) : (
        <Card className="border-primary/20">
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span className="flex items-center gap-2">
                <Plus className="h-5 w-5" />
                {editingConfig ? 'Edit Registry Configuration' : 'Add Registry Configuration'}
              </span>
              <Button 
                variant="ghost" 
                size="sm" 
                onClick={() => {
                  setShowAddForm(false)
                  setEditingConfig(null)
                  resetForm()
                }}
              >
                <X className="h-4 w-4" />
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Config Format Selection */}
            <div className="space-y-3">
              <Label>Configuration Format</Label>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                {configFormats.map((format) => {
                  const Icon = format.icon
                  return (
                    <Card 
                      key={format.value}
                      className={`cursor-pointer transition-colors ${
                        selectedFormat === format.value 
                          ? 'border-primary bg-primary/5' 
                          : 'border-muted hover:border-primary/50'
                      }`}
                      onClick={() => setSelectedFormat(format.value)}
                    >
                      <CardContent className="p-4 text-center">
                        <Icon className="h-6 w-6 mx-auto mb-2" />
                        <h4 className="font-medium text-sm">{format.label}</h4>
                        <p className="text-xs text-muted-foreground mt-1">{format.description}</p>
                      </CardContent>
                    </Card>
                  )
                })}
              </div>
            </div>

            {/* File Upload Area */}
            <div 
              className={`border-2 border-dashed rounded-lg p-6 text-center transition-colors ${
                dragActive ? 'border-primary bg-primary/5' : 'border-muted-foreground/25'
              }`}
              onDrop={handleDrop}
              onDragOver={(e) => { e.preventDefault(); setDragActive(true) }}
              onDragEnter={(e) => { e.preventDefault(); setDragActive(true) }}
              onDragLeave={(e) => { e.preventDefault(); setDragActive(false) }}
            >
              <Upload className="h-6 w-6 mx-auto mb-2 text-muted-foreground" />
              <p className="text-sm font-medium mb-1">Upload {configFormats.find(f => f.value === selectedFormat)?.label}</p>
              <p className="text-xs text-muted-foreground mb-3">
                Drag and drop your config file here, or click to browse
              </p>
              <input
                type="file"
                accept=".json"
                onChange={handleFileUpload}
                className="hidden"
                id="config-upload"
              />
              <Button variant="outline" size="sm" onClick={() => document.getElementById('config-upload')?.click()}>
                <FileText className="h-4 w-4 mr-2" />
                Browse Files
              </Button>
            </div>

            {/* Manual Configuration Form */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="name">Registry Name</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="docker-hub"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="registry_url">Registry URL</Label>
                <Input
                  id="registry_url"
                  value={formData.registry_url}
                  onChange={(e) => setFormData({ ...formData, registry_url: e.target.value })}
                  placeholder="https://registry-1.docker.io"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="username">Username</Label>
                <Input
                  id="username"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  placeholder="your-username"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password/Token</Label>
                <Input
                  id="password"
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  placeholder="your-password-or-token"
                />
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <Switch
                id="is_default"
                checked={formData.is_default}
                onCheckedChange={(checked) => setFormData({ ...formData, is_default: checked })}
              />
              <Label htmlFor="is_default">Set as default registry</Label>
            </div>

            {configFileContent && (
              <div className="space-y-2">
                <Label>Uploaded Configuration</Label>
                <Textarea
                  value={configFileContent}
                  onChange={(e) => setConfigFileContent(e.target.value)}
                  rows={4}
                  className="font-mono text-xs"
                  placeholder="Configuration file content will appear here..."
                />
              </div>
            )}

            <div className="flex gap-2">
              <Button 
                onClick={editingConfig ? updateConfig : createConfig}
                disabled={loading || !formData.name || !formData.registry_url}
                className="flex-1"
              >
                {loading ? 'Saving...' : (editingConfig ? 'Update Registry' : 'Create Registry')}
              </Button>
              <Button 
                variant="outline" 
                onClick={() => {
                  setShowAddForm(false)
                  setEditingConfig(null)
                  resetForm()
                }}
              >
                Cancel
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Registry List */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold">Configured Registries</h3>
        
        {configs.map((config) => (
          <Card key={config.id} className={config.is_default ? 'border-primary' : ''}>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    <Globe className="h-4 w-4 text-muted-foreground" />
                    <h4 className="font-semibold">{config.name}</h4>
                    {config.is_default && (
                      <Badge variant="default" className="flex items-center gap-1">
                        <Star className="h-3 w-3" />
                        Default
                      </Badge>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground mb-1">{config.registry_url}</p>
                  <div className="flex items-center gap-4 text-xs text-muted-foreground">
                    <span className="flex items-center gap-1">
                      <User className="h-3 w-3" />
                      {config.username || 'No username'}
                    </span>
                    <span className="flex items-center gap-1">
                      <Key className="h-3 w-3" />
                      {config.password ? 'Password set' : 'No password'}
                    </span>
                    <span>Created: {new Date(config.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {!config.is_default && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setDefault(config.id)}
                      disabled={loading}
                      title="Set as default"
                    >
                      <Star className="h-4 w-4" />
                    </Button>
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => startEdit(config)}
                    title="Edit registry"
                  >
                    <Edit className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => deleteConfig(config.id)}
                    disabled={loading}
                    title="Delete registry"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
        
        {configs.length === 0 && !loading && (
          <div className="text-center py-8 text-muted-foreground">
            <Database className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">No registry configurations found</p>
            <p className="text-xs">Add a registry configuration to get started</p>
          </div>
        )}
      </div>
    </div>
  )
}