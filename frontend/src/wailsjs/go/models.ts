export namespace backend {
	
	export class CookieInfo {
	    name: string;
	    value: string;
	    domain: string;
	    path: string;
	    expires: number;
	    httpOnly: boolean;
	    secure: boolean;
	    sameSite: string;
	
	    static createFrom(source: any = {}) {
	        return new CookieInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	        this.domain = source["domain"];
	        this.path = source["path"];
	        this.expires = source["expires"];
	        this.httpOnly = source["httpOnly"];
	        this.secure = source["secure"];
	        this.sameSite = source["sameSite"];
	    }
	}
	export class ProxyIPHealthResult {
	    proxyId: string;
	    ok: boolean;
	    source: string;
	    error: string;
	    ip: string;
	    fraudScore: number;
	    isResidential: boolean;
	    isBroadcast: boolean;
	    country: string;
	    region: string;
	    city: string;
	    asOrganization: string;
	    rawData: Record<string, any>;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyIPHealthResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyId = source["proxyId"];
	        this.ok = source["ok"];
	        this.source = source["source"];
	        this.error = source["error"];
	        this.ip = source["ip"];
	        this.fraudScore = source["fraudScore"];
	        this.isResidential = source["isResidential"];
	        this.isBroadcast = source["isBroadcast"];
	        this.country = source["country"];
	        this.region = source["region"];
	        this.city = source["city"];
	        this.asOrganization = source["asOrganization"];
	        this.rawData = source["rawData"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ProxyTestResult {
	    proxyId: string;
	    ok: boolean;
	    latencyMs: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyId = source["proxyId"];
	        this.ok = source["ok"];
	        this.latencyMs = source["latencyMs"];
	        this.error = source["error"];
	    }
	}
	export class ProxyValidationResult {
	    supported: boolean;
	    errorMsg: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supported = source["supported"];
	        this.errorMsg = source["errorMsg"];
	    }
	}
	export class SnapshotInfo {
	    snapshotId: string;
	    profileId: string;
	    name: string;
	    sizeMB: number;
	    createdAt: string;
	    filePath?: string;
	
	    static createFrom(source: any = {}) {
	        return new SnapshotInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.snapshotId = source["snapshotId"];
	        this.profileId = source["profileId"];
	        this.name = source["name"];
	        this.sizeMB = source["sizeMB"];
	        this.createdAt = source["createdAt"];
	        this.filePath = source["filePath"];
	    }
	}

}

export namespace backup {
	
	export class ManifestEntry {
	    id: string;
	    category: string;
	    entryType: string;
	    required: boolean;
	    archivePath: string;
	    description?: string;
	
