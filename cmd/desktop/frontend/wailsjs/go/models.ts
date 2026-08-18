export namespace config {
	
	export class Settings {
	    model_path: string;
	    api_port: string;
	    context_size: number;
	    gpu_layers: number;
	    force_cpu: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model_path = source["model_path"];
	        this.api_port = source["api_port"];
	        this.context_size = source["context_size"];
	        this.gpu_layers = source["gpu_layers"];
	        this.force_cpu = source["force_cpu"];
	    }
	}

}

