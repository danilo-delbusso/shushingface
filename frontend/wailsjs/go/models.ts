export namespace ai {
	
	export class ModelInfo {
	    id: string;
	    name: string;
	    category: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.category = source["category"];
	    }
	}
	export class ProviderInfo {
	    id: string;
	    displayName: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.displayName = source["displayName"];
	    }
	}

}

export namespace config {
	
	export class Connection {
	    id: string;
	    name: string;
	    providerId: string;
	    apiKey: string;
	    baseUrl?: string;
	
	    static createFrom(source: any = {}) {
	        return new Connection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.providerId = source["providerId"];
	        this.apiKey = source["apiKey"];
	        this.baseUrl = source["baseUrl"];
	    }
	}
	export class FewShotExample {
	    input: string;
	    output: string;
	
	    static createFrom(source: any = {}) {
	        return new FewShotExample(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.input = source["input"];
	        this.output = source["output"];
	    }
	}
	export class RefinementProfile {
	    id: string;
	    name: string;
	    icon: string;
	    connectionId?: string;
	    model: string;
	    prompt: string;
	    examples?: FewShotExample[];
	    temperature?: number;
	    topP?: number;
	
	    static createFrom(source: any = {}) {
	        return new RefinementProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.icon = source["icon"];
	        this.connectionId = source["connectionId"];
	        this.model = source["model"];
	        this.prompt = source["prompt"];
	        this.examples = this.convertValues(source["examples"], FewShotExample);
	        this.temperature = source["temperature"];
	        this.topP = source["topP"];
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
	export class legacyProviderConfig {
	    name: string;
	    apiKey: string;
	    baseUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new legacyProviderConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.apiKey = source["apiKey"];
	        this.baseUrl = source["baseUrl"];
	    }
	}
	export class Settings {
	    connections: Connection[];
	    transcriptionConnectionId: string;
	    transcriptionModel: string;
	    transcriptionLanguage?: string;
	    refinementConnectionId: string;
	    refinementModel: string;
	    refinementProfiles: RefinementProfile[];
	    activeProfileId: string;
	    globalRules?: string;
	    builtInRules?: string;
	    setupComplete: boolean;
	    theme: string;
	    autoPaste: boolean;
	    autoCopy: boolean;
	    enableHistory: boolean;
	    enableIndicator: boolean;
	    enableNotifications: boolean;
	    inputDeviceId?: string;
	    providerId?: string;
	    providerApiKey?: string;
	    providerBaseUrl?: string;
	    providers?: Record<string, legacyProviderConfig>;
	    transcriptionProviderId?: string;
	    refinementProviderId?: string;
	    systemPrompt?: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connections = this.convertValues(source["connections"], Connection);
	        this.transcriptionConnectionId = source["transcriptionConnectionId"];
	        this.transcriptionModel = source["transcriptionModel"];
	        this.transcriptionLanguage = source["transcriptionLanguage"];
	        this.refinementConnectionId = source["refinementConnectionId"];
	        this.refinementModel = source["refinementModel"];
	        this.refinementProfiles = this.convertValues(source["refinementProfiles"], RefinementProfile);
	        this.activeProfileId = source["activeProfileId"];
	        this.globalRules = source["globalRules"];
	        this.builtInRules = source["builtInRules"];
	        this.setupComplete = source["setupComplete"];
	        this.theme = source["theme"];
	        this.autoPaste = source["autoPaste"];
	        this.autoCopy = source["autoCopy"];
	        this.enableHistory = source["enableHistory"];
	        this.enableIndicator = source["enableIndicator"];
	        this.enableNotifications = source["enableNotifications"];
	        this.inputDeviceId = source["inputDeviceId"];
	        this.providerId = source["providerId"];
	        this.providerApiKey = source["providerApiKey"];
	        this.providerBaseUrl = source["providerBaseUrl"];
	        this.providers = this.convertValues(source["providers"], legacyProviderConfig, true);
	        this.transcriptionProviderId = source["transcriptionProviderId"];
	        this.refinementProviderId = source["refinementProviderId"];
	        this.systemPrompt = source["systemPrompt"];
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
	
	export class PasteStatus {
	    available: boolean;
	    installCmd: string;
	
	    static createFrom(source: any = {}) {
	        return new PasteStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.installCmd = source["installCmd"];
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
	    error?: string;

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
	        this.error = source["error"];
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

export namespace platform {
	
	export class Info {
	    os: string;
	    displayServer: string;
	    desktop: string;
	    packageManager: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.displayServer = source["displayServer"];
	        this.desktop = source["desktop"];
	        this.packageManager = source["packageManager"];
	    }
	}

}

