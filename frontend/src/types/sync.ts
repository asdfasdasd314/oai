export interface SyncTime {
    id: string;
    date: Date;
    label: string;
    recurrenceInterval: number; // Days between syncs
}

export interface SyncListProps {
    initialTimes?: SyncTime[];
    onSync?: (times: SyncTime[]) => Promise<void>;
    title?: string;
} 