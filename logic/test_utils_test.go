package logic

import "github.com/architectmind/backend/model"

func makeCtx(nodes map[string]model.SystemNode, edges []model.SystemEdge) model.TopologyContext {
	ctx := model.TopologyContext{
		Nodes:    make([]model.SystemNode, 0, len(nodes)),
		Edges:    edges,
		NodeByID: nodes,
		Outgoing: make(map[string][]string),
	}
	for _, node := range nodes {
		ctx.Nodes = append(ctx.Nodes, node)
	}
	for _, edge := range edges {
		ctx.Outgoing[edge.Source] = append(ctx.Outgoing[edge.Source], edge.Target)
	}
	return ctx
}
