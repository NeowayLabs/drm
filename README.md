# drm

The _Direct Rendering Manager_ (DRM) is a kernel framework to manage _Graphical Processing Units_ (GPU).
It's a kernel abstractions to GPU drivers and a userspace API 
designed to support the needs of complex graphics devices.

DRM was first implemented on Linux but ported to FreeBSD, NetBSD and Solaris (others?). It's the lower level interface between opengl and the graphics card. With this Go library, theoretically, now it's possible to create a X server or a pure OpenGL library in Go (no bindings).

## Rationale

Enables the creation of a graphics stack in Go, avoiding the overhead of the existing C bindings.
Another possibility is using Go to make GPGPU (like opencl).

## Examples

See the [examples](https://github.com/tiago4orion/drm/tree/master/_examples) directory.

