"use client"

import { useState, useEffect } from "react"
import TimePicker from "./TimePicker"

// Define the SyncTime type for individual sync entries with recurrence
export interface SyncTime {
    id: string
    date: Date
    label: string
    recurrenceInterval: number // Days between syncs
}

// Props for the SyncList component
interface SyncListProps {
    initialTimes?: SyncTime[]
    onSync?: (times: SyncTime[]) => Promise<void>
    title?: string
}

export default function SyncList({ initialTimes = [], onSync = async () => { } }: SyncListProps) {
    const title = "Sync Times";

    const [times, setTimes] = useState<SyncTime[]>(initialTimes);
    const [newDate, setNewDate] = useState<string>((new Date()).toISOString().split('T')[0]);
    const [newTime, setNewTime] = useState<string>((new Date()).toTimeString().slice(0, 5));
    const [newLabel, setNewLabel] = useState("");
    const [recurrenceInterval, setRecurrenceInterval] = useState(7); // Default to weekly
    const [isLoading, setIsLoading] = useState(false);
    const [message, setMessage] = useState<{ type: string; text: string } | null>(null);

    // Show a message/toast
    const showMessage = (text: string, type: "success" | "error" = "success") => {
        setMessage({ type, text });
        setTimeout(() => setMessage(null), 3000);
    }

    // Format date for display
    const formatDate = (date: Date): string => {
        return date.toLocaleDateString("en-US", {
            year: "numeric",
            month: "long",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit"
        })
    }

    // Validate if the selected date and time is in the future
    const isValidDateTime = (date: string, time: string): boolean => {
        if (!date || !time) return false;
        
        const selectedDateTime = new Date(`${date}T${time}`);
        const now = new Date();

        return selectedDateTime > now;
    }

    // Check if two sync times will collide
    const willCollide = (time1: SyncTime, time2: SyncTime): boolean => {
        // If they have different recurrence intervals, they won't collide
        if (time1.recurrenceInterval !== time2.recurrenceInterval) {
            return false;
        }

        // Get the time components (hours, minutes, seconds) for both times
        const getTimeComponents = (date: Date) => ({
            hours: date.getHours(),
            minutes: date.getMinutes(),
            seconds: date.getSeconds()
        });

        const t1 = getTimeComponents(time1.date);
        const t2 = getTimeComponents(time2.date);

        // Compare time components
        if (t1.hours === t2.hours && t1.minutes === t2.minutes && t1.seconds === t2.seconds) {
            // Check if one will end up on the other at some point
            
            // Calculate the difference in the starting day
            const dayDiff = Math.abs(
                Math.floor((time1.date.getTime() - time2.date.getTime()) / (1000 * 60 * 60 * 24))
            );
            
            // If the day difference is divisible by either recurrence interval, they will collide
            if (dayDiff % time1.recurrenceInterval === 0) {
                return true;
            }
        }

        return false;
    }

    // Add a new time to the list
    const addTime = () => {
        if (!newDate || !newTime) {
            showMessage("Please select both date and time", "error")
            return
        }

        if (!isValidDateTime(newDate, newTime)) {
            showMessage("Please select a future date and time", "error")
            return
        }

        const dateObj = new Date(`${newDate}T${newTime}`)
        const newTimeEntry: SyncTime = {
            id: Math.random().toString(36).substring(2, 11),
            date: dateObj,
            label: newLabel.trim() || formatDate(dateObj),
            recurrenceInterval: recurrenceInterval,
        }

        // Check for collisions with existing times
        const collision = times.find(existingTime => willCollide(newTimeEntry, existingTime));
        if (collision) {
            showMessage(
                `This sync time conflicts with an existing sync at ${formatDate(collision.date)} (${collision.label})`, 
                "error"
            );
            return;
        }

        setTimes([...times, newTimeEntry])
        setNewLabel("")
        // Keep the recurrence interval as is for the next entry

        showMessage(`Added ${formatDate(newTimeEntry.date)} (every ${recurrenceInterval} days)`)
    }

    // Remove a time from the list
    const removeTime = (id: string) => {
        setTimes(times.filter((time) => time.id !== id))
        showMessage("The sync schedule has been removed from the list")
    }

    // Sync times with the backend
    const handleSync = async () => {
        setIsLoading(true)
        try {
            await onSync(times)
            showMessage(`Synced ${times.length} schedules with the database`)
        } catch (error) {
            showMessage("Failed to sync schedules with the database", "error")
        } finally {
            setIsLoading(false)
        }
    }

    // Format the recurrence interval in a human-readable way
    const formatRecurrence = (days: number) => {
        if (days === 1) return "Daily"
        if (days === 7) return "Weekly"
        if (days === 14) return "Bi-weekly"
        if (days === 30) return "Monthly"
        return `Every ${days} days`
    }

    return (
        <div className="max-w-3xl mx-auto p-5">
            {message && (
                <div className={`mb-4 p-4 rounded-lg ${
                    message.type === "success" 
                        ? "bg-green-100 text-green-700" 
                        : "bg-red-100 text-red-700"
                }`}>
                    {message.text}
                </div>
            )}

            <div className="bg-white rounded-lg shadow-lg overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-200">
                    <h2 className="text-2xl font-semibold text-gray-800">{title}</h2>
                </div>

                <div className="p-6">
                    {/* List of existing times */}
                    {times.length > 0 ? (
                        <div className="space-y-4">
                            {times.map((time) => (
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
                            ))}
                        </div>
                    ) : (
                        <div className="text-center text-gray-500 py-8">
                            No sync schedules added yet. Add your first schedule below.
                        </div>
                    )}

                    {/* Add new time form */}
                    <div className="mt-8 space-y-6">
                        <div className="space-y-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label htmlFor="date" className="block text-sm font-medium text-gray-700">Start Date</label>
                                    <input
                                        type="date"
                                        id="date"
                                        value={newDate}
                                        min={(new Date()).toISOString().split('T')[0]}
                                        onChange={(e) => setNewDate(e.target.value)}
                                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                                    />
                                </div>

                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-1">Start Time</label>
                                    <TimePicker
                                        value={newTime}
                                        onChange={setNewTime}
                                        minTime={newDate === (new Date()).toISOString().split('T')[0] ? (new Date()).toTimeString().slice(0, 5) : undefined}
                                        className="mt-1"
                                    />
                                </div>
                            </div>

                            <div>
                                <label htmlFor="label" className="block text-sm font-medium text-gray-700">Label (optional)</label>
                                <input
                                    type="text"
                                    id="label"
                                    placeholder="Enter a label for this sync schedule"
                                    value={newLabel}
                                    onChange={(e) => setNewLabel(e.target.value)}
                                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                                />
                            </div>

                            <div>
                                <div className="flex justify-between items-center mb-2">
                                    <label htmlFor="recurrence" className="block text-sm font-medium text-gray-700">Sync Frequency</label>
                                    <span className="text-sm text-gray-600">{formatRecurrence(recurrenceInterval)}</span>
                                </div>
                                <input
                                    type="range"
                                    id="recurrence"
                                    min="1"
                                    max="30"
                                    step="1"
                                    value={recurrenceInterval}
                                    onChange={(e) => setRecurrenceInterval(Number.parseInt(e.target.value))}
                                    className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
                                />
                                <div className="flex justify-between text-xs text-gray-500 mt-1">
                                    <span>Daily</span>
                                    <span>Weekly</span>
                                    <span>Monthly</span>
                                </div>
                            </div>
                        </div>

                        <button 
                            className="w-full py-2 px-4 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
                            onClick={addTime}
                        >
                            + Add Sync Schedule
                        </button>
                    </div>
                </div>

                <div className="px-6 py-4 bg-gray-50 border-t border-gray-200">
                    <button 
                        className="w-full py-2 px-4 bg-green-600 text-white rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
                        onClick={handleSync}
                        disabled={times.length === 0 || isLoading}
                    >
                        {isLoading ? "Syncing..." : "Save Sync Schedules"}
                    </button>
                </div>
            </div>
        </div>
    )
}
