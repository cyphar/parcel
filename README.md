## parcel ##

Some ideas I had about how distribution of OCI images. My hope is that this
system of distribution could be implemented without a need to lock users into a
specific API or provider. Current container image distribution systems have a
significant flaw in that the name of each image is intricately tied to the
source of the image (similar to how Go imports define where the source code
lives). As a result, distributing such images out-of-band requires working
around this system. My intent is that this work could lead to a distribution
system that is as unopinionated as possible.

A [description of the design](DESIGN.md) has been written up, in the hopes that
it may be combined with other ideas and included in the OCI image
specification.

### License ###

`parcel` is licensed under the terms of the Apache 2.0 license.

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
