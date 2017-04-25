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

package v0

import "github.com/cyphar/parcel/specs-go"

// Distribution describes a "Distribution Object", used in the consumer and
// fetching stages of the parcel specification.
type Distribution struct {
	specs.Versioned

	// IndexURIs represents the set of URI templates that can be expanded and
	// used to download the image's index JSON.
	IndexURIs []Template `json:"indexuris"`

	// BlobURIs represents the set of URI templates that can be expanded and
	// used to download the image's blobs.
	BlobURIs []Template `json:"bloburis"`
}
