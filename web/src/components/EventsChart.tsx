export default function EventsChart() {
  // Mock data for the chart
  const data = Array.from({ length: 24 }, (_, i) => ({
    hour: `${i}:00`,
    events: Math.floor(Math.random() * 100) + 10,
    alerts: Math.floor(Math.random() * 20) + 2
  }))

  const maxEvents = Math.max(...data.map(d => d.events))

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h3 className="text-lg font-medium text-gray-900 mb-4">24-Hour Activity</h3>
      <div className="h-64">
        <div className="flex items-end justify-between h-full space-x-1">
          {data.map((item, index) => (
            <div key={index} className="flex-1 flex flex-col items-center">
              <div className="w-full flex flex-col justify-end h-48">
                <div 
                  className="bg-blue-500 rounded-t-sm transition-all duration-300 hover:bg-blue-600"
                  style={{ height: `${(item.events / maxEvents) * 100}%` }}
                  title={`${item.hour}: ${item.events} events`}
                ></div>
              </div>
              {index % 4 === 0 && (
                <span className="text-xs text-gray-500 mt-1">{item.hour}</span>
              )}
            </div>
          ))}
        </div>
      </div>
      <div className="mt-4 flex justify-center space-x-6 text-sm">
        <div className="flex items-center">
          <div className="w-3 h-3 bg-blue-500 rounded mr-2"></div>
          <span className="text-gray-600">Total Events</span>
        </div>
        <div className="flex items-center">
          <div className="w-3 h-3 bg-red-500 rounded mr-2"></div>
          <span className="text-gray-600">Critical Alerts</span>
        </div>
      </div>
    </div>
  )
} 