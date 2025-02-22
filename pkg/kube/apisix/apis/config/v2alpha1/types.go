// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package v2alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +genclient
// +genclient:noStatus

const (
	// OpEqual means the equal ("==") operator in nginxVars.
	OpEqual = "Equal"
	// OpNotEqual means the not equal ("~=") operator in nginxVars.
	OpNotEqual = "NotEqual"
	// OpGreaterThan means the greater than (">") operator in nginxVars.
	OpGreaterThan = "GreaterThan"
	// OpGreaterThanEqual means the greater than (">=") operator in nginxVars.
	OpGreaterThanEqual = "GreaterThanEqual"
	// OpLessThan means the less than ("<") operator in nginxVars.
	OpLessThan = "LessThan"
	// OpLessThanEqual means the less than equal ("<=") operator in nginxVars.
	OpLessThanEqual = "LessThanEqual"
	// OpRegexMatch means the regex match ("~~") operator in nginxVars.
	OpRegexMatch = "RegexMatch"
	// OpRegexNotMatch means the regex not match ("!~~") operator in nginxVars.
	OpRegexNotMatch = "RegexNotMatch"
	// OpRegexMatchCaseInsensitive means the regex match "~*" (case insensitive mode) operator in nginxVars.
	OpRegexMatchCaseInsensitive = "RegexMatchCaseInsensitive"
	// OpRegexNotMatchCaseInsensitive means the regex not match "!~*" (case insensitive mode) operator in nginxVars.
	OpRegexNotMatchCaseInsensitive = "RegexNotMatchCaseInsensitive"
	// OpIn means the in operator ("in") in nginxVars.
	OpIn = "In"
	// OpNotIn means the not in operator ("not_in") in nginxVars.
	OpNotIn = "NotIn"

	// ScopeQuery means the route match expression subject is in the querystring.
	ScopeQuery = "Query"
	// ScopeHeader means the route match expression subject is in request headers.
	ScopeHeader = "Header"
	// ScopePath means the route match expression subject is the uri path.
	ScopePath = "Path"
	// ScopeCookie means the route match expression subject is in cookie.
	ScopeCookie = "Cookie"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ApisixRoute is used to define the route rules and upstreams for Apache APISIX.
type ApisixRoute struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec              *ApisixRouteSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

// ApisixRouteSpec is the spec definition for ApisixRouteSpec.
type ApisixRouteSpec struct {
	HTTP []*ApisixRouteHTTP `json:"http,omitempty" yaml:"http,omitempty"`
}

// ApisixRouteHTTP represents a single route in for HTTP traffic.
type ApisixRouteHTTP struct {
	// The rule name, cannot be empty.
	Name string `json:"name" yaml:"name"`
	// Route priority, when multiple routes contains
	// same URI path (for path matching), route with
	// higher priority will take effect.
	Priority int                   `json:"priority,omitempty" yaml:"priority,omitempty"`
	Match    *ApisixRouteHTTPMatch `json:"match,omitempty" yaml:"match,omitempty"`
	// Deprecated: Backend will be removed in the future, use Backends instead.
	Backend *ApisixRouteHTTPBackend `json:"backend" yaml:"backend"`
	// Backends represents potential backends to proxy after the route
	// rule matched. When number of backends are more than one, traffic-split
	// plugin in APISIX will be used to split traffic based on the backend weight.
	Backends []*ApisixRouteHTTPBackend `json:"backends" yaml:"backends"`
	Plugins  []*ApisixRouteHTTPPlugin  `json:"plugins,omitempty" yaml:"plugins,omitempty"`
}

// ApisixRouteHTTPMatch represents the match condition for hitting this route.
type ApisixRouteHTTPMatch struct {
	// URI path predicates, at least one path should be
	// configured, path could be exact or prefix, for prefix path,
	// append "*" after it, for instance, "/foo*".
	Paths []string `json:"paths" yaml:"paths"`
	// HTTP request method predicates.
	Methods []string `json:"methods,omitempty" yaml:"methods,omitempty"`
	// HTTP Host predicates, host can be a wildcard domain or
	// an exact domain. For wildcard domain, only one generic
	// level is allowed, for instance, "*.foo.com" is valid but
	// "*.*.foo.com" is not.
	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	// Remote address predicates, items can be valid IPv4 address
	// or IPv6 address or CIDR.
	RemoteAddrs []string `json:"remoteAddrs,omitempty" yaml:"remoteAddrs,omitempty"`
	// NginxVars represents generic match predicates,
	// it uses Nginx variable systems, so any predicate
	// like headers, querystring and etc can be leveraged
	// here to match the route.
	// For instance, it can be:
	// nginxVars:
	//   - subject: "$remote_addr"
	//     op: in
	//     value:
	//       - "127.0.0.1"
	//       - "10.0.5.11"
	NginxVars []ApisixRouteHTTPMatchExpr `json:"exprs,omitempty" yaml:"exprs,omitempty"`
}

// ApisixRouteHTTPMatchExpre represents a binary route match expression .
type ApisixRouteHTTPMatchExpr struct {
	// Subject is the expression subject, it can
	// be any string composed by literals and nginx
	// vars.
	Subject ApisixRouteHTTPMatchExprSubject `json:"subject" yaml:"subject"`
	// Op is the operator.
	Op string `json:"op" yaml:"op"`
	// Set is an array type object of the expression.
	// It should be used when the Op is "in" or "not_in";
	Set []string `json:"set" yaml:"set"`
	// Value is the normal type object for the expression,
	// it should be used when the Op is not "in" and "not_in".
	// Set and Value are exclusive so only of them can be set
	// in the same time.
	Value *string `json:"value" yaml:"value"`
}

// ApisixRouteHTTPMatchExprSubject describes the route match expression subject.
type ApisixRouteHTTPMatchExprSubject struct {
	// The subject scope, can be:
	// ScopeQuery, ScopeHeader, ScopePath
	// when subject is ScopePath, Name field
	// will be ignored.
	Scope string `json:"scope" yaml:"scope"`
	// The name of subject.
	Name string `json:"name" yaml:"name"`
}

// ApisixRouteHTTPBackend represents a HTTP backend (a Kuberentes Service).
type ApisixRouteHTTPBackend struct {
	// The name (short) of the service, note cross namespace is forbidden,
	// so be sure the ApisixRoute and Service are in the same namespace.
	ServiceName string `json:"serviceName" yaml:"serviceName"`
	// The service port, could be the name or the port number.
	ServicePort intstr.IntOrString `json:"servicePort" yaml:"servicePort"`
	// The resolve granularity, can be "endpoints" or "service",
	// when set to "endpoints", the pod ips will be used; other
	// wise, the service ClusterIP or ExternalIP will be used,
	// default is endpoints.
	ResolveGranularity string `json:"resolveGranularity" yaml:"resolveGranularity"`
	// Weight of this backend.
	Weight int `json:"weight" yaml:"weight"`
}

// ApisixRouteHTTPPlugin represents an APISIX plugin.
type ApisixRouteHTTPPlugin struct {
	// The plugin name.
	Name string `json:"name" yaml:"name"`
	// Whether this plugin is in use, default is true.
	Enable bool `json:"enable" yaml:"enable"`
	// Plugin configuration.
	// TODO we may use protobuf to define it.
	Config ApisixRouteHTTPPluginConfig `json:"config" yaml:"config"`
}

// ApisixRouteHTTPPluginConfig is the configuration for
// any plugins.
type ApisixRouteHTTPPluginConfig map[string]interface{}

func (p ApisixRouteHTTPPluginConfig) DeepCopyInto(out *ApisixRouteHTTPPluginConfig) {
	b, _ := json.Marshal(&p)
	_ = json.Unmarshal(b, out)
}

func (p *ApisixRouteHTTPPluginConfig) DeepCopy() *ApisixRouteHTTPPluginConfig {
	if p == nil {
		return nil
	}
	out := new(ApisixRouteHTTPPluginConfig)
	p.DeepCopyInto(out)
	return out
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ApisixRouteList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata" yaml:"metadata"`
	Items           []ApisixRoute `json:"items,omitempty" yaml:"items,omitempty"`
}
