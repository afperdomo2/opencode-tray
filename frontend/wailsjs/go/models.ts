export namespace main {
	
	export class ServiceStatusResult {
	    running: boolean;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new ServiceStatusResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.url = source["url"];
	    }
	}

}

export namespace shell {
	
	export class Settings {
	    language: string;
	    autoStart: boolean;
	    refreshMin: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.autoStart = source["autoStart"];
	        this.refreshMin = source["refreshMin"];
	    }
	}

}

