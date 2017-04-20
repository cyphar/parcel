## parcel ##

Some ideas I had about how distribution of OCI images. My hope is that this
system of distribution could be implemented without a need to lock users into a
specific API or provider. Current container image distribution systems have a
significant flaw in that the name of each image is intricately tied to the
source of the image (similar to how Go imports define where the source code
lives). As a result, distributing such images out-of-band requires working
around this system. My intent is that this work could lead to a distribution
system that is as unopinionated as possible.
