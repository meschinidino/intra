package broker

import "intra/internal/policy"

func actionOf(s string) policy.Action {
	switch policy.Action(s) {
	case policy.ActionRead, policy.ActionWrite, policy.ActionAdmin, policy.ActionExport, policy.ActionConnect, policy.ActionEgress:
		return policy.Action(s)
	default:
		return policy.ActionAdmin // unknown actions fail closed under normal grant scope
	}
}
