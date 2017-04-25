# OCI Parcel Extension #

This document is an extension to the [Open Container Initiative (OCI) image
specification][image-spec], adding a discovery and read-only distribution
interface. The intention of this extension is to be as protocol-agnostic as
possible with regards to the distribution protocol (with the possibility for
extensions to the discovery protocol to make it also protocol-agnostic).

The current state of [container image distribution][docker-distribution] has
several issues, which this extension attempts to improve:

* The `docker://` protocol and schema are not truly state-less HTTP, and
  therefore cannot be implemented by a "dumb" CDN. By necessity a stateful
  application must be run by a distributor, which is not always reasonable or
  possible. It also makes caching harder to implement for something like
  Varnish.

* The `docker://` protocol is the only "official" way of distributing such
  images, which makes other methods of distribution (saving an image and then
  distributing it via FTP, BitTorrent, etc) out-of-band and not supported.
  While this extension does not require that all such methods be support, it
  elevates their usefulness by making them much more supportable.

* Image "naming" and distribution are linked, tying the orthogonal issues of
  identity and source-of-files. This further complicates the jobs of CDNs,
  requiring them to provide DNS round-robin style distribution rather than
  GNU/Linux distribution "mirroring".

The usage of the terms MUST, MUST NOT, MAY, SHOULD, and SHOULD NOT in this document
is described by [RFC 2119][rfc-2119]. Lower-case usage of these terms SHOULD
NOT be taken to be the same as upper-case usage.

This document makes usage of the [JSON interchange format][ecma-404], and when
specifying the fields of JSON objects a similar style to the OCI
[image][image-spec] and [runtime][runtime-spec] specifications is used.

[image-spec]: https://github.com/opencontainers/image-spec
[docker-distribution]: https://github.com/docker/distribution
[rfc-2119]: https://www.ietf.org/rfc/rfc2119.txt
[ecma-404]: https://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf
[runtime-spec]: https://github.com/opencontainers/runtime-spec

## Implementation ##

The following sections describe in detail the steps required to go from the
[discovery stage][image-discovery] through parsing of [distribution
URLs][distribution-url] to arrive at OCI image [blob and index
fetching][image-fetching].

[image-discovery]: #image-discovery
[distribution-url]: #distribution-url
[image-fetching]: #image-blob-retrieval

### Image Discovery ###

*In order for a user to be able to describe what image they want to download, a
discovery system is required. Note that while this discovery system MUST be
implemented, implementations MAY choose to allow users to bypass the discovery
URLs and directly specify the distribution URI-reference.*

The purpose of this section is to describe how to resolve a "discovery URI"
into a "distribution URI-reference". Implementations of this section are
referred to as "explorers".

A "discovery URI" is defined by the following syntax, where `authority`,
`path-rootless`, and `segment-nz` and are defined by [RFC 3986 &sect;
3][rfc-3986-s3]. In the following section, `<authority>` and `<path>` refer to
the `authority` section and the "rest" of the URI.

    distribution-uri   = authority "/" path-rootless
                       / segment-nz

An explorer MAY define a "default authority", such that if `authority` is not
specified by a `distribution-uri` then it MUST be treated as though
`distribution-uri` has the "default authority" specified instead.  If an
explorer does not define a "default authority", it MUST emit an error if an
`authority` was not specified.

An explorer of this specification MUST implement at the least the following
alias resolution steps, and MUST implement them in-order. An explorer MAY
extend these steps by adding extra intermediate stages in the conversion to a
distribution URI-reference.

If any step successfully modifies the value of `<authority>` then the
implementation MUST restart the alias resolution process with the new value of
`<authority>`.

1. The explorer MUST attempt to resolve `opencontainers-parcel.cyphar.<authority>`
   through the [Domain Name System][rfc-1035]. If the name successfully
   resolves to a `CNAME` or `TXT` record then the explorer MUST treat the value
   of the record as though `<authority>` was the value of the record. If the
   `<authority>` is not a valid `authority` value as defined by [RFC 3986
   &sect; 3][rfc-3986-s3], then the explorer MUST emit an error.

With a fully resolved `<authority>` value, an explorer MUST then attempt to
access and parse the contents of the URL (as a [discovery object][discovery-json])
`https://<authority>/.well-known/com.cyphar.opencontainers-parcel` (using [HTTP
over TLS][rfc-7230-s2.7.2]). If an error occurred while attempting to
**access** the URL, the explorer MAY attempt to resolve additional out-of-spec
URLs (though it MUST inform the user and MUST NOT attempt additional resolution
if **access**ing the URL did not result in error).

