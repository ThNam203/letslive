import { useState } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Copy, RefreshCw } from 'lucide-react'
import { User } from '@/types/user'
import { toast } from 'react-toastify'

export default function ApiKeyTab({user, updateUser} : {user: User | undefined, updateUser: (user: User) => void}) {
  const [apiKey, setApiKey] = useState(user ? user.streamAPIKey : "")

  const generateNewApiKey = () => {
    if (!user) return

    const newApiKey = 'yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy'
    updateUser({...user, streamAPIKey: newApiKey})
    setApiKey(newApiKey)
  }

  const copyApiKey = () => {
    navigator.clipboard.writeText(apiKey)
    toast.success('API Key copied to clipboard')
  }

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="api-key">Your API Key</Label>
        <div className="flex">
          <Input
            id="api-key"
            value={apiKey}
            readOnly
            className="flex-grow"
          />
          <Button onClick={copyApiKey} className="ml-2">
            <Copy className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <Button onClick={generateNewApiKey}>
        <RefreshCw className="mr-2 h-4 w-4" /> Generate New API Key
      </Button>
    </div>
  )
}

