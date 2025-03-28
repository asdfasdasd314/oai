import React, { useState, useEffect, useRef } from 'react';

interface TimePickerProps {
    value: string;
    onChange: (time: string) => void;
    minTime?: string;
    className?: string;
}

const TimePicker: React.FC<TimePickerProps> = ({ value, onChange, minTime, className = '' }) => {
    const [hour, setHour] = useState('00');
    const [minute, setMinute] = useState('00');
    const [isPM, setIsPM] = useState(false);

    // Initialize selected values from props
    useEffect(() => {
        if (value) {
            const [hour, minute] = value.split(':');
            const hourNum = parseInt(hour);
            if (hourNum > 12) {
                setHour((hourNum - 12).toString().padStart(2, '0'));
                setIsPM(true);
            } else {
                setHour(hourNum.toString().padStart(2, '0'));
                setIsPM(false);
            }
            setMinute(minute);
        }
    }, [value]);

    const handleHourChange = (hour: string) => {
        setHour(hour);
        const hourNum = parseInt(hour);
        let finalHour: string;
        if (isPM) {
            finalHour = hourNum === 12 ? '12' : (hourNum + 12).toString().padStart(2, '0');
        } else {
            finalHour = hourNum === 12 ? '00' : hour;
        }
        onChange(`${finalHour}:${minute}`);
    };

    const handleMinuteChange = (minute: string) => {
        setMinute(minute);
        const hourNum = parseInt(hour);
        let finalHour: string;
        if (isPM) {
            finalHour = hourNum === 12 ? '12' : (hourNum + 12).toString().padStart(2, '0');
        } else {
            finalHour = hourNum === 12 ? '00' : hour;
        }
        onChange(`${finalHour}:${minute}`);
    };

    const handleAMPMChange = (pm: boolean) => {
        setIsPM(pm);
        const hourNum = parseInt(hour);
        let finalHour: string;
        if (pm) {
            finalHour = hourNum === 12 ? '12' : (hourNum + 12).toString().padStart(2, '0');
        } else {
            finalHour = hourNum === 12 ? '00' : hourNum.toString().padStart(2, '0');
        }
        onChange(`${finalHour}:${minute}`);
    };

    return (
        <div className={`flex flex-col items-center ${className}`}>
            {/* Hours */}
            <input
                placeholder="hh"
                type="number"
                value={hour}
                onChange={handleHourChange}
                 min="1"
                 max="12"
            />

            {/* Minutes */}
            <input
                placeholder="mm"
                type="number"
                value={minute}
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