<!-- TODO: Should we allow users to specify this with #known-schemes ? -->

If the explorer could not successfully locate a [discovery
object][discovery-json], it MUST assume that the contents of the [discovery
object][discovery-json] is the following:

```json
{
	"parcelVersion": "0.0.0",
    "disturi": {
        "template": "/{parcel.version}/{parcel.discovery.name}"
    }
}
```

The [distribution URI-reference][distribution-url] MUST be computed by
evaluating the "disturi" field as a [URI
template][uri-template] with at least the `parcel` and `parcel.discovery`
namespace of [variables][uri-template-variables] defined. The semantic meaning
of "distribution URI-reference" is defined in the [distribution URI-reference
section][distribution-url].

[rfc-3986-s3]: https://tools.ietf.org/html/rfc3986#section-3
[rfc-1035]: https://tools.ietf.org/html/rfc1035
[rfc-2068]: https://tools.ietf.org/html/rfc2068
[rfc-5246]: https://tools.ietf.org/html/rfc5246
[rfc-7230-s2.7.2]: https://tools.ietf.org/html/rfc7230#section-2.7.2
[discovery-json]: #discovery-json
[uri-template]: #uri-templates
[uri-template-variables]: #uri-template-variables
[version]: #version
[distribution-url]: #distribution-url

#### Discovery JSON ####

*The discovery JSON object is used to specify the template for the distribution
URL. The purpose of the second level of indirection is to allow for separation
between the authority dereferencing of images and their storage (also allowing
for multiplexing and layered versioning of images using immutable storage).*

The following fields are defined by this specification, and MUST at least be
implemented. Additional fields MAY be supported by implementations, however if
an additional field is not supported by an implementation it MUST be ignored by
the implementation.

The URL defined in `disturi` MUST be computed with at least the `parcel` and
`parcel.discovery` namespaces of [variables][uri-template-variables] defined.

* **`parcelVersion`** (string, REQUIRED)

  The [version of this specification][version] implemented by the server. An
  implementation SHOULD emit a warning if it encounters an unsupported version,
  and MAY emit an error.

* **`disturi`** (object ([URI template][uri-template]), REQUIRED)

  The template for the [distribution URI-reference][distribution-url], used in
  the retrieval of the blob and index objects stored in an OCI image. If, when
  expanded, the distribution URI-reference is not a valid `URI-Reference` (as
  defined by [RFC 3986 &sect; 4.1][rfc-3986-s4.1]), the explorer MUST emit an
  error.

[uri-template-variables]: #uri-template-variables
[version]: #version
[distribution-url]: #distribution-url
[uri-template]: #uri-templates
[rfc-3986-s4.1]: https://tools.ietf.org/html/rfc3986#section-4.1

### Distribution URL ###

The purpose of this section is to describe how to resolve a "distribution
URI-reference" to [index and blob retrieval URL templates][image-blob-retrieval]
which can be used to download the blobs of an image. The process of discovering
a distribution URI-reference for a given "user friendly" discovery URI is
described in the [image discovery section][image-discovery]. An implementation
of this section is referred to as a "consumer".

The syntax of a "distribution URI-reference" (and part of the semantic meaning)
is described by `URI-reference` in [RFC 3986 &sect; 4.1][rfc-3986-s4.1]. If a
consumer encounters an invalid "distribution URI-reference", it MUST emit an
error.

The consumer MUST resolve the distribution URI-reference as a "URI reference"
as described in [RFC 3986 &sect; 5][rfc-3986-s5] to provide a fully qualified
"distribution URL".  If necessary to resolve the distribution URI-reference,
the "base URI" used MUST be `http://<authority>/`, with `<authority>` defined
through the same resolution process as in the [discovery stage][image-discovery].

After resolving the URI-reference, the distribution URL MUST be resolved as
described in [known schemas][known-schemas] and its contents parsed as a
[distribution object][distribution-json]. If an error occurs during resolution
or parsing, the consumer MUST emit an error.

[image-blob-retrieval]: #image-blob-retrieval
[image-discovery]: #image-discovery
[rfc-3986-s4.1]: https://tools.ietf.org/html/rfc3986#section-4.1
[rfc-3986-s5]: https://tools.ietf.org/html/rfc3986#section-5
[known-schemas]: #known-schemas
[distribution-json]: #distribution-object

