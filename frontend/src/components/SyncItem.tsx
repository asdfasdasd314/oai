import { useState } from "react";
import { SyncTime } from "../types/sync";
import { formatDate, formatRecurrence } from "../utils/sync";
import TimePicker from "./TimePicker";

interface SyncItemProps {
    time: SyncTime;
    onRemove: (id: string) => void;
    onUpdate: (id: string, updatedTime: SyncTime) => void;
}

export default function SyncItem({ time, onRemove, onUpdate }: SyncItemProps) {
    const [isEditing, setIsEditing] = useState(false);
    const [editedLabel, setEditedLabel] = useState(time.label);
    const [editedDate, setEditedDate] = useState(time.date.toISOString().split('T')[0]);
    const [editedTime, setEditedTime] = useState(time.date.toTimeString().slice(0, 5));
    const [editedRecurrence, setEditedRecurrence] = useState(time.recurrenceInterval);

    const handleSave = () => {
        const dateObj = new Date(`${editedDate}T${editedTime}`);
        onUpdate(time.id, {
            ...time,
            label: editedLabel.trim() || formatDate(dateObj),
            date: dateObj,
            recurrenceInterval: editedRecurrence
        });
        setIsEditing(false);
    };

    if (isEditing) {
        return (
            <div className="p-4 bg-gray-50 rounded-lg space-y-4">
                <div className="grid grid-cols-2 gap-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700">Date</label>
                        <input
                            type="date"
                            value={editedDate}
                            onChange={(e) => setEditedDate(e.target.value)}
                            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Time</label>
                        <TimePicker
                            value={editedTime}
                            onChange={setEditedTime}
                            className="mt-1"
                        />
                    </div>
                </div>
                <div>
                    <label className="block text-sm font-medium text-gray-700">Label</label>
                    <input
                        type="text"
                        value={editedLabel}
                        onChange={(e) => setEditedLabel(e.target.value)}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                    />
                </div>
                <div>
                    <div className="flex justify-between items-center mb-2">
                        <label className="block text-sm font-medium text-gray-700">Sync Frequency</label>
                        <span className="text-sm text-gray-600">{formatRecurrence(editedRecurrence)}</span>
                    </div>
                    <input
                        type="range"
                        min="1"
                        max="30"
                        step="1"
                        value={editedRecurrence}
                        onChange={(e) => setEditedRecurrence(Number.parseInt(e.target.value))}
                        className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
                    />
                </div>
                <div className="flex justify-end space-x-2">
                    <button
                        onClick={() => setIsEditing(false)}
                        className="px-3 py-1 text-sm text-gray-600 hover:text-gray-900"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        className="px-3 py-1 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
                    >
                        Save
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
            <div className="space-y-1">
                <span className="block font-medium text-gray-900">{time.label}</span>
                <span className="block text-sm text-gray-600">Starting: {formatDate(time.date)}</span>
                <span className="block text-sm text-gray-600">
                    <span className="inline-block mr-1">↻</span>
                    {formatRecurrence(time.recurrenceInterval)}
                </span>
            </div>
            <div className="flex space-x-2">
                <button 
                    className="p-2 text-gray-400 hover:text-blue-500 transition-colors"
                    onClick={() => setIsEditing(true)}
                    aria-label="Edit"
                >
                    ✎
                </button>
                <button 
                    className="p-2 text-gray-400 hover:text-red-500 transition-colors"
                    onClick={() => onRemove(time.id)}
                    aria-label="Remove"
                >
                    ✕
                </button>
            </div>
        </div>
    );
}