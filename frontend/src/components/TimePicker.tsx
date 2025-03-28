import React, { useState, useEffect, useRef } from 'react';

interface TimePickerProps {
    value: string;
    onChange: (time: string) => void;
    minTime?: string;
    className?: string;
}

const TimePicker: React.FC<TimePickerProps> = ({ value, onChange, minTime, className = '' }) => {
    const [hours, setHours] = useState<string[]>([]);
    const [minutes, setMinutes] = useState<string[]>([]);
    const [selectedHour, setSelectedHour] = useState('00');
    const [selectedMinute, setSelectedMinute] = useState('00');
    const [isPM, setIsPM] = useState(false);
    const hoursRef = useRef<HTMLDivElement>(null);
    const minutesRef = useRef<HTMLDivElement>(null);

    // Generate hours (01-12)
    useEffect(() => {
        const hoursList = Array.from({ length: 12 }, (_, i) => 
            (i + 1).toString().padStart(2, '0')
        );
        setHours(hoursList);
    }, []);

    // Generate minutes (00-59)
    useEffect(() => {
        const minutesList = Array.from({ length: 60 }, (_, i) => 
            i.toString().padStart(2, '0')
        );
        setMinutes(minutesList);
    }, []);

    // Initialize selected values from props
    useEffect(() => {
        if (value) {
            const [hour, minute] = value.split(':');
            const hourNum = parseInt(hour);
            if (hourNum > 12) {
                setSelectedHour((hourNum - 12).toString().padStart(2, '0'));
                setIsPM(true);
            } else {
                setSelectedHour(hourNum.toString().padStart(2, '0'));
                setIsPM(false);
            }
            setSelectedMinute(minute);
        }
    }, [value]);

    const handleHourChange = (hour: string) => {
        setSelectedHour(hour);
        const hourNum = parseInt(hour);
        let finalHour: string;
        if (isPM) {
            finalHour = hourNum === 12 ? '12' : (hourNum + 12).toString().padStart(2, '0');
        } else {
            finalHour = hourNum === 12 ? '00' : hour;
        }
        onChange(`${finalHour}:${selectedMinute}`);
    };

    const handleMinuteChange = (minute: string) => {
        setSelectedMinute(minute);
        const hourNum = parseInt(selectedHour);
        let finalHour: string;
        if (isPM) {
            finalHour = hourNum === 12 ? '12' : (hourNum + 12).toString().padStart(2, '0');
        } else {
            finalHour = hourNum === 12 ? '00' : selectedHour;
        }
        onChange(`${finalHour}:${minute}`);
    };

    const handleAMPMChange = (pm: boolean) => {
        setIsPM(pm);
        const hourNum = parseInt(selectedHour);
        let finalHour: string;
        if (pm) {
            finalHour = hourNum === 12 ? '12' : (hourNum + 12).toString().padStart(2, '0');
        } else {
            finalHour = hourNum === 12 ? '00' : hourNum.toString().padStart(2, '0');
        }
        onChange(`${finalHour}:${selectedMinute}`);
    };

    return (
        <div className={`flex flex-col items-center ${className}`}>
            {/* Hours */}
            <input
                placeholder="hh"
                type="number"
                value={hours}
                onChange={handleHourChange}
                 min="1"
                 max="12"
            />

            {/* Minutes */}
            <input
                placeholder="mm"
                type="number"
                value={minuges}
                onChange={handleMinuteChange}
                 min="0"
                 max="59"
            />

            {/* AM/PM Toggle */}
            <div className="mt-2 flex rounded-lg overflow-hidden bg-gray-100">
                <button
                    className={`px-4 py-1 text-sm transition-colors ${
                        !isPM 
                            ? 'bg-blue-600 text-white' 
                            : 'text-gray-600 hover:bg-gray-200'
                    }`}
                    onClick={() => handleAMPMChange(false)}
                >
                    AM
                </button>
                <button
                    className={`px-4 py-1 text-sm transition-colors ${
                        isPM 
                            ? 'bg-blue-600 text-white' 
                            : 'text-gray-600 hover:bg-gray-200'
                    }`}
                    onClick={() => handleAMPMChange(true)}
                >
                    PM
                </button>
            </div>
        </div>
    );
};

export default TimePicker;