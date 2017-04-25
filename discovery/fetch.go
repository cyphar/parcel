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
)

// DiscoveryPath is the path which contains the discovery object JSON.
const DiscoveryPath = "/.well-known/cyphar.opencontainers.parcel.v0.json"

// toDiscoveryURL takes a discovery URI and converts it to a discovery URL.
func toDiscoveryURL(name string) (string, error) {
	URL, err := url.Parse("https://" + name)
	if err != nil {
		return "", err
	}
	URL.Path = DiscoveryPath
	URL.RawPath = DiscoveryPath
	return URL.String(), nil
}

// Fetch takes an already-resolved "canonical" discovery URI and fetches and
// parses the discovery object JSON, returning an error if fetching or parsing
// failed.
//
// TODO: Handle the "default object".
func Fetch(name string) (parcelv0.Discovery, error) {
	discoveryURL, err := toDiscoveryURL(name)
	if err != nil {
		return parcelv0.Discovery{}, err
	}

	resp, err := http.Get(discoveryURL)
	if err != nil {
		return parcelv0.Discovery{}, err
	}
	defer resp.Body.Close()

	var discovery parcelv0.Discovery
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return parcelv0.Discovery{}, err
	}
	if discovery.ParcelVersion != parcelv0.Version {
		// TODO: Make this more sane.
		return parcelv0.Discovery{}, fmt.Errorf("discovery fetch: unknown version: %s", discovery.ParcelVersion)
	}
	return discovery, nil
}
