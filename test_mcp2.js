const { spawn } = require('child_process');
const figmaMcp = spawn('npx', ['-y', '@yhy2001/figma-mcp-server'], { env: { ...process.env, FIGMA_PERSONAL_ACCESS_TOKEN: process.env.FIGMA_TOKEN } });

figmaMcp.stdout.on('data', data => console.log('OUT:', data.toString()));
figmaMcp.stderr.on('data', data => console.error('ERR:', data.toString()));

const initReq = { jsonrpc: "2.0", id: 1, method: "initialize", params: { protocolVersion: "2024-11-05", capabilities: {}, clientInfo: { name: "test", version: "1" } } };
figmaMcp.stdin.write(JSON.stringify(initReq) + "\n");

setTimeout(() => {
  const toolsReq = { jsonrpc: "2.0", id: 2, method: "tools/list", params: {} };
  figmaMcp.stdin.write(JSON.stringify(toolsReq) + "\n");
}, 2000);
setTimeout(() => process.exit(0), 4000);
