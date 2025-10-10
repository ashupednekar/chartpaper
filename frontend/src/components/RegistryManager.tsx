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
  Shield,
  CheckCircle,
  AlertCircle,
  Settings
} from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './ui/dialog'

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

export function RegistryManager() {
  const [configs, setConfigs] = useState<RegistryConfig[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [editingConfig, setEditingConfig] = useState<RegistryConfig | null>(null)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [dragActive, setDragActive] = useState(false)

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
        setShowCreateDialog(false)
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
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDragActive(false)
    
    const files = Array.from(e.dataTransfer.files)
    files.forEach(file => {
      if (file.type === 'application/json' || file.name.endsWith('.json')) {
        const reader = new FileReader()
        reader.onload = (event) => {
          try {
            const config = JSON.parse(event.target?.result as string)
            // Parse Docker config format
            if (config.auths) {
              Object.entries(config.auths).forEach(([registry, auth]: [string, any]) => {
                const decoded = atob(auth.auth || '')
                const [username, password] = decoded.split(':')
                
                setFormData({
                  name: registry.replace('https://', '').replace('http://', ''),
                  registry_url: registry,
                  username: username || '',
                  password: password || '',
                  is_default: false
                })
                setShowCreateDialog(true)
              })
            }
          } catch (err) {
            setError('Invalid configuration file format')
          }
        }
        reader.readAsText(file)
      }
    })
  }

  return (
    <div className="space-y-6">
      <div className="text-center py-8">
        <Settings className="h-12 w-12 mx-auto mb-4 text-primary" />
        <h2 className="text-2xl font-bold mb-2">Registry Configuration</h2>
        <p className="text-muted-foreground max-w-2xl mx-auto">
          Manage your container registry configurations for accessing private Helm charts.
          Upload Docker config files or create configurations manually.
        </p>
      </div>

      {error && (
        <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      {/* Upload Area */}
      <Card 
        className={`border-2 border-dashed transition-colors ${
          dragActive ? 'border-primary bg-primary/5' : 'border-muted-foreground/25'
        }`}
        onDrop={handleDrop}
        onDragOver={(e) => { e.preventDefault(); setDragActive(true) }}
        onDragLeave={() => setDragActive(false)}
      >
        <CardContent className="p-8 text-center">
          <Upload className="h-8 w-8 mx-auto mb-4 text-muted-foreground" />
          <h3 className="font-semibold mb-2">Upload Registry Configuration</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Drag and drop Docker config.json files or click to browse
          </p>
          <Button variant="outline">
            <Upload className="h-4 w-4 mr-2" />
            Browse Files
          </Button>
        </CardContent>
      </Card>

      {/* Registry Configurations */}
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold">Registry Configurations</h3>
        <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
          <DialogTrigger asChild>
            <Button onClick={() => { resetForm(); setShowCreateDialog(true) }}>
              <Plus className="h-4 w-4 mr-2" />
              Add Registry
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                {editingConfig ? 'Edit Registry Configuration' : 'Add Registry Configuration'}
              </DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div>
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="docker-hub"
                />
              </div>
              <div>
                <Label htmlFor="registry_url">Registry URL</Label>
                <Input
                  id="registry_url"
                  value={formData.registry_url}
                  onChange={(e) => setFormData({ ...formData, registry_url: e.target.value })}
                  placeholder="https://registry-1.docker.io"
                />
              </div>
              <div>
                <Label htmlFor="username">Username</Label>
                <Input
                  id="username"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  placeholder="your-username"
                />
              </div>
              <div>
                <Label htmlFor="password">Password/Token</Label>
                <Input
                  id="password"
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  placeholder="your-password-or-token"
                />
              </div>
              <div className="flex items-center space-x-2">
                <Switch
                  id="is_default"
                  checked={formData.is_default}
                  onCheckedChange={(checked) => setFormData({ ...formData, is_default: checked })}
                />
                <Label htmlFor="is_default">Set as default registry</Label>
              </div>
              <div className="flex gap-2">
                <Button 
                  onClick={editingConfig ? updateConfig : createConfig}
                  disabled={loading}
                  className="flex-1"
                >
                  {loading ? 'Saving...' : (editingConfig ? 'Update' : 'Create')}
                </Button>
                <Button 
                  variant="outline" 
                  onClick={() => {
                    setShowCreateDialog(false)
                    setEditingConfig(null)
                    resetForm()
                  }}
                >
                  Cancel
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {/* Registry List */}
      <div className="grid gap-4">
        {configs.map((config) => (
          <Card key={config.id} className={config.is_default ? 'border-primary' : ''}>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    <h4 className="font-semibold">{config.name}</h4>
                    {config.is_default && (
                      <Badge variant="default" className="flex items-center gap-1">
                        <Star className="h-3 w-3" />
                        Default
                      </Badge>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground mb-1">{config.registry_url}</p>
                  <p className="text-xs text-muted-foreground">
                    Username: {config.username || 'None'} • 
                    Created: {new Date(config.created_at).toLocaleDateString()}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  {!config.is_default && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setDefault(config.id)}
                      disabled={loading}
                    >
                      <Star className="h-4 w-4" />
                    </Button>
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => startEdit(config)}
                  >
                    <Edit className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => deleteConfig(config.id)}
                    disabled={loading}
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

      {/* Edit Dialog */}
      <Dialog open={!!editingConfig} onOpenChange={(open: boolean) => !open && setEditingConfig(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Registry Configuration</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label htmlFor="edit-name">Name</Label>
              <Input
                id="edit-name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="docker-hub"
              />
            </div>
            <div>
              <Label htmlFor="edit-registry_url">Registry URL</Label>
              <Input
                id="edit-registry_url"
                value={formData.registry_url}
                onChange={(e) => setFormData({ ...formData, registry_url: e.target.value })}
                placeholder="https://registry-1.docker.io"
              />
            </div>
            <div>
              <Label htmlFor="edit-username">Username</Label>
              <Input
                id="edit-username"
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                placeholder="your-username"
              />
            </div>
            <div>
              <Label htmlFor="edit-password">Password/Token</Label>
              <Input
                id="edit-password"
                type="password"
                value={formData.password}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                placeholder="your-password-or-token"
              />
            </div>
            <div className="flex items-center space-x-2">
              <Switch
                id="edit-is_default"
                checked={formData.is_default}
                onCheckedChange={(checked) => setFormData({ ...formData, is_default: checked })}
              />
              <Label htmlFor="edit-is_default">Set as default registry</Label>
            </div>
            <div className="flex gap-2">
              <Button 
                onClick={updateConfig}
                disabled={loading}
                className="flex-1"
              >
                {loading ? 'Updating...' : 'Update'}
              </Button>
              <Button 
                variant="outline" 
                onClick={() => {
                  setEditingConfig(null)
                  resetForm()
                }}
              >
                Cancel
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}