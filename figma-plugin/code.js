figma.showUI(__html__, { width: 260, height: 110 });

function extractNode(node) {
  // Extract essential properties
  const obj = { id: node.id, name: node.name, type: node.type };
  
  const propertiesToKeep = [
    'layoutMode', 'primaryAxisSizingMode', 'counterAxisSizingMode',
    'primaryAxisAlignItems', 'counterAxisAlignItems', 'paddingLeft',
    'paddingRight', 'paddingTop', 'paddingBottom', 'itemSpacing',
    'fills', 'strokes', 'strokeWeight', 'cornerRadius',
    'characters', 'fontSize', 'fontName', 'fontWeight', 'textAlignHorizontal', 'textAlignVertical'
  ];

  for (const prop of propertiesToKeep) {
    if (prop in node) {
      try {
        const val = node[prop];
        if (val !== undefined && val !== null) {
          // Handle figma.mixed which is a Symbol
          if (typeof val === 'symbol') {
            obj[prop] = "mixed";
          } else if (Array.isArray(val) || typeof val === 'object') {
            obj[prop] = JSON.parse(JSON.stringify(val));
          } else {
            obj[prop] = val;
          }
        }
      } catch (e) {
        // Ignore getters that throw if not applicable
      }
    }
  }

  // Recursively extract children
  if ('children' in node && node.children.length > 0) {
    obj.children = node.children.map(child => extractNode(child));
  }

  return obj;
}

figma.ui.onmessage = msg => {
  if (msg.type === 'get-selection') {
    const selection = figma.currentPage.selection;
    if (selection.length === 0) {
      figma.ui.postMessage({ type: 'error', message: 'Please select an element on the canvas first!' });
      return;
    }
    
    try {
      const data = extractNode(selection[0]);
      figma.ui.postMessage({ type: 'selection-data', data: data });
    } catch (err) {
      figma.ui.postMessage({ type: 'error', message: 'Error extracting node: ' + err.toString() });
    }
  }
};
