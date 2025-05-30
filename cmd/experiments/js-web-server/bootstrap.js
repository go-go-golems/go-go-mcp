let globalCounter = 0;

registerHandler("GET", "/", () => "JS playground online");
registerHandler("GET", "/health", () => ({ok: true, counter: globalCounter}));
registerHandler("POST", "/counter", () => ({count: ++globalCounter}));

console.log("Bootstrap complete - server ready");