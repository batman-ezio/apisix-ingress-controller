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
	"errors"

	"github.com/apache/apisix-ingress-controller/pkg/id"
	configv1 "github.com/apache/apisix-ingress-controller/pkg/kube/apisix/apis/config/v1"
	apisix "github.com/apache/apisix-ingress-controller/pkg/types/apisix/v1"
	apisixv1 "github.com/apache/apisix-ingress-controller/pkg/types/apisix/v1"
)

var (
	// ErrEmptyCert means the cert field in Kubernetes Secret is not found.
	ErrEmptyCert = errors.New("missing cert field")
	// ErrEmptyPrivKey means the key field in Kubernetes Secret is not found.
	ErrEmptyPrivKey = errors.New("missing key field")
)

func (t *translator) TranslateSSL(tls *configv1.ApisixTls) (*apisixv1.Ssl, error) {
	s, err := t.SecretLister.Secrets(tls.Spec.Secret.Namespace).Get(tls.Spec.Secret.Name)
	if err != nil {
		return nil, err
	}
	cert, ok := s.Data["cert"]
	if !ok {
		return nil, ErrEmptyCert
	}
	key, ok := s.Data["key"]
	if !ok {
		return nil, ErrEmptyPrivKey
	}
	var snis []string
	snis = append(snis, tls.Spec.Hosts...)
	fullname := tls.Namespace + "_" + tls.Name
	ssl := &apisix.Ssl{
		ID:       id.GenID(fullname),
		FullName: fullname,
		Snis:     snis,
		Cert:     string(cert),
		Key:      string(key),
		Status:   1,
	}
	return ssl, nil
}
