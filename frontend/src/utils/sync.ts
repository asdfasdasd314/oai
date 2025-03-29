import { SyncTime } from '../types/sync';

export const formatDate = (date: Date): string => {
    return date.toLocaleDateString("en-US", {
        year: "numeric",
        month: "long",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit"
    });
};

export const formatRecurrence = (days: number): string => {
    if (days === 1) return "Daily";
    if (days === 7) return "Weekly";
    if (days === 14) return "Bi-weekly";
    if (days === 30) return "Monthly";
    return `Every ${days} days`;
};

export const isValidDateTime = (date: string, time: string): boolean => {
    if (!date || !time) return false;
    
    const selectedDateTime = new Date(`${date}T${time}`);
    const now = new Date();

    return selectedDateTime > now;
}; 