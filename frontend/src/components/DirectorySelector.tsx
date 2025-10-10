import { useState } from 'react'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Card, CardContent } from './ui/card'
import { FolderOpen, RefreshCw, Search } from 'lucide-react'

interface DirectorySelectorProps {
  directory: string
  onDirectoryChange: (dir: string) => void
  onRefresh: () => void
  onRefreshStored: () => void
  loading: boolean
}

export function DirectorySelector({ 
  directory, 
  onDirectoryChange, 
  onRefresh, 
  onRefreshStored,
  loading 
}: DirectorySelectorProps) {
  const [inputValue, setInputValue] = useState(directory)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onDirectoryChange(inputValue)
  }

  return (
    <Card>
      <CardContent className="pt-6">
        <form onSubmit={handleSubmit} className="flex gap-2">
          <div className="flex-1 relative">
            <FolderOpen className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              placeholder="Enter directory path (e.g., ./charts, /path/to/helm/charts)"
              className="pl-10"
            />
          </div>
          <Button type="submit" disabled={loading}>
            <Search className="h-4 w-4 mr-2" />
            Scan Directory
          </Button>
          <Button 
            type="button" 
            variant="outline" 
            onClick={onRefreshStored}
            disabled={loading}
            title="Refresh stored charts from database"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            Stored
          </Button>
        </form>
        <div className="flex justify-between items-center mt-2">
          <p className="text-sm text-muted-foreground">
            Directory: <code className="bg-muted px-1 py-0.5 rounded">{directory}</code>
          </p>
          <p className="text-sm text-muted-foreground">
            Charts are automatically saved to database when fetched
          </p>
        </div>
      </CardContent>
    </Card>
  )
}