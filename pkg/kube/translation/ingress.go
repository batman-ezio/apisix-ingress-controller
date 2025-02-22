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
package translation

import (
	"bytes"
	"strings"

	"go.uber.org/zap"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/apache/apisix-ingress-controller/pkg/id"
	"github.com/apache/apisix-ingress-controller/pkg/log"
	apisixv1 "github.com/apache/apisix-ingress-controller/pkg/types/apisix/v1"
)

func (t *translator) translateIngressV1(ing *networkingv1.Ingress) ([]*apisixv1.Route, []*apisixv1.Upstream, error) {
	var (
		routes    []*apisixv1.Route
		upstreams []*apisixv1.Upstream
	)

	for _, rule := range ing.Spec.Rules {
		for _, pathRule := range rule.HTTP.Paths {
			var (
				ups *apisixv1.Upstream
				err error
			)
			if pathRule.Backend.Service != nil {
				ups, err = t.translateUpstreamFromIngressV1(ing.Namespace, pathRule.Backend.Service)
				if err != nil {
					log.Errorw("failed to translate ingress backend to upstream",
						zap.Error(err),
						zap.Any("ingress", ing),
					)
					return nil, nil, err
				}
				upstreams = append(upstreams, ups)
			}
			uris := []string{pathRule.Path}
			if pathRule.PathType != nil && *pathRule.PathType == networkingv1.PathTypePrefix {
				// As per the specification of Ingress path matching rule:
				// if the last element of the path is a substring of the
				// last element in request path, it is not a match, e.g. /foo/bar
				// matches /foo/bar/baz, but does not match /foo/barbaz.
				// While in APISIX, /foo/bar matches both /foo/bar/baz and
				// /foo/barbaz.
				// In order to be conformant with Ingress specification, here
				// we create two paths here, the first is the path itself
				// (exact match), the other is path + "/*" (prefix match).
				prefix := pathRule.Path
				if strings.HasSuffix(prefix, "/") {
					prefix += "*"
				} else {
					prefix += "/*"
				}
				uris = append(uris, prefix)
			}
			route := &apisixv1.Route{
				Metadata: apisixv1.Metadata{
					FullName: composeIngressRouteName(rule.Host, pathRule.Path),
				},
				Host: rule.Host,
				Uris: uris,
			}
			route.ID = id.GenID(route.FullName)
			route.Name = route.FullName
			if ups != nil {
				route.UpstreamId = ups.ID
			}
			routes = append(routes, route)
		}
	}
	return routes, upstreams, nil
}

func (t *translator) translateIngressV1beta1(ing *networkingv1beta1.Ingress) ([]*apisixv1.Route, []*apisixv1.Upstream, error) {
	var (
		routes    []*apisixv1.Route
		upstreams []*apisixv1.Upstream
	)

	for _, rule := range ing.Spec.Rules {
		for _, pathRule := range rule.HTTP.Paths {
			var (
				ups *apisixv1.Upstream
				err error
			)
			if pathRule.Backend.ServiceName != "" {
				ups, err = t.translateUpstreamFromIngressV1beta1(ing.Namespace, pathRule.Backend.ServiceName, pathRule.Backend.ServicePort)
				if err != nil {
					log.Errorw("failed to translate ingress backend to upstream",
						zap.Error(err),
						zap.Any("ingress", ing),
					)
					return nil, nil, err
				}
				upstreams = append(upstreams, ups)
			}
			uris := []string{pathRule.Path}
			if pathRule.PathType != nil && *pathRule.PathType == networkingv1beta1.PathTypePrefix {
				// As per the specification of Ingress path matching rule:
				// if the last element of the path is a substring of the
				// last element in request path, it is not a match, e.g. /foo/bar
				// matches /foo/bar/baz, but does not match /foo/barbaz.
				// While in APISIX, /foo/bar matches both /foo/bar/baz and
				// /foo/barbaz.
				// In order to be conformant with Ingress specification, here
				// we create two paths here, the first is the path itself
				// (exact match), the other is path + "/*" (prefix match).
				prefix := pathRule.Path
				if strings.HasSuffix(prefix, "/") {
					prefix += "*"
				} else {
					prefix += "/*"
				}
				uris = append(uris, prefix)
			}
			route := &apisixv1.Route{
				Metadata: apisixv1.Metadata{
					FullName: composeIngressRouteName(rule.Host, pathRule.Path),
				},
				Host: rule.Host,
				Uris: uris,
			}
			route.Name = route.FullName
			route.ID = id.GenID(route.FullName)
			if ups != nil {
				route.UpstreamId = ups.ID
			}
			routes = append(routes, route)
		}
	}
	return routes, upstreams, nil
}

