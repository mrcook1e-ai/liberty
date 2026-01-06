export namespace main {
	
	export class AppSettings {
	    discord_token: string;
	    discord_guild: string;
	    discord_channel: string;
	    work_dir: string;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.discord_token = source["discord_token"];
	        this.discord_guild = source["discord_guild"];
	        this.discord_channel = source["discord_channel"];
	        this.work_dir = source["work_dir"];
	    }
	}

}