#### Distribution Object ####

*The discovery JSON object is used to specify the templates for the index and
blob URLs, allowing image servers to explain with significant freedom how to
attain the various blobs in an OCI image.*

The following fields are defined by this specification, and MUST at least be
implemented. Additional fields MAY be supported by implementations, however if
an additional field is not supported by an implementation it MUST be ignored by
the implementation.

The URLs defined in `indexuris` and `bloburis` MUST be computed with at least
the `parcel`, `parcel.fetch` (with the exception of `parcel.fetch.blob`), and
`parcel.discovery` namespaces of [variables][uri-template-variables] defined.

* **`parcelVersion`** (string, REQUIRED)

  The [version of this specification][version] implemented by the server. An
  implementation SHOULD emit a warning if it encounters an unsupported version,
  and MAY emit an error.

* **`indexuris`** (array of object ([URI template][uri-template]), REQUIRED)

  The list of URI-reference templates which MUST be used by fetchers to get the
  contents of the OCI image's [image index][oci-image-index]. Fetchers SHOULD
  use the contents of the image index in order to ascertain (by walking OCI's
  `Descriptor` paths) the minimal set of blobs that must be downloaded to
  fulfil the user's request.

  A fetcher SHOULD NOT re-fetch the image index more than once in a single
  image download. However, the fetcher MAY fetch the image index using
  different template entries in the `indexuris` array in subsequent fetches.

* **`bloburis`** (array of object ([URI template][uri-template]), REQUIRED)

  The list of URI-reference templates which MUST be used by fetchers to get the
  contents of any content-addressable blob (in the context of the distribution
  URL). The `parcel.fetch.blob` namespace of
  [variables][uri-template-variables] MUST be defined when computing templates
  in this array.

  A fetcher SHOULD NOT re-fetch a given blob more than once in a single image
  download. Fetchers MAY fetch blobs using different template entries in the
  `bloburis` array.

> **NOTE**: Servers MAY wish to use the array of `bloburis` to allow for
>           specifying "official" mirrors of blobs (where the mirrors may wish
>           to use further load-balancers). As such, implementations MAY wish
>           to take this intention into account and round-robin (as well as
>           parallelise) the fetching of blobs.

[distribution-url]: #distribution-url
[uri-template]: #uri-templates
[uri-template-variables]: #uri-template-variables
[rfc-3986-s4.1]: https://tools.ietf.org/html/rfc3986#section-4.1
[oci-image-index]: https://github.com/opencontainers/image-spec/blob/v1.0.0-rc5/image-index.md

### Image Blob Retrieval ###

*The final stage of fetching an OCI image is actually resolving, fetching and
parsing the [OCI image index][oci-image-index] and various associated blobs.*

The purpose of this section is to describe how to use the [`indexuris` and
`bloburis` arrays][distribution-json] sourced during the [distribution
stage][distribution-url] of parcel fetching. An implementation of this section
is referred to as a "fetcher".

After evaluation of the URI-reference templates as described by in the
[distribution object section][distribution-json], the syntax of the index and
blob URI-references (and part of their semantic meaning) is described by
`URI-reference` in [RFC 3986 &sect; 4.1][rfc-3986-s4.1].

The fetcher MUST resolve both index and blob URI-references as "URI reference",
as described in [RFC 3986 &sect; 5][rfc-3986-s5] to produce fully qualified
index and blob URLs.  If necessary to resolve the distribution URI-reference,
the "base URI" used MUST be the [distribution URL][distribution-url] used to
download the [distribution object][distribution-json].

If an index or blob URI-reference is invalid, the fetcher MUST act as though
the invalid URI-reference was not present in the original set of URI-references
(though it SHOULD emit a warning). If there are no valid URI-references in the
set of blob or index URI-references, the fetcher MUST emit an error.

After resolving the URI-references, the resultant URLs MUST be resolved as
described in [known schemas][known-schemas] and their contents SHOULD be parsed
as appropriate for their [OCI image `mediaType`s][oci-image-mediatype]. If an
unknown `mediaType` is encountered, fetchers SHOULD download the blob without
parsing it (though they SHOULD emit a warning that the downloaded image may be
incomplete).

> **NOTE**: While the algorithm for deciding what blobs are necessary to
> download is **not** specified by this document, fetchers SHOULD attempt a
> recursive [OCI `Descriptor`][oci-descriptor] walk to decide the set of blobs
> that are necessary to fulfil a top-level `index` entry requirement (which is
> retrieved from the `indexuris` and parsed as an [OCI image
> index][oci-image-index]).

