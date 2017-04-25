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
	"net/url"

	parcelv0 "github.com/cyphar/parcel/specs-go/v0"
	"github.com/jtacoma/uritemplates"
)

// Explore takes a discovery URI (the caller does not need to pass it through
// discovery.Resolve), fetches it using discovery.Fetch, evaluates the
// distribution URI template and returns a URL which the caller can fetch to
// get the distribution object.
func Explore(name string) (string, error) {
	// Make the name canonical.
	name, err := Resolve(name)
	if err != nil {
		return "", err
	}

	// Get and parse the discovery blobs.
	discovery, err := Fetch(name)
	if err != nil {
		return "", err
	}
	distTemplate, err := uritemplates.Parse(discovery.DistributionURI.Template)
	if err != nil {
		return "", err
	}

	// Expand the distribution template.
	dist, err := distTemplate.Expand(map[string]interface{}{
		"parcel.version": parcelv0.Version,
		// TODO: Add authority and userAuthority.
		// TODO: Add name, nameDigest, digestAlgorithm.
	})
	if err != nil {
		return "", err
	}

	// Parse the distribution URI.
	distUri, err := url.Parse(dist)
	if err != nil {
		return "", err
	}

	// Create http://<authority>/
	baseUri, err := url.Parse("http://" + name)
	if err != nil {
		return "", err
	}
	baseUri.RawPath = ""
	baseUri.Path = ""

	// Compute the relative URI as per RFC 3986.
	uri := baseUri.ResolveReference(distUri)
	return uri.String(), nil
}
