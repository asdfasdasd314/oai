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

    const handleScroll = (ref: React.RefObject<HTMLDivElement>, type: 'hour' | 'minute') => {
        if (ref.current) {
            const container = ref.current;
            const itemHeight = 24; // height of each item
            const scrollPosition = container.scrollTop;
            const selectedIndex = Math.round(scrollPosition / itemHeight);
            
            if (type === 'hour' && hours[selectedIndex]) {
                handleHourChange(hours[selectedIndex]);
            } else if (type === 'minute' && minutes[selectedIndex]) {
                handleMinuteChange(minutes[selectedIndex]);
            }
        }
    };

    return (
        <div className={`flex flex-col items-center ${className}`}>
            <div className="flex overflow-hidden rounded-lg bg-white shadow-sm">
                {/* Hours */}
                <div 
                    ref={hoursRef}
                    className="h-24 w-12 overflow-y-auto scroll-smooth bg-gray-50 relative"
                    style={{
                        scrollSnapType: 'y mandatory',
                        WebkitOverflowScrolling: 'touch'
                    }}
                    onScroll={() => handleScroll(hoursRef, 'hour')}
                >
                    <div className="py-12">
                        {hours.map((hour) => (
                            <div
                                key={hour}
                                data-value={hour}
                                className={`h-6 flex items-center justify-center text-sm cursor-pointer transition-colors
                                    ${selectedHour === hour 
                                        ? 'bg-blue-100 rounded text-blue-600 font-medium' 
                                        : 'text-gray-600 hover:text-gray-900'}`}
                                onClick={() => handleHourChange(hour)}
                            >
                                {hour}
                            </div>
                        ))}
                    </div>
                </div>

                {/* Separator */}
                <div className="flex items-center px-1">
                    <span className="text-gray-400">:</span>
                </div>

                {/* Minutes */}
                <div 
                    ref={minutesRef}
                    className="h-24 w-12 overflow-y-auto scroll-smooth bg-gray-50 relative"
                    style={{
                        scrollSnapType: 'y mandatory',
                        WebkitOverflowScrolling: 'touch'
                    }}
                    onScroll={() => handleScroll(minutesRef, 'minute')}
                >
                    <div className="py-12">
                        {minutes.map((minute) => (
                            <div
                                key={minute}
                                data-value={minute}
                                className={`h-6 flex items-center justify-center text-sm cursor-pointer transition-colors
                                    ${selectedMinute === minute 
                                        ? 'bg-blue-100 rounded text-blue-600 font-medium' 
                                        : 'text-gray-600 hover:text-gray-900'}`}
                                onClick={() => handleMinuteChange(minute)}
                            >
                                {minute}
                            </div>
                        ))}
                    </div>
                </div>
            </div>

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