[distribution-json]: #distribution-object
[distribution-url]: #distribution-url
[rfc-3986-s4.1]: https://tools.ietf.org/html/rfc3986#section-4.1
[rfc-3986-s5]: https://tools.ietf.org/html/rfc3986#section-5
[known-schemas]: #known-schemas
[oci-image-mediatype]: https://github.com/opencontainers/image-spec/blob/v1.0.0-rc5/media-types.md
[oci-descriptor]: https://github.com/opencontainers/image-spec/blob/v1.0.0-rc5/descriptor.md
[oci-image-index]: https://github.com/opencontainers/image-spec/blob/v1.0.0-rc5/image-index.md

<!-- ### Example ### -->

## URI Templates ##

*A URI template is used to allow for more generic descriptions of URLs without
requiring that the server enumerate all possible URLs.*

A URI template is defined as a JSON object, containing the template string.
Note that clients which retrieve URI template files MUST NOT assume that the
contents of the URI template are constant between distinct retrievals.

The following fields are defined by this specification, and MUST at least be
implemented. Additional fields MAY be supported by implementations, however if
an additional field is not supported by an implementation it MUST be ignored by
the implementation.

* **`template`** (string, REQUIRED) specifies the template string to be
  expanded, as defined by [RFC 6570][rfc-6570].

<!-- TODO: Do we need a "default" string->string map? -->

Implementations MUST implement Level 4 template expansions (as defined by [RFC
6570 &sect 1.2][rfc-6570-s1.2]). However, implementors SHOULD note that this
specification does not require support for composite data type substitution
(arrays or key-value mappings) and thus only string substitution MUST be
implemented.

[rfc-6570]: https://tools.ietf.org/html/rfc6570
[rfc-6570-s1.2]: https://tools.ietf.org/html/rfc6570#section-1.2

<!-- ## Glossary ## -->

## Prior Art ##

This extension was heavily influenced by the [AppC image discovery
specification][aci-discovery], as well as personal concerns of the author with
regards to the current (centralised and protocol-centric) state of container
image distribution.

In addition, one of the very large concerns of this extension was to ensure
that pre-existing distribution systems (such as the [Open Build Service][obs])
will be able to seamlessly publish these sorts of images (and that their
pre-existing CDN integrations would also operate smoothly), without requiring
RPM wrapping around all of the blobs.

The eventual intention is that projects like [openSUSE's
`containment-rpm`][containment-rpm] will be able to produce `parcel`
repositories (with the [discovery][image-discovery] being published separately
on the main [`opensuse.org`][opensuse] website), taking advantage of the
existing [OBS infrastructure][opensuse-obs] and CDN setup.

[aci-discovery]: https://github.com/appc/spec/blob/v0.8.10/spec/discovery.md
[obs]: http://openbuildservice.org
[image-discovery]: #image-discovery
[containment-rpm]: https://github.com/SUSE/containment-rpm-docker
[opensuse]: https://opensuse.org
[opensuse-obs]: https://build.opensuse.org

## Version ##

This document is versioned in accordance with [Semantic Versioning
v2.0.0][semver]. The current version of this document is **`0.0.0`**, and is
currently considered a **DRAFT**.

[semver]: http://semver.org/spec/v2.0.0.html

## Copyright ##

This document is licensed under the Apache 2.0 license.

