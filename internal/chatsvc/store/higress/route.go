package higress

import gatewayservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/gateway"

type higressRoute struct {
	Name     string                 `json:"name"`
	Version  string                 `json:"version"`
	Domains  []string               `json:"domains"`
	Path     *higressRoutePredicate `json:"path"`
	Methods  []string               `json:"methods"`
	Services []higressUpstream      `json:"services"`
}

type higressRoutePredicate struct {
	MatchType     string `json:"matchType"`
	MatchValue    string `json:"matchValue"`
	CaseSensitive *bool  `json:"caseSensitive"`
}

type higressUpstream struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Port    *int   `json:"port"`
	Weight  *int   `json:"weight"`
}

func toRoutes(values []higressRoute) []gatewayservice.Route {
	routes := make([]gatewayservice.Route, 0, len(values))
	for _, value := range values {
		routes = append(routes, toRoute(value))
	}
	return routes
}

func toRoute(value higressRoute) gatewayservice.Route {
	var path *gatewayservice.RoutePredicate
	if value.Path != nil {
		path = &gatewayservice.RoutePredicate{
			Type:          value.Path.MatchType,
			Value:         value.Path.MatchValue,
			CaseSensitive: value.Path.CaseSensitive,
		}
	}

	backends := make([]gatewayservice.Backend, 0, len(value.Services))
	for _, service := range value.Services {
		backends = append(backends, gatewayservice.Backend{
			Name:    service.Name,
			Version: service.Version,
			Port:    service.Port,
			Weight:  service.Weight,
		})
	}

	return gatewayservice.Route{
		Name:     value.Name,
		Version:  value.Version,
		Domains:  value.Domains,
		Path:     path,
		Methods:  value.Methods,
		Backends: backends,
	}
}
