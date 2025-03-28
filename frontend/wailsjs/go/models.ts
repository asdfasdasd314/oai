export namespace main {
	
	export class SyncTime {
	    CurrTimestamp: number;
	    DaysBetweenSync: number;
	
	    static createFrom(source: any = {}) {
	        return new SyncTime(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CurrTimestamp = source["CurrTimestamp"];
	        this.DaysBetweenSync = source["DaysBetweenSync"];
	    }
	}

}