func (t *translator) translateUpstreamFromIngressV1(namespace string, backend *networkingv1.IngressServiceBackend) (*apisixv1.Upstream, error) {
	var svcPort int32
	if backend.Port.Name != "" {
		svc, err := t.ServiceLister.Services(namespace).Get(backend.Name)
		if err != nil {
			return nil, err
		}
		for _, port := range svc.Spec.Ports {
			if port.Name == backend.Port.Name {
				svcPort = port.Port
				break
			}
		}
		if svcPort == 0 {
			return nil, &translateError{
				field:  "service",
				reason: "port not found",
			}
		}
	} else {
		svcPort = backend.Port.Number
	}
	ups, err := t.TranslateUpstream(namespace, backend.Name, svcPort)
	if err != nil {
		return nil, err
	}
	ups.FullName = apisixv1.ComposeUpstreamName(namespace, backend.Name, svcPort)
	ups.Name = ups.FullName
	ups.ID = id.GenID(ups.FullName)
	return ups, nil
}

func (t *translator) translateIngressExtensionsV1beta1(ing *extensionsv1beta1.Ingress) ([]*apisixv1.Route, []*apisixv1.Upstream, error) {
	var (
		routes    []*apisixv1.Route
		upstreams []*apisixv1.Upstream
	)

	for _, rule := range ing.Spec.Rules {
		for _, pathRule := range rule.HTTP.Paths {
			var (
				ups *apisixv1.Upstream
				err error
			)
			if pathRule.Backend.ServiceName != "" {
				// Structure here is same to ingress.extensions/v1beta1, so just use this method.
				ups, err = t.translateUpstreamFromIngressV1beta1(ing.Namespace, pathRule.Backend.ServiceName, pathRule.Backend.ServicePort)
				if err != nil {
					log.Errorw("failed to translate ingress backend to upstream",
						zap.Error(err),
						zap.Any("ingress", ing),
					)
					return nil, nil, err
				}
				upstreams = append(upstreams, ups)
			}
			uris := []string{pathRule.Path}
			if pathRule.PathType != nil && *pathRule.PathType == extensionsv1beta1.PathTypePrefix {
				// As per the specification of Ingress path matching rule:
				// if the last element of the path is a substring of the
				// last element in request path, it is not a match, e.g. /foo/bar
				// matches /foo/bar/baz, but does not match /foo/barbaz.
				// While in APISIX, /foo/bar matches both /foo/bar/baz and
				// /foo/barbaz.
				// In order to be conformant with Ingress specification, here
				// we create two paths here, the first is the path itself
				// (exact match), the other is path + "/*" (prefix match).
				prefix := pathRule.Path
				if strings.HasSuffix(prefix, "/") {
					prefix += "*"
				} else {
					prefix += "/*"
				}
				uris = append(uris, prefix)
			}
			route := &apisixv1.Route{
				Metadata: apisixv1.Metadata{
					FullName: composeIngressRouteName(rule.Host, pathRule.Path),
				},
				Host: rule.Host,
				Uris: uris,
			}
			route.Name = route.FullName
			route.ID = id.GenID(route.FullName)
			if ups != nil {
				route.UpstreamId = ups.ID
			}
			routes = append(routes, route)
		}
	}
	return routes, upstreams, nil
}

func (t *translator) translateUpstreamFromIngressV1beta1(namespace string, svcName string, svcPort intstr.IntOrString) (*apisixv1.Upstream, error) {
	var portNumber int32
	if svcPort.Type == intstr.String {
		svc, err := t.ServiceLister.Services(namespace).Get(svcName)
		if err != nil {
			return nil, err
		}
		for _, port := range svc.Spec.Ports {
			if port.Name == svcPort.StrVal {
				portNumber = port.Port
				break
			}
		}
		if portNumber == 0 {
			return nil, &translateError{
				field:  "service",
				reason: "port not found",
			}
		}
	} else {
		portNumber = svcPort.IntVal
	}
	ups, err := t.TranslateUpstream(namespace, svcName, portNumber)
	if err != nil {
		return nil, err
	}
	ups.FullName = apisixv1.ComposeUpstreamName(namespace, svcName, portNumber)
	ups.Name = ups.FullName
	ups.ID = id.GenID(ups.FullName)
	return ups, nil
}

func composeIngressRouteName(host, path string) string {
	p := make([]byte, 0, len(host)+len(path)+len("ingress")+2)
	buf := bytes.NewBuffer(p)

	buf.WriteString("ingress")
	buf.WriteByte('_')
	buf.WriteString(host)
	buf.WriteByte('_')
	buf.WriteString(path)

	return buf.String()

}
