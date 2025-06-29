import { AlertTriangle, Shield, Globe, Clock } from 'lucide-react'

interface Event {
  id: number
  type: string
  message: string
  timestamp: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  sourceIP: string
}

export default function RecentEvents() {
  // Mock recent events data
  const recentEvents: Event[] = [
    {
      id: 1,
      type: 'SSH Brute Force',
      message: 'Multiple failed login attempts detected',
      timestamp: new Date(Date.now() - 300000).toISOString(),
      severity: 'high',
      sourceIP: '192.168.1.100'
    },
    {
      id: 2, 
      type: 'Port Scan',
      message: 'Suspicious port scanning activity',
      timestamp: new Date(Date.now() - 600000).toISOString(),
      severity: 'medium',
      sourceIP: '10.0.0.25'
    },
    {
      id: 3,
      type: 'Malware Detection',
      message: 'Malicious payload identified',
      timestamp: new Date(Date.now() - 900000).toISOString(),
      severity: 'critical',
      sourceIP: '172.16.0.50'
    },
    {
      id: 4,
      type: 'DDoS Attempt',
      message: 'High volume traffic from single source',
      timestamp: new Date(Date.now() - 1200000).toISOString(),
      severity: 'high',
      sourceIP: '203.0.113.10'
    },
    {
      id: 5,
      type: 'Unauthorized Access',
      message: 'Login from unusual location',
      timestamp: new Date(Date.now() - 1500000).toISOString(),
      severity: 'medium',
      sourceIP: '198.51.100.20'
    }
  ]

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
        return <AlertTriangle className="w-4 h-4 text-red-600" />
      case 'high':
        return <Shield className="w-4 h-4 text-orange-600" />
      case 'medium':
        return <Globe className="w-4 h-4 text-yellow-600" />
      default:
        return <Clock className="w-4 h-4 text-blue-600" />
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-50 border-red-200'
      case 'high':
        return 'bg-orange-50 border-orange-200'
      case 'medium':
        return 'bg-yellow-50 border-yellow-200'
      default:
        return 'bg-blue-50 border-blue-200'
    }
  }

  const getTimeAgo = (timestamp: string) => {
    const now = new Date()
    const eventTime = new Date(timestamp)
    const diffInMinutes = Math.floor((now.getTime() - eventTime.getTime()) / (1000 * 60))
    
    if (diffInMinutes < 1) return 'Just now'
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`
    
    const hours = Math.floor(diffInMinutes / 60)
    if (hours < 24) return `${hours}h ago`
    
    const days = Math.floor(hours / 24)
    return `${days}d ago`
  }

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-medium text-gray-900">Recent Events</h3>
        <button className="text-sm text-blue-600 hover:text-blue-800">View All</button>
      </div>
      
      <div className="space-y-3">
        {recentEvents.map((event) => (
          <div 
            key={event.id} 
            className={`p-3 rounded-lg border ${getSeverityColor(event.severity)} transition-colors hover:shadow-sm`}
          >
            <div className="flex items-start justify-between">
              <div className="flex items-start space-x-3 flex-1">
                <div className="flex-shrink-0 mt-1">
                  {getSeverityIcon(event.severity)}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center space-x-2">
                    <span className="text-sm font-medium text-gray-900">{event.type}</span>
                    <span className="text-xs text-gray-500">from {event.sourceIP}</span>
                  </div>
                  <p className="text-sm text-gray-600 mt-1">{event.message}</p>
                </div>
              </div>
              <div className="flex-shrink-0 text-xs text-gray-500">
                {getTimeAgo(event.timestamp)}
              </div>
            </div>
          </div>
        ))}
      </div>
      
      {recentEvents.length === 0 && (
        <div className="text-center py-8">
          <Shield className="mx-auto h-8 w-8 text-gray-400" />
          <p className="mt-2 text-sm text-gray-500">No recent events</p>
        </div>
      )}
    </div>
  )
} 