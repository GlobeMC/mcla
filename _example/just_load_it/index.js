
'use strict';

function sleep(ms){
	return new Promise((re) => {
		setTimeout(re, ms)
	})
}

(async function(){
	const go = new Go()
	const MCLA_WASM_URL = 'https://globemc.github.io/mcla/mcla.wasm'

	const res = await WebAssembly.instantiateStreaming(fetch(MCLA_WASM_URL), go.importObject)
	go.run(res.instance)
	while(!window.MCLA){
		await sleep(10)
	}

	console.log('MCLA:', MCLA.version)
})();
