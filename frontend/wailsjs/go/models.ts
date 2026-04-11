export namespace config {
	
	export class ProviderConfig {
	    name: string;
	    apiKey: string;
	    baseUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.apiKey = source["apiKey"];
	        this.baseUrl = source["baseUrl"];
	    }
	}
	export class RefinementProfile {
	    id: string;
	    name: string;
	    icon: string;
	    model: string;
	    prompt: string;
	
	    static createFrom(source: any = {}) {
	        return new RefinementProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.icon = source["icon"];
	        this.model = source["model"];
	        this.prompt = source["prompt"];
	    }
	}
	export class Settings {
	    providers: Record<string, ProviderConfig>;
	    transcriptionProviderId: string;
	    transcriptionModel: string;
	    refinementProviderId: string;
	    refinementProfiles: RefinementProfile[];
	    activeProfileId: string;
	    systemPrompt?: string;
	    refinementModel?: string;
	    setupComplete: boolean;
	    theme: string;
	    autoCopy: boolean;
	    enableHistory: boolean;
	    enableIndicator: boolean;
	    enableNotifications: boolean;
	    inputDeviceId?: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providers = this.convertValues(source["providers"], ProviderConfig, true);
	        this.transcriptionProviderId = source["transcriptionProviderId"];
	        this.transcriptionModel = source["transcriptionModel"];
	        this.refinementProviderId = source["refinementProviderId"];
	        this.refinementProfiles = this.convertValues(source["refinementProfiles"], RefinementProfile);
	        this.activeProfileId = source["activeProfileId"];
	        this.systemPrompt = source["systemPrompt"];
	        this.refinementModel = source["refinementModel"];
	        this.setupComplete = source["setupComplete"];
	        this.theme = source["theme"];
	        this.autoCopy = source["autoCopy"];
	        this.enableHistory = source["enableHistory"];
	        this.enableIndicator = source["enableIndicator"];
	        this.enableNotifications = source["enableNotifications"];
	        this.inputDeviceId = source["inputDeviceId"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace desktop {
	
	export class PlatformInfo {
	    os: string;
	    desktop: string;
	
	    static createFrom(source: any = {}) {
	        return new PlatformInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.desktop = source["desktop"];
	    }
	}
	export class ProcessResult {
	    transcript: string;
	    refined: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProcessResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.transcript = source["transcript"];
	        this.refined = source["refined"];
	        this.error = source["error"];
	    }
	}

}

export namespace history {
	
	export class Record {
	    id: number;
	    // Go type: time
	    timestamp: any;
	    rawTranscript: string;
	    refinedMessage: string;
	    activeApp: string;
	
	    static createFrom(source: any = {}) {
	        return new Record(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.rawTranscript = source["rawTranscript"];
	        this.refinedMessage = source["refinedMessage"];
	        this.activeApp = source["activeApp"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

