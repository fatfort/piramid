import { useState, useEffect } from 'react'
import { 
  Activity, 
  AlertTriangle, 
  Globe, 
  TrendingUp,
  Clock,
  Ban
} from 'lucide-react'
import WorldMap from '../components/WorldMap'
import EventsChart from '../components/EventsChart'
import RecentEvents from '../components/RecentEvents'

interface Stats {
  total_events: number
  unique_ips: number
  banned_ips: number
  recent_events: number
  top_countries: Array<{ country: string; count: number }>
  last_updated: number
}

export default function DashboardPage() {
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    fetchStats()
    // Refresh stats every 30 seconds
    const interval = setInterval(fetchStats, 30000)
    return () => clearInterval(interval)
  }, [])

  const fetchStats = async () => {
    try {
      const response = await fetch('/api/stats/overview', {
        credentials: 'include',
      })

      if (!response.ok) {
        throw new Error('Failed to fetch stats')
      }

      const data = await response.json()
      if (data.success) {
        setStats(data.data)
      } else {
        throw new Error(data.error || 'Failed to fetch stats')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch stats')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="p-6">
        <div className="animate-pulse space-y-6">
          <div className="h-8 bg-gray-200 rounded w-64"></div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="card p-6">
                <div className="h-6 bg-gray-200 rounded w-32 mb-2"></div>
                <div className="h-8 bg-gray-200 rounded w-20"></div>
              </div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
          {error}
        </div>
      </div>
    )
  }

  const statCards = [
    {
      name: 'Total Events',
      value: stats?.total_events || 0,
      icon: Activity,
      color: 'text-blue-600',
      bgColor: 'bg-blue-100',
    },
    {
      name: 'Unique IPs',
      value: stats?.unique_ips || 0,
      icon: Globe,
      color: 'text-green-600',
      bgColor: 'bg-green-100',
    },
    {
      name: 'Recent Events',
      value: stats?.recent_events || 0,
      icon: Clock,
      color: 'text-yellow-600',
      bgColor: 'bg-yellow-100',
    },
    {
      name: 'Banned IPs',
      value: stats?.banned_ips || 0,
      icon: Ban,
      color: 'text-red-600',
      bgColor: 'bg-red-100',
    },
  ]

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Security Dashboard</h1>
          <p className="text-gray-600">Real-time monitoring and threat detection</p>
        </div>
        <div className="flex items-center text-sm text-gray-500">
          <Clock className="h-4 w-4 mr-1" />
          Last updated: {stats?.last_updated ? new Date(stats.last_updated * 1000).toLocaleTimeString() : 'Never'}
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat) => (
          <div key={stat.name} className="card p-6">
            <div className="flex items-center">
              <div className={`p-3 rounded-lg ${stat.bgColor}`}>
                <stat.icon className={`h-6 w-6 ${stat.color}`} />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">{stat.name}</p>
                <p className="text-2xl font-bold text-gray-900">
                  {stat.value.toLocaleString()}
                </p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Charts and Map */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* World Map */}
        <div className="card p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Threat Origins</h2>
            <Globe className="h-5 w-5 text-gray-400" />
          </div>
          <WorldMap />
        </div>

        {/* Events Chart */}
        <div className="card p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Events Over Time</h2>
            <TrendingUp className="h-5 w-5 text-gray-400" />
          </div>
          <EventsChart />
        </div>
      </div>

      {/* Recent Events */}
      <div className="card">
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Recent Security Events</h2>
            <AlertTriangle className="h-5 w-5 text-gray-400" />
          </div>
        </div>
        <RecentEvents />
      </div>
    </div>
  )
} 