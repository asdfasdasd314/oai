function SyncItem({ time }) {
    return (
        <div key={time.id} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
            <div className="space-y-1">
                <span className="block font-medium text-gray-900">{time.label}</span>
                <span className="block text-sm text-gray-600">Starting: {formatDate(time.date)}</span>
                <span className="block text-sm text-gray-600">
                    <span className="inline-block mr-1">↻</span>
                    {formatRecurrence(time.recurrenceInterval)}
                </span>
            </div>
            <button 
                className="p-2 text-gray-400 hover:text-red-500 transition-colors"
                onClick={() => removeTime(time.id)}
                aria-label="Remove"
            >
                ✕
            </button>
        </div>
    );
}
    
export default SyncItem;