```
Copyright (C) 2017 SUSE LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

## Appendix ##

This section and sub-sections define supplementary definitions of syntax and
other semantics. They SHOULD NOT be used outside of the context of this
specifiation.

### URI Template Variables ###

The following variables are defined by this specification. If a variable is
listed as REQUIRED then an implementation MUST allow substitution of that
variable, and OPTIONAL variables mean that an implementation SHOULD allow
substitution of that variable.

These variables are namespaced -- implementations MAY extend the following list
of variables (and SHOULD also namespace their variables), but all variables in
the `parcel` namespace (those prefixed with `parcel.`) MUST NOT be used except
as specified in this document (they are reserved for future extensions).

* **`parcel.version`** (string, REQUIRED)

  The version of this specification implemented by the explorer. It MUST be the
  same value as specified [in the version section of this document][version].

The following variables are defined from the [discovery stage][image-discovery]
onwards. If an implementation did not consume a [discovery
object][discovery-json], it MUST use the values defined as the default
[discovery object][discovery-json] in the [discovery stage][image-discovery].

* **`parcel.discovery.authority`** (string, REQUIRED)

  The final value of `<authority>` computed by the explorer, which MAY be
  different from the `<authority>` specified by the user.

* **`parcel.discovery.userAuthority`** (string, REQUIRED)

  The value of `<authority>` as provided by the user, which MAY be different
  from `parcel.authority`.

* **`parcel.discovery.name`** (string, REQUIRED)

  The value of `<name>` as specified by the user.

* **`parcel.discovery.nameDigest`** (string, REQUIRED)

  The lowercase hexadecimal representation of the digest of the `<name>`
  specified by the user, using the digest specified by the [discovery
  object][discovery-json]. If the explorer does not support the digest
  specified, it MUST emit an error.

* **`parcel.discovery.digestAlgorithm`** (string, REQUIRED)

  The name of the digest algorithm specified by the [discovery
  object][discovery-json].

The following variables are defined from the [blob retrieval
stage][image-fetching] stage onwards. An implementation MUST define these
variables when fetching a `bloburi`.

* **`parcel.fetch.blob.algorithm`** (string, REQUIRED)

  The name of the digest algorithm used for producing the digest of an OCI
  image blob, as specified by the [OCI descriptor][oci-descriptor] that
  resulted in the blob being fetched. The syntax and semantic meaning of this
  value is described in the [OCI image specification][oci-digests] as
  `algorithm`.

* **`parcel.fetch.blob.digest`** (string, REQUIRED)

  The lowercase hexadecimal representation of the blob digest. The syntax and
  semantic meaning of this value is described in the [OCI image
  specification][oci-digests] as `hex`.

[version]: #version
[image-discovery]: #image-discovery
[discovery-json]: #discovery-json
[image-fetching]: #image-blob-retrieval
[oci-descriptor]: https://github.com/opencontainers/image-spec/blob/v1.0.0-rc5/descriptor.md
[oci-digests]: https://github.com/opencontainers/image-spec/blob/v1.0.0-rc5/descriptor.md#digests-and-verification

### Known Schemes ###

While implementations of this specification MAY implement additional scheme
support, any implementation MUST obey this section (if instructed to by the
main document).

If the scheme or protocol in the given URI is not supported, the implementation
MUST emit an error. Consumers MUST implement at least the following schemes and
protocols:

* `http` refers to the [HyperText Transfer Protocol][rfc-2616] (any version),
  as defined by [RFC 7230 &sect; 2.7.1][rfc-7230-s2.7.1].

* `https` refers to the [HyperText Transfer Protocol][rfc-2616] (any version)
  using [Transport Level Security][rfc-5246] (any version), as defined by [RFC
  7230 &sect; 2.7.2][rfc-7230-s2.7.2].

In addition, the following protocols SHOULD be implemented by implementations.
If an implementation implements the following schemes, the semantics MUST match
those described below.

* `ftp` refers to the [File Transfer Protocol][rfc-959]. Consumers SHOULD
  attempt to log in with `anonymous` credentials before prompting a user for
  credentials.

* `magnet` refers to a [BitTorrent Magnet URI][bep-0009], with an optional [RFC
  3986 &sect; 3.5][rfc-3986-s3.5] `fragment`. If the URI indicates a
  [BitTorrent info-hash][bep-0003], `fragment` indicates the filename within
  the info dictionary that the implementation MUST use as the contents of the
  `magnet` URI. If the URI does not indicate an info-hash, `fragment` MUST be
  ignored by implementations.

* `ipfs` and `ipns` are reserved for use by [Interplanetary File System
  URIs][ipfs]. Consumers MUST emit an error if they encounter these schemes
  (this may change in the future).

[rfc-2616]: https://tools.ietf.org/html/rfc2616
[rfc-7230-s2.7.1]: https://tools.ietf.org/html/rfc7230#section-2.7.1
[rfc-5246]: https://tools.ietf.org/html/rfc5246
[rfc-7230-s2.7.2]: https://tools.ietf.org/html/rfc7230#section-2.7.2
[rfc-959]: https://tools.ietf.org/html/rfc959
[bep-0009]: http://bittorrent.org/beps/bep_0009.html#magnet-uri-format
[rfc-3986-s3.5]: https://tools.ietf.org/html/rfc3986#section-3.5
[bep-0003]: http://bittorrent.org/beps/bep_0003.html#metainfo-files
[ipfs]: https://ipfs.io
