import { useState, useEffect } from 'react'
import { Shield, Trash2, Plus, Globe } from 'lucide-react'

interface BannedIP {
  id: number
  ipAddress: string
  reason: string
  bannedAt: string
  expiresAt?: string
  country?: string
  city?: string
  permanent: boolean
}

export default function BansPage() {
  const [bannedIPs, setBannedIPs] = useState<BannedIP[]>([])
  const [loading, setLoading] = useState(true)
  const [showAddForm, setShowAddForm] = useState(false)
  const [newIP, setNewIP] = useState('')
  const [newReason, setNewReason] = useState('')
  const [isPermanent, setIsPermanent] = useState(true)

  useEffect(() => {
    // Mock data for now - in production this would fetch from API
    const mockBans: BannedIP[] = [
      {
        id: 1,
        ipAddress: '192.168.1.100',
        reason: 'SSH brute force attack',
        bannedAt: new Date(Date.now() - 86400000).toISOString(),
        country: 'US',
        city: 'New York',
        permanent: false,
        expiresAt: new Date(Date.now() + 86400000).toISOString()
      },
      {
        id: 2,
        ipAddress: '10.0.0.50',
        reason: 'Multiple failed login attempts',
        bannedAt: new Date(Date.now() - 172800000).toISOString(),
        country: 'CN',
        city: 'Beijing',
        permanent: true
      },
      {
        id: 3,
        ipAddress: '172.16.0.25',
        reason: 'Malware detected',
        bannedAt: new Date(Date.now() - 259200000).toISOString(),
        country: 'RU',
        city: 'Moscow',
        permanent: true
      }
    ]
    
    setTimeout(() => {
      setBannedIPs(mockBans)
      setLoading(false)
    }, 1000)
  }, [])

  const handleAddBan = () => {
    if (!newIP.trim() || !newReason.trim()) return

    const newBan: BannedIP = {
      id: Date.now(),
      ipAddress: newIP.trim(),
      reason: newReason.trim(),
      bannedAt: new Date().toISOString(),
      permanent: isPermanent,
      expiresAt: isPermanent ? undefined : new Date(Date.now() + 86400000).toISOString()
    }

    setBannedIPs([newBan, ...bannedIPs])
    setNewIP('')
    setNewReason('')
    setShowAddForm(false)
  }

  const handleRemoveBan = (id: number) => {
    setBannedIPs(bannedIPs.filter(ban => ban.id !== id))
  }

  const isExpired = (ban: BannedIP) => {
    if (ban.permanent || !ban.expiresAt) return false
    return new Date(ban.expiresAt) < new Date()
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">IP Bans</h1>
          <p className="text-gray-600 mt-1">Manage blocked IP addresses</p>
        </div>
        <button
          onClick={() => setShowAddForm(true)}
          className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
        >
          <Plus className="w-4 h-4" />
          Ban IP
        </button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-3 rounded-full bg-red-100">
              <Shield className="w-6 h-6 text-red-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Total Bans</p>
              <p className="text-2xl font-bold text-gray-900">{bannedIPs.length}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-3 rounded-full bg-orange-100">
              <Globe className="w-6 h-6 text-orange-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Permanent</p>
              <p className="text-2xl font-bold text-gray-900">
                {bannedIPs.filter(ban => ban.permanent).length}
              </p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-3 rounded-full bg-green-100">
              <Shield className="w-6 h-6 text-green-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Temporary</p>
              <p className="text-2xl font-bold text-gray-900">
                {bannedIPs.filter(ban => !ban.permanent).length}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Add Ban Form */}
      {showAddForm && (
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Add IP Ban</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                IP Address
              </label>
              <input
                type="text"
                placeholder="192.168.1.100"
                value={newIP}
                onChange={(e) => setNewIP(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Reason
              </label>
              <input
                type="text"
                placeholder="SSH brute force attack"
                value={newReason}
                onChange={(e) => setNewReason(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
            <div className="flex items-center">
              <input
                type="checkbox"
                id="permanent"
                checked={isPermanent}
                onChange={(e) => setIsPermanent(e.target.checked)}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label htmlFor="permanent" className="ml-2 block text-sm text-gray-900">
                Permanent ban
              </label>
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleAddBan}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
              >
                Add Ban
              </button>
              <button
                onClick={() => setShowAddForm(false)}
                className="px-4 py-2 bg-gray-300 text-gray-700 rounded-lg hover:bg-gray-400"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Banned IPs Table */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Banned IP Addresses</h3>
        </div>
        
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  IP Address
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Reason
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Location
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Banned At
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {bannedIPs.map((ban) => (
                <tr key={ban.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                    {ban.ipAddress}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-900 max-w-xs truncate">
                    {ban.reason}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {ban.country && ban.city ? `${ban.city}, ${ban.country}` : 'Unknown'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {new Date(ban.bannedAt).toLocaleString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {ban.permanent ? (
                      <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">
                        Permanent
                      </span>
                    ) : isExpired(ban) ? (
                      <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800">
                        Expired
                      </span>
                    ) : (
                      <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-yellow-100 text-yellow-800">
                        Temporary
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    <button
                      onClick={() => handleRemoveBan(ban.id)}
                      className="text-red-600 hover:text-red-900"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {bannedIPs.length === 0 && (
          <div className="text-center py-12">
            <Shield className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">No banned IPs</h3>
            <p className="mt-1 text-sm text-gray-500">Get started by banning your first IP address.</p>
          </div>
        )}
      </div>
    </div>
  )
} 