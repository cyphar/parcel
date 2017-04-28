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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	parcelv0 "github.com/cyphar/parcel/specs-go/v0"
	"github.com/jtacoma/uritemplates"
)

// Discover takes a discovery URI (the caller does not need to pass it through
// discovery.Resolve), fetches it using discovery.Fetch, evaluates the
// distribution URI template, fetches and parses the distribution object and
// returns said distribution object. If provided, variables will be mutated to
// contain the spec-defined variables (generated during the discovery stage).
func Discover(name string, variables map[string]interface{}) (string, parcelv0.Distribution, error) {
	if variables == nil {
		variables = make(map[string]interface{})
	}

	// Parse the original URI.
	userNameUri, err := url.Parse("//" + name)
	if err != nil {
		return "", parcelv0.Distribution{}, err
	}

	// Make the name canonical.
	name, err = Resolve(name)
	if err != nil {
		return "", parcelv0.Distribution{}, err
	}

	// Parse the canonical URI.
	nameUri, err := url.Parse("//" + name)
	if err != nil {
		return "", parcelv0.Distribution{}, err
	}

	// Get and parse the discovery blobs.
	_, discovery, err := Fetch(name)
	if err != nil {
		return "", parcelv0.Distribution{}, err
	}
	distTemplate, err := uritemplates.Parse(discovery.DistributionURI.Template)
	if err != nil {
		return "", parcelv0.Distribution{}, err
	}

	// Modify the variables as per the spec.
	variables["parcel.version"] = parcelv0.Version
	variables["parcel.discovery.authority"] = nameUri.Host
	variables["parcel.discovery.userAuthority"] = userNameUri.Host
	variables["parcel.discovery.name"] = userNameUri.Path

	// Expand the distribution template.
	dist, err := distTemplate.Expand(variables)
	if err != nil {
		return "", parcelv0.Distribution{}, err
	}

	// Parse the distribution URI.
	distUri, err := url.Parse(dist)
	if err != nil {
		return "", parcelv0.Distribution{}, err
	}

	// Create http://<authority>/
	baseUri := *nameUri
	baseUri.Scheme = "http"
	baseUri.RawPath = ""
	baseUri.Path = ""

	// Compute the full URL as per RFC 3986.
	distURL := (&baseUri).ResolveReference(distUri)

	// Fetch the distribution object.
	resp, err := http.Get(distURL.String())
	if err != nil {
		return "", parcelv0.Distribution{}, err
	}
	defer resp.Body.Close()

	var distribution parcelv0.Distribution
	if err := json.NewDecoder(resp.Body).Decode(&distribution); err != nil {
		return "", parcelv0.Distribution{}, err
	}
	if distribution.ParcelVersion != parcelv0.Version {
		// TODO: Make this more sane.
		return "", parcelv0.Distribution{}, fmt.Errorf("discover: unknown version: %s", distribution.ParcelVersion)
	}
	return distURL.String(), distribution, nil
}
