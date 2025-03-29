import React, { useState, useEffect, useRef } from 'react';

interface TimePickerProps {
    value: string;
    onChange: (value: string) => void;
    minTime?: string;
    className?: string;
}

export default function TimePicker({ value, onChange, minTime, className = "" }: TimePickerProps) {
    const [hours, setHours] = useState<string[]>([]);
    const [minutes, setMinutes] = useState<string[]>([]);
    const [selectedHour, setSelectedHour] = useState<string>("00");
    const [selectedMinute, setSelectedMinute] = useState<string>("00");
    const [isPM, setIsPM] = useState(false);
    const hoursRef = useRef<HTMLDivElement>(null);
    const minutesRef = useRef<HTMLDivElement>(null);
    const scrollTimeout = useRef<number>();

    const ITEM_HEIGHT = 40; // Height of each item
    const VISIBLE_ITEMS = 5; // Number of visible items
    const CONTAINER_HEIGHT = ITEM_HEIGHT * 3; // Show 3 items (1 selected + 1 above + 1 below)
    const PADDING = ITEM_HEIGHT; // Padding to allow scrolling to first/last items

    // Generate hours (1-12) and minutes (00-59)
    useEffect(() => {
        const hoursList = Array.from({ length: 12 }, (_, i) => 
            String(i + 1).padStart(2, '0')
        );
        const minutesList = Array.from({ length: 60 }, (_, i) => 
            String(i).padStart(2, '0')
        );
        setHours(hoursList);
        setMinutes(minutesList);
    }, []);

    // Initialize selected values from props
    useEffect(() => {
        if (value) {
            const [time, period] = value.split(' ');
            const [hrs, mins] = time.split(':');
            const hour = parseInt(hrs);
            const hourStr = String(hour > 12 ? hour - 12 : hour).padStart(2, '0');
            setSelectedHour(hourStr);
            setSelectedMinute(mins);
            setIsPM(hour >= 12);

            // Scroll to initial positions
            requestAnimationFrame(() => {
                if (hoursRef.current) {
                    const hourIndex = hours.indexOf(hourStr);
                    if (hourIndex !== -1) {
                        hoursRef.current.scrollTop = hourIndex * ITEM_HEIGHT;
                    }
                }
                if (minutesRef.current) {
                    const minuteIndex = minutes.indexOf(mins);
                    if (minuteIndex !== -1) {
                        minutesRef.current.scrollTop = minuteIndex * ITEM_HEIGHT;
                    }
                }
            });
        }
    }, [value, hours, minutes]);

    // Handle hour change
    const handleHourChange = (hour: string) => {
        const hourNum = parseInt(hour);
        if (hourNum < 1 || hourNum > 12) return;
        
        setSelectedHour(hour);
        const finalHour = isPM ? (hourNum === 12 ? 12 : hourNum + 12) : (hourNum === 12 ? 0 : hourNum);
        onChange(`${String(finalHour).padStart(2, '0')}:${selectedMinute}`);
    };

    // Handle minute change
    const handleMinuteChange = (minute: string) => {
        const minuteNum = parseInt(minute);
        if (minuteNum < 0 || minuteNum > 59) return;
        
        setSelectedMinute(minute);
        const hourNum = parseInt(selectedHour);
        const finalHour = isPM ? (hourNum === 12 ? 12 : hourNum + 12) : (hourNum === 12 ? 0 : hourNum);
        onChange(`${String(finalHour).padStart(2, '0')}:${minute}`);
    };

    // Handle AM/PM change
    const handleAMPMChange = (pm: boolean) => {
        setIsPM(pm);
        const hourNum = parseInt(selectedHour);
        const finalHour = pm ? (hourNum === 12 ? 12 : hourNum + 12) : (hourNum === 12 ? 0 : hourNum);
        onChange(`${String(finalHour).padStart(2, '0')}:${selectedMinute}`);
    };

    // Handle scroll events
    const handleScroll = (ref: React.RefObject<HTMLDivElement>, type: 'hour' | 'minute') => {
        // Clear any existing timeout
        if (scrollTimeout.current) {
            window.clearTimeout(scrollTimeout.current);
        }

        // Set a new timeout
        scrollTimeout.current = window.setTimeout(() => {
            if (!ref.current) return;
            
            const container = ref.current;
            const scrollTop = container.scrollTop;
            const selectedIndex = Math.round(scrollTop / ITEM_HEIGHT);
            
            if (type === 'hour' && selectedIndex < hours.length) {
                const hour = hours[selectedIndex];
                if (hour) {
                    handleHourChange(hour);
                    // Snap to position after selection
                    container.scrollTop = selectedIndex * ITEM_HEIGHT;
                }
            } else if (type === 'minute' && selectedIndex < minutes.length) {
                const minute = minutes[selectedIndex];
                if (minute) {
                    handleMinuteChange(minute);
                    // Snap to position after selection
                    container.scrollTop = selectedIndex * ITEM_HEIGHT;
                }
            }
        }, 150); // Wait for scroll to finish
    };

    return (
        <div className={`flex items-center space-x-2 ${className}`}>
            <div className="relative w-20">
                <div 
                    ref={hoursRef}
                    className="h-[120px] overflow-y-auto scroll-smooth snap-y snap-mandatory relative z-20"
                    onScroll={() => handleScroll(hoursRef, 'hour')}
                >
                    <div className="absolute inset-0 pointer-events-none z-10" style={{
                        background: 'linear-gradient(to bottom, white 0%, rgba(255,255,255,0) 25%, rgba(255,255,255,0) 75%, white 100%)'
                    }} />
                    <div className="py-[40px] relative z-0">
                        {hours.map((hour) => (
                            <div
                                key={hour}
                                className={`h-10 flex items-center justify-center snap-center cursor-pointer ${
                                    selectedHour === hour ? 'text-blue-600 font-medium' : 'text-gray-600'
                                }`}
                                onClick={() => handleHourChange(hour)}
                            >
                                {hour}
                            </div>
                        ))}
                    </div>
                </div>
                <div className="absolute inset-x-0 top-[40px] h-10 bg-gray-100/50 pointer-events-none z-0" />
            </div>

            <span className="text-gray-600">:</span>

            <div className="relative w-20">
                <div 
                    ref={minutesRef}
                    className="h-[120px] overflow-y-auto scroll-smooth snap-y snap-mandatory relative z-20"
                    onScroll={() => handleScroll(minutesRef, 'minute')}
                >
                    <div className="absolute inset-0 pointer-events-none z-10" style={{
                        background: 'linear-gradient(to bottom, white 0%, rgba(255,255,255,0) 25%, rgba(255,255,255,0) 75%, white 100%)'
                    }} />
                    <div className="py-[40px] relative z-0">
                        {minutes.map((minute) => (
                            <div
                                key={minute}
                                className={`h-10 flex items-center justify-center snap-center cursor-pointer ${
                                    selectedMinute === minute ? 'text-blue-600 font-medium' : 'text-gray-600'
                                }`}
                                onClick={() => handleMinuteChange(minute)}
                            >
                                {minute}
                            </div>
                        ))}
                    </div>
                </div>
                <div className="absolute inset-x-0 top-[40px] h-10 bg-gray-100/50 pointer-events-none z-0" />
            </div>

            <div className="flex space-x-1 ml-2">
                <button
                    onClick={() => handleAMPMChange(false)}
                    className={`px-2 py-1 text-sm rounded ${
                        !isPM ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600'
                    }`}
                >
                    AM
                </button>
                <button
                    onClick={() => handleAMPMChange(true)}
                    className={`px-2 py-1 text-sm rounded ${
                        isPM ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600'
                    }`}
                >
                    PM
                </button>
            </div>
        </div>
    );
}