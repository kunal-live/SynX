export namespace filesystem {
	
	export class FileItem {
	    name: string;
	    path: string;
	    size: number;
	    dir: boolean;
	    // Go type: time
	    modified: any;
	    sha256?: string;
	
	    static createFrom(source: any = {}) {
	        return new FileItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.dir = source["dir"];
	        this.modified = this.convertValues(source["modified"], null);
	        this.sha256 = source["sha256"];
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

export namespace history {
	
	export class Entry {
	    id: string;
	    transfer_id: string;
	    peer_id: string;
	    file_name: string;
	    size: number;
	    direction: string;
	    status: string;
	    completed_at: number;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.transfer_id = source["transfer_id"];
	        this.peer_id = source["peer_id"];
	        this.file_name = source["file_name"];
	        this.size = source["size"];
	        this.direction = source["direction"];
	        this.status = source["status"];
	        this.completed_at = source["completed_at"];
	    }
	}

}

export namespace peers {
	
	export class Peer {
	    id: string;
	    name: string;
	    address: string;
	    port: number;
	    platform: string;
	    version: string;
	    token: string;
	    public_key: string;
	    // Go type: time
	    last_seen: any;
	    status: string;
	    trusted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Peer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.address = source["address"];
	        this.port = source["port"];
	        this.platform = source["platform"];
	        this.version = source["version"];
	        this.token = source["token"];
	        this.public_key = source["public_key"];
	        this.last_seen = this.convertValues(source["last_seen"], null);
	        this.status = source["status"];
	        this.trusted = source["trusted"];
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

export namespace transfer {
	
	export class Transfer {
	    id: string;
	    peer_id: string;
	    peer_name: string;
	    direction: string;
	    source_path: string;
	    destination_path: string;
	    name: string;
	    size: number;
	    done: number;
	    progress: number;
	    speed: string;
	    eta: number;
	    chunk_size: number;
	    total_chunks: number;
	    completed_chunks: number;
	    status: string;
	    sha256: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    started_at: any;
	    // Go type: time
	    completed_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new Transfer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.peer_id = source["peer_id"];
	        this.peer_name = source["peer_name"];
	        this.direction = source["direction"];
	        this.source_path = source["source_path"];
	        this.destination_path = source["destination_path"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.done = source["done"];
	        this.progress = source["progress"];
	        this.speed = source["speed"];
	        this.eta = source["eta"];
	        this.chunk_size = source["chunk_size"];
	        this.total_chunks = source["total_chunks"];
	        this.completed_chunks = source["completed_chunks"];
	        this.status = source["status"];
	        this.sha256 = source["sha256"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.completed_at = this.convertValues(source["completed_at"], null);
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

