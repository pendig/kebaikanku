export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set([]),
	mimeTypes: {},
	_: {
		client: {start:"_app/immutable/entry/start.D65CpDy5.js",app:"_app/immutable/entry/app.CyQbZo3B.js",imports:["_app/immutable/entry/start.D65CpDy5.js","_app/immutable/chunks/fQ1lxVII.js","_app/immutable/chunks/BuXPuSXF.js","_app/immutable/entry/app.CyQbZo3B.js","_app/immutable/chunks/BuXPuSXF.js","_app/immutable/chunks/kNaey6uv.js","_app/immutable/chunks/xihTtKlq.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js'))
		],
		remotes: {
			
		},
		routes: [
			
		],
		prerendered_routes: new Set(["/"]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();
