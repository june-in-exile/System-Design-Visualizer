package logic

import (
	"fmt"
	"github.com/architectmind/backend/model"
)

// checkSPOF detects load balancers and reverse proxies with only one downstream service node,
// but suppresses the warning if that service node has Replicas > 1.
func checkSPOF(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for id, node := range ctx.NodeByID {
		if !model.NodeHasRole(node, "load_balancer") && !model.NodeHasRole(node, "reverse_proxy") {
			continue
		}
		targets := ctx.Outgoing[id]
		var serviceIDs []string
		isRedundant := false

		for _, targetID := range targets {
			target, ok := ctx.NodeByID[targetID]
			if !ok || !model.NodeHasRole(target, "service") {
				continue
			}
			serviceIDs = append(serviceIDs, targetID)

			// Check if this specific service node has multiple replicas
			props, err := model.ParseNodeProperties(target)
			if err == nil {
				if svcProps, ok := props.(*model.ServiceProperties); ok && svcProps.Replicas > 1 {
					isRedundant = true
				}
			}
		}

		// If there's only 1 distinct service node AND its replicas = 1, it's a SPOF
		if len(serviceIDs) == 1 && !isRedundant {
			warnings = append(warnings, Warning{
				Rule: "spof",
				Message: fmt.Sprintf("⚠️ Single Point of Failure (SPOF) Detected: Load Balancer %q is connected to only 1 Service node replica.",
					node.Label),
				Solution: "Increase the number of Service nodes or increase the Replicas count in the properties panel.",
				NodeIDs:  append([]string{id}, serviceIDs...),
			})
		} else if len(serviceIDs) > 1 {
			// Multiple distinct nodes also solve the SPOF
			isRedundant = true
		}
	}
	return warnings
}

// checkEntryPointSPOF warns if there is only one entry point node (LB/RP) in the system.
func checkEntryPointSPOF(ctx model.TopologyContext, role, rule, emoji, label string) []Warning {
	var matched []model.SystemNode
	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, role) {
			matched = append(matched, node)
		}
	}
	if len(matched) != 1 {
		return nil
	}

	node := matched[0]
	props, err := model.ParseNodeProperties(node)
	if err == nil {
		// Check Replicas > 1
		switch p := props.(type) {
		case *model.LoadBalancerProperties:
			if p.Replicas > 1 {
				return nil
			}
		case *model.ReverseProxyProperties:
			if p.Replicas > 1 {
				return nil
			}
		}
	}

	return []Warning{{
		Rule:     rule,
		Message:  fmt.Sprintf("%s Entry SPOF: Only 1 %s exists in the entire architecture.", emoji, label),
		Solution: fmt.Sprintf("Deploying multiple %s instances or setting Replicas to 2+ in the properties panel is recommended for production.", label),
		NodeIDs:  []string{node.ID},
	}}
}

// checkNoAutoScalingSingle warns if a service has only one replica and auto-scaling is disabled.
func checkNoAutoScalingSingle(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "service") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		svcProps, ok := props.(*model.ServiceProperties)
		if !ok {
			continue
		}
		if !svcProps.AutoScaling && svcProps.Replicas == 1 {
			warnings = append(warnings, Warning{
				Rule:     "no_autoscaling_single",
				Message:  fmt.Sprintf("⚠️ Single Replica Lacks Availability: Service %q has only 1 Replica and Auto Scaling is disabled.", node.Label),
				Solution: "Enable Auto Scaling or increase Replicas to 2+ to ensure availability and prevent bottlenecks during traffic surges.",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkNoHealthCheckBehindLB warns if a service without health check is connected to a Load Balancer.
func checkNoHealthCheckBehindLB(ctx model.TopologyContext) []Warning {
	var warnings []Warning

	// Build a map of nodes that are Load Balancers
	lbNodes := make(map[string]bool)
	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, "load_balancer") {
			lbNodes[node.ID] = true
		}
	}

	// For each edge from LB to Service, check if Service has healthCheck
	for _, edge := range ctx.Edges {
		if !lbNodes[edge.Source] {
			continue
		}
		target, ok := ctx.NodeByID[edge.Target]
		if !ok || !model.NodeHasRole(target, "service") {
			continue
		}

		props, err := model.ParseNodeProperties(target)
		if err != nil {
			continue
		}
		svcProps, ok := props.(*model.ServiceProperties)
		if !ok {
			continue
		}
		if !svcProps.HealthCheck {
			warnings = append(warnings, Warning{
				Rule:     "no_healthcheck_behind_lb",
				Message:  fmt.Sprintf("🏥 Health Check Missing: Service %q is behind a Load Balancer but has no health check enabled.", target.Label),
				Solution: "Expose a /health endpoint in the service and configure a health check interval in the Load Balancer to automatically detect and remove unhealthy instances.",
				NodeIDs:  []string{target.ID},
			})
		}
	}
	return warnings
}

// checkServerlessReplicas warns if a serverless service has manually set replicas.
func checkServerlessReplicas(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "service") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		svcProps, ok := props.(*model.ServiceProperties)
		if !ok {
			continue
		}
		if svcProps.ComputeType == "serverless" && svcProps.Replicas > 1 {
			warnings = append(warnings, Warning{
				Rule:     "serverless_replicas",
				Message:  fmt.Sprintf("⚠️ Serverless Does Not Need Manual Replicas: Service %q is a serverless computation but has multiple replicas set manually.", node.Label),
				Solution: "Replicas are managed automatically by the cloud platform in serverless mode. Setting them manually is redundant. Remove the replicas setting and let the platform scale automatically.",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}
