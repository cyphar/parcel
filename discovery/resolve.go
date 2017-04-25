/*
 * parcel: OCI discovery and read-only distribution extensions
 * Copyright (C) 2017 SUSE LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package discovery

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// DiscoveryTXTFormat describes the format used for parcel aliases.
var DiscoveryTXTFormat = regexp.MustCompile(`^cyphar\.opencontainers\.parcel\.v0=(.+);$`)

// Resolve takes a "discovery URI" and resolves it to the "canonical" discovery
// URI, which can be used to resolve and access the discovery object. Note that
// "default authority" handling is not implemented in Resolve, it must be
// handled by the caller.
//
// If an error is encountered during the resolution of the TXT entries for the
// DNS hostname, alias resolution is halted and the currently resolved alias is
// returned (with no error set). If there is more than one valid discovery TXT
// record, Resolve will choose a random record.
func Resolve(name string) (string, error) {
	if !strings.Contains(name, "/") {
		// Callers have to prepend authorities.
		return "", fmt.Errorf("discovery resolve: no authority specified")
	}

	// Parse the URI. There shouldn't be a scheme.
	discoveryURL, err := url.Parse("//" + name)
	if err != nil {
		return "", fmt.Errorf("discovery resolve %s: invalid discovery URI: %v", name, err)
	}
	// Split the host[:port].
	host := discoveryURL.Host
	port := ""
	if strings.Contains(host, ":") {
		host, port, err = net.SplitHostPort(host)
		if err != nil {
			return "", fmt.Errorf("discovery resolve %s: invalid [host:port] discovery URI: %v", name, err)
		}
	}

	// Evaluate TXT records.
	old := ""
	for old != host {
		old = host

		txts, err := net.LookupTXT(host)
		if err != nil {
			// Resolution is finished. This ignores resolution errors, but we don't care because the caller will have to hit
			break
		}

		var records []string
		for _, txt := range txts {
			matches := DiscoveryTXTFormat.FindStringSubmatch(txt)
			if len(matches) != 2 {
				continue
			}
			records = append(records, matches[1])
		}
		if len(records) == 0 {
			break
		}
		if len(records) > 1 {
			// TODO: Emit a warning.
		}
		host = records[0]
	}

	if port != "" {
		host = net.JoinHostPort(host, port)
	}

	discoveryURL.Host = host
	resolved := strings.TrimPrefix(discoveryURL.String(), "//")
	return resolved, nil
}