	    static createFrom(source: any = {}) {
	        return new ManifestEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.entryType = source["entryType"];
	        this.required = source["required"];
	        this.archivePath = source["archivePath"];
	        this.description = source["description"];
	    }
	}
	export class ManifestAppInfo {
	    name: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new ManifestAppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	    }
	}
	export class Manifest {
	    format: string;
	    manifestVersion: number;
	    createdAt: string;
	    app: ManifestAppInfo;
	    entries: ManifestEntry[];
	
	    static createFrom(source: any = {}) {
	        return new Manifest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.manifestVersion = source["manifestVersion"];
	        this.createdAt = source["createdAt"];
	        this.app = this.convertValues(source["app"], ManifestAppInfo);
	        this.entries = this.convertValues(source["entries"], ManifestEntry);
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
	
	
	export class ScopeEntry {
	    id: string;
	    category: string;
	    entryType: string;
	    required: boolean;
	    sourcePath: string;
	    archivePath: string;
	    exists: boolean;
	    description?: string;
	
	    static createFrom(source: any = {}) {
	        return new ScopeEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.entryType = source["entryType"];
	        this.required = source["required"];
	        this.sourcePath = source["sourcePath"];
	        this.archivePath = source["archivePath"];
	        this.exists = source["exists"];
	        this.description = source["description"];
	    }
	}
	export class Scope {
	    format: string;
	    manifestVersion: number;
	    appRoot: string;
	    entries: ScopeEntry[];
	
	    static createFrom(source: any = {}) {
	        return new Scope(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.manifestVersion = source["manifestVersion"];
	        this.appRoot = source["appRoot"];
	        this.entries = this.convertValues(source["entries"], ScopeEntry);
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

export namespace browser {
	
	export class CoreExtendedInfo {
	    coreId: string;
	    chromeVersion: string;
	    instanceCount: number;
	
	    static createFrom(source: any = {}) {
	        return new CoreExtendedInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coreId = source["coreId"];
	        this.chromeVersion = source["chromeVersion"];
	        this.instanceCount = source["instanceCount"];
	    }
	}
	export class CoreInput {
	    coreId: string;
	    coreName: string;
	    corePath: string;
	    isDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CoreInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coreId = source["coreId"];
	        this.coreName = source["coreName"];
	        this.corePath = source["corePath"];
	        this.isDefault = source["isDefault"];
	    }
	}
	export class CoreValidateResult {
	    valid: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new CoreValidateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.message = source["message"];
	    }
	}
	export class Group {
	    groupId: string;
	    groupName: string;
	    parentId: string;
	    sortOrder: number;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Group(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groupId = source["groupId"];
	        this.groupName = source["groupName"];
	        this.parentId = source["parentId"];
	        this.sortOrder = source["sortOrder"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class GroupInput {
	    groupName: string;
	    parentId: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new GroupInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groupName = source["groupName"];
	        this.parentId = source["parentId"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class GroupWithCount {
	    groupId: string;
	    groupName: string;
	    parentId: string;
	    sortOrder: number;
	    createdAt: string;
	    updatedAt: string;
	    instanceCount: number;
	
	    static createFrom(source: any = {}) {
	        return new GroupWithCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groupId = source["groupId"];
	        this.groupName = source["groupName"];
	        this.parentId = source["parentId"];
	        this.sortOrder = source["sortOrder"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.instanceCount = source["instanceCount"];
	    }
	}
	export class Profile {
	    profileId: string;
	    profileName: string;
	    userDataDir: string;
	    coreId: string;
	    fingerprintArgs: string[];
	    proxyId: string;
	    proxyConfig: string;
	    proxyBindSourceId: string;
	    proxyBindSourceUrl: string;
	    proxyBindName: string;
	    proxyBindUpdatedAt: string;
	    launchArgs: string[];
	    tags: string[];
	    keywords: string[];
	    groupId: string;
	    launchCode: string;
	    running: boolean;
	    debugPort: number;
	    debugReady: boolean;
	    pid: number;
	    runtimeWarning: string;
	    lastError: string;
	    createdAt: string;
	    updatedAt: string;
	    lastStartAt: string;
	    lastStopAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profileId = source["profileId"];
	        this.profileName = source["profileName"];
	        this.userDataDir = source["userDataDir"];
	        this.coreId = source["coreId"];
	        this.fingerprintArgs = source["fingerprintArgs"];
	        this.proxyId = source["proxyId"];
	        this.proxyConfig = source["proxyConfig"];
	        this.proxyBindSourceId = source["proxyBindSourceId"];
	        this.proxyBindSourceUrl = source["proxyBindSourceUrl"];
	        this.proxyBindName = source["proxyBindName"];
	        this.proxyBindUpdatedAt = source["proxyBindUpdatedAt"];
	        this.launchArgs = source["launchArgs"];
	        this.tags = source["tags"];
	        this.keywords = source["keywords"];
	        this.groupId = source["groupId"];
	        this.launchCode = source["launchCode"];
	        this.running = source["running"];
	        this.debugPort = source["debugPort"];
	        this.debugReady = source["debugReady"];
	        this.pid = source["pid"];
	        this.runtimeWarning = source["runtimeWarning"];
	        this.lastError = source["lastError"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.lastStartAt = source["lastStartAt"];
	        this.lastStopAt = source["lastStopAt"];
	    }
	}
	export class ProfileInput {
	    profileName: string;
	    userDataDir: string;
	    coreId: string;
	    fingerprintArgs: string[];
	    proxyId: string;
	    proxyConfig: string;
	    launchArgs: string[];
	    tags: string[];
	    keywords: string[];
	    groupId: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfileInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profileName = source["profileName"];
	        this.userDataDir = source["userDataDir"];
	        this.coreId = source["coreId"];
	        this.fingerprintArgs = source["fingerprintArgs"];
	        this.proxyId = source["proxyId"];
	        this.proxyConfig = source["proxyConfig"];
	        this.launchArgs = source["launchArgs"];
	        this.tags = source["tags"];
	        this.keywords = source["keywords"];
	        this.groupId = source["groupId"];
	    }
	}
	export class Settings {
	    userDataRoot: string;
	    defaultFingerprintArgs: string[];
	    defaultLaunchArgs: string[];
	    defaultProxy: string;
	    startReadyTimeoutMs: number;
	    startStableWindowMs: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.userDataRoot = source["userDataRoot"];
	        this.defaultFingerprintArgs = source["defaultFingerprintArgs"];
	        this.defaultLaunchArgs = source["defaultLaunchArgs"];
	        this.defaultProxy = source["defaultProxy"];
	        this.startReadyTimeoutMs = source["startReadyTimeoutMs"];
	        this.startStableWindowMs = source["startStableWindowMs"];
	    }
	}
	export class Tab {
	    tabId: string;
	    title: string;
	    url: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Tab(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tabId = source["tabId"];
	        this.title = source["title"];
	        this.url = source["url"];
	        this.active = source["active"];
	    }
	}

}

export namespace config {
	
	export class BrowserBookmark {
	    name: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new BrowserBookmark(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	    }
	}
	export class BrowserCore {
	    coreId: string;
	    coreName: string;
	    corePath: string;
	    isDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BrowserCore(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coreId = source["coreId"];
	        this.coreName = source["coreName"];
	        this.corePath = source["corePath"];
	        this.isDefault = source["isDefault"];
	    }
	}
	export class BrowserProxy {
	    proxyId: string;
	    proxyName: string;
	    proxyConfig: string;
	    dnsServers?: string;
	    groupName?: string;
	    sortOrder?: number;
	    sourceId?: string;
	    sourceUrl?: string;
	    sourceNamePrefix?: string;
	    sourceAutoRefresh?: boolean;
	    sourceRefreshIntervalM?: number;
	    sourceLastRefreshAt?: string;
	    lastLatencyMs: number;
	    lastTestOk: boolean;
	    lastTestedAt: string;
	    lastIPHealthJson?: string;
	
	    static createFrom(source: any = {}) {
	        return new BrowserProxy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyId = source["proxyId"];
	        this.proxyName = source["proxyName"];
	        this.proxyConfig = source["proxyConfig"];
	        this.dnsServers = source["dnsServers"];
	        this.groupName = source["groupName"];
	        this.sortOrder = source["sortOrder"];
	        this.sourceId = source["sourceId"];
	        this.sourceUrl = source["sourceUrl"];
	        this.sourceNamePrefix = source["sourceNamePrefix"];
	        this.sourceAutoRefresh = source["sourceAutoRefresh"];
	        this.sourceRefreshIntervalM = source["sourceRefreshIntervalM"];
	        this.sourceLastRefreshAt = source["sourceLastRefreshAt"];
	        this.lastLatencyMs = source["lastLatencyMs"];
	        this.lastTestOk = source["lastTestOk"];
	        this.lastTestedAt = source["lastTestedAt"];
	        this.lastIPHealthJson = source["lastIPHealthJson"];
	    }
	}

}

export namespace launchcode {
	
	export class LaunchRequestParams {
	    launchArgs: string[];
	    startUrls: string[];
	    skipDefaultStartUrls: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LaunchRequestParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.launchArgs = source["launchArgs"];
	        this.startUrls = source["startUrls"];
	        this.skipDefaultStartUrls = source["skipDefaultStartUrls"];
	    }
	}

}

export namespace logger {
	
	export class MemoryLogEntry {
	    time: string;
	    level: string;
	    component: string;
	    message: string;
	    fields?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new MemoryLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.level = source["level"];
	        this.component = source["component"];
	        this.message = source["message"];
	        this.fields = source["fields"];
	    }
	}
	export class MethodInterceptor {
	
	
	    static createFrom(source: any = {}) {
	        return new MethodInterceptor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace rpa {
	
	export class FlowEdge {
	    edgeId: string;
	    sourceNodeId: string;
	    targetNodeId: string;
	    condition: string;
	
	    static createFrom(source: any = {}) {
	        return new FlowEdge(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.edgeId = source["edgeId"];
	        this.sourceNodeId = source["sourceNodeId"];
	        this.targetNodeId = source["targetNodeId"];
	        this.condition = source["condition"];
	    }
	}
	export class FlowPosition {
	    x: number;
	    y: number;
	
	    static createFrom(source: any = {}) {
	        return new FlowPosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	    }
	}
	export class FlowNode {
	    nodeId: string;
	    nodeType: string;
	    label: string;
	    position: FlowPosition;
	    config: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new FlowNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeId = source["nodeId"];
	        this.nodeType = source["nodeType"];
	        this.label = source["label"];
	        this.position = this.convertValues(source["position"], FlowPosition);
	        this.config = source["config"];
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
	export class FlowVariable {
	    name: string;
	    type: string;
	    defaultValue: string;
	
	    static createFrom(source: any = {}) {
	        return new FlowVariable(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.defaultValue = source["defaultValue"];
	    }
	}
	export class FlowDocument {
	    schemaVersion: number;
	    variables: FlowVariable[];
	    nodes: FlowNode[];
	    edges: FlowEdge[];
	
	    static createFrom(source: any = {}) {
	        return new FlowDocument(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schemaVersion = source["schemaVersion"];
	        this.variables = this.convertValues(source["variables"], FlowVariable);
	        this.nodes = this.convertValues(source["nodes"], FlowNode);
	        this.edges = this.convertValues(source["edges"], FlowEdge);
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
	export class FlowStep {
	    stepId: string;
	    stepName: string;
	    stepType: string;
	    config: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new FlowStep(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stepId = source["stepId"];
	        this.stepName = source["stepName"];
	        this.stepType = source["stepType"];
	        this.config = source["config"];
	    }
	}
	export class Flow {
	    flowId: string;
	    flowName: string;
	    groupId: string;
	    steps: FlowStep[];
	    document: FlowDocument;
	    sourceType: string;
	    sourceXml: string;
	    shareCode: string;
	    version: number;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Flow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.flowId = source["flowId"];
	        this.flowName = source["flowName"];
	        this.groupId = source["groupId"];
	        this.steps = this.convertValues(source["steps"], FlowStep);
	        this.document = this.convertValues(source["document"], FlowDocument);
	        this.sourceType = source["sourceType"];
	        this.sourceXml = source["sourceXml"];
	        this.shareCode = source["shareCode"];
	        this.version = source["version"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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
	
	
	export class FlowGroup {
	    groupId: string;
	    groupName: string;
	    sortOrder: number;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new FlowGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groupId = source["groupId"];
	        this.groupName = source["groupName"];
	        this.sortOrder = source["sortOrder"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class FlowGroupInput {
	    groupName: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new FlowGroupInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groupName = source["groupName"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	
	
	
	
	export class FlowXMLImportInput {
	    flowName: string;
	    xmlText: string;
	    groupId: string;
	
	    static createFrom(source: any = {}) {
	        return new FlowXMLImportInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.flowName = source["flowName"];
	        this.xmlText = source["xmlText"];
	        this.groupId = source["groupId"];
	    }
	}
	export class Run {
	    runId: string;
	    taskId: string;
	    flowId: string;
	    triggerType: string;
	    status: string;
	    summary: string;
	    startedAt: string;
	    finishedAt: string;
	    errorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new Run(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.taskId = source["taskId"];
	        this.flowId = source["flowId"];
	        this.triggerType = source["triggerType"];
	        this.status = source["status"];
	        this.summary = source["summary"];
	        this.startedAt = source["startedAt"];
	        this.finishedAt = source["finishedAt"];
	        this.errorMessage = source["errorMessage"];
	    }
	}
	export class RunTarget {
	    runTargetId: string;
	    runId: string;
	    profileId: string;
	    profileName: string;
	    status: string;
	    stepIndex: number;
	    startedAt: string;
	    finishedAt: string;
	    errorMessage: string;
	    debugPort: number;
	
	    static createFrom(source: any = {}) {
	        return new RunTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runTargetId = source["runTargetId"];
	        this.runId = source["runId"];
	        this.profileId = source["profileId"];
	        this.profileName = source["profileName"];
	        this.status = source["status"];
	        this.stepIndex = source["stepIndex"];
	        this.startedAt = source["startedAt"];
	        this.finishedAt = source["finishedAt"];
	        this.errorMessage = source["errorMessage"];
	        this.debugPort = source["debugPort"];
	    }
	}
	export class Task {
	    taskId: string;
	    taskName: string;
	    flowId: string;
	    executionOrder: string;
	    taskType: string;
	    scheduleConfig: Record<string, any>;
	    enabled: boolean;
	    lastRunAt: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.taskId = source["taskId"];
	        this.taskName = source["taskName"];
	        this.flowId = source["flowId"];
	        this.executionOrder = source["executionOrder"];
	        this.taskType = source["taskType"];
	        this.scheduleConfig = source["scheduleConfig"];
	        this.enabled = source["enabled"];
	        this.lastRunAt = source["lastRunAt"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class TaskTarget {
	    taskId: string;
	    profileId: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new TaskTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.taskId = source["taskId"];
	        this.profileId = source["profileId"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class TaskDetail {
	    task?: Task;
	    targets: TaskTarget[];
	
	    static createFrom(source: any = {}) {
	        return new TaskDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.task = this.convertValues(source["task"], Task);
	        this.targets = this.convertValues(source["targets"], TaskTarget);
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
	
	export class Template {
	    templateId: string;
	    templateName: string;
	    description: string;
	    tags: string[];
	    flowSnapshot: Flow;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Template(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.templateId = source["templateId"];
	        this.templateName = source["templateName"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	        this.flowSnapshot = this.convertValues(source["flowSnapshot"], Flow);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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

