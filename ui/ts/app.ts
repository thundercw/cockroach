// source: app.ts
/// <reference path="../bower_components/mithriljs/mithril.d.ts" />

/// <reference path="pages/navigation.ts" />
/// <reference path="pages/graph.ts" />
/// <reference path="pages/log.ts" />
/// <reference path="pages/nodes.ts" />
/// <reference path="pages/stores.ts" />

// Author: Bram Gruneir (bram+code@cockroachlabs.com)

m.mount(document.getElementById("header"), AdminViews.SubModules.TitleBar);

m.route.mode = "hash";
m.route(document.getElementById("root"), "/nodes", {
  "/graph": AdminViews.Graph.Page,
  "/logs": AdminViews.Log.Page,
  "/logs/:node_id": AdminViews.Log.Page,
  "/node": AdminViews.Nodes.NodesPage,
  "/nodes": AdminViews.Nodes.NodesPage,
  "/node/:node_id": AdminViews.Nodes.NodePage,
  "/nodes/:node_id": AdminViews.Nodes.NodePage,
  "/node/:node_id/:detail": AdminViews.Nodes.NodePage,
  "/nodes/:node_id/:detail": AdminViews.Nodes.NodePage,
  "/store": AdminViews.Stores.StorePage,
  "/stores": AdminViews.Stores.StoresPage,
  "/store/:store_id": AdminViews.Stores.StorePage,
  "/stores/:store_id": AdminViews.Stores.StorePage,
  "/store/:store_id/:detail": AdminViews.Stores.StorePage,
  "/stores/:store_id/:detail": AdminViews.Stores.StorePage,